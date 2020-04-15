package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	api "github.com/willie68/AutoRestIoT/api"
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/health"
	"github.com/willie68/AutoRestIoT/model"
	"github.com/willie68/AutoRestIoT/worker"
	"gopkg.in/yaml.v3"

	"github.com/willie68/AutoRestIoT/internal/crypt"

	consulApi "github.com/hashicorp/consul/api"
	config "github.com/willie68/AutoRestIoT/config"
	"github.com/willie68/AutoRestIoT/logging"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

	flag "github.com/spf13/pflag"
)

/*
apiVersion implementing api version for this service
*/
const apiVersion = "1"
const servicename = "autorest-srv"

var port int
var sslport int
var system string
var serviceURL string
var registryURL string
var apikey string
var ssl bool
var configFile string
var serviceConfig config.Config
var consulAgent *consulApi.Agent
var log logging.ServiceLogger

func init() {
	// variables for parameter override
	ssl = false
	log.Info("init service")
	flag.IntVarP(&port, "port", "p", 0, "port of the http server.")
	flag.IntVarP(&sslport, "sslport", "t", 0, "port of the https server.")
	flag.StringVarP(&system, "systemid", "s", "", "this is the systemid of this service. Used for the apikey generation")
	flag.StringVarP(&configFile, "config", "c", config.File, "this is the path and filename to the config file")
	flag.StringVarP(&serviceURL, "serviceURL", "u", "", "service url from outside")
	flag.StringVarP(&registryURL, "registryURL", "r", "", "registry url where to connect to consul")
}

func routes() *chi.Mux {
	myHandler := api.NewSysAPIHandler(serviceConfig.SystemID, apikey)
	baseURL := fmt.Sprintf("/api/v%s", apiVersion)
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Logger,
		middleware.DefaultCompress,
		middleware.Recoverer,
		myHandler.Handler,
	)

	router.Route("/", func(r chi.Router) {
		r.With(api.BasicAuth(servicename)).Mount(baseURL+"/models", api.ModelRoutes())
		r.With(api.BasicAuth(servicename)).Mount(baseURL+"/files", api.FilesRoutes())
		r.With(api.BasicAuth(servicename)).Mount(baseURL+"/users", api.UsersRoutes())
		r.With(api.BasicAuth(servicename)).Mount(baseURL+"/admin", api.AdminRoutes())
		r.Mount("/health", health.Routes())
	})
	return router
}

func healthRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Logger,
		middleware.DefaultCompress,
		middleware.Recoverer,
	)

	router.Route("/", func(r chi.Router) {
		r.Mount("/health", health.Routes())
	})
	return router
}

func main() {
	log.Info("starting server")
	flag.Parse()

	config.File = configFile
	if err := config.Load(); err != nil {
		log.Alertf("can't load config file: %s", err.Error())
	}
	serviceConfig = config.Get()
	initConfig()
	initGraylog()

	healthCheckConfig := health.CheckConfig(serviceConfig.HealthCheck)

	defer log.Close()

	if serviceConfig.SystemID == "" {
		log.Fatal("system id not given, can't start! Please use config file or -s parameter")
	}

	gc := crypt.GenerateCertificate{
		Organization: "MCS Media Computer Spftware",
		Host:         "127.0.0.1",
		ValidFor:     10 * 365 * 24 * time.Hour,
		IsCA:         false,
		EcdsaCurve:   "P256",
		Ed25519Key:   true,
	}

	if serviceConfig.Sslport > 0 {
		ssl = true
		log.Info("ssl active")
	}

	storage := &dao.MongoDAO{}
	storage.InitDAO(config.Get().MongoDB)
	dao.SetStorage(storage)
	health.InitHealthSystem(healthCheckConfig)
	apikey = getApikey()
	log.Infof("systemid: %s", serviceConfig.SystemID)
	log.Infof("apikey: %s", apikey)
	log.Infof("ssl: %t", ssl)
	log.Infof("serviceURL: %s", serviceConfig.ServiceURL)
	if serviceConfig.RegistryURL != "" {
		log.Infof("registryURL: %s", serviceConfig.RegistryURL)
	}
	router := routes()
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Infof("%s %s", method, route)
		return nil
	}

	//fmt.Println(docgen.MarkdownRoutesDoc(router, docgen.MarkdownOpts{}))

	if err := chi.Walk(router, walkFunc); err != nil {
		log.Alertf("Logging err: %s", err.Error())
	}
	log.Info("Health routes")
	healthRouter := healthRoutes()
	if err := chi.Walk(healthRouter, walkFunc); err != nil {
		log.Alertf("Logging err: %s", err.Error())
	}

	initAutoRest()

	var sslsrv *http.Server
	var srv *http.Server
	if ssl {
		tlsConfig, err := gc.GenerateTLSConfig()
		if err != nil {
			log.Alertf("logging err: %s", err.Error())
		}
		sslsrv = &http.Server{
			Addr:         "0.0.0.0:" + strconv.Itoa(serviceConfig.Sslport),
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      router,
			TLSConfig:    tlsConfig,
		}
		go func() {
			log.Infof("starting https server on address: %s", sslsrv.Addr)
			if err := sslsrv.ListenAndServeTLS("", ""); err != nil {
				log.Alertf("error starting server: %s", err.Error())
			}
		}()
		srv = &http.Server{
			Addr:         "0.0.0.0:" + strconv.Itoa(serviceConfig.Port),
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      healthRouter,
		}
		go func() {
			log.Infof("starting http server on address: %s", srv.Addr)
			if err := srv.ListenAndServe(); err != nil {
				log.Alertf("error starting server: %s", err.Error())
			}
		}()
	} else {
		// own http server for the healthchecks
		srv = &http.Server{
			Addr:         "0.0.0.0:" + strconv.Itoa(serviceConfig.Port),
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      router,
		}
		go func() {
			log.Infof("starting http server on address: %s", srv.Addr)
			if err := srv.ListenAndServe(); err != nil {
				log.Alertf("error starting server: %s", err.Error())
			}
		}()
	}

	if serviceConfig.RegistryURL != "" {
		initRegistry()
	}

	//go importData("E:/temp/backup/schematic/dev")
	go generateTempData()

	osc := make(chan os.Signal, 1)
	signal.Notify(osc, os.Interrupt)
	<-osc

	log.Info("waiting for clients")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	srv.Shutdown(ctx)
	if ssl {
		sslsrv.Shutdown(ctx)
	}

	log.Info("finished")

	os.Exit(0)
}

func initGraylog() {
	log.GelfURL = serviceConfig.Logging.Gelfurl
	log.GelfPort = serviceConfig.Logging.Gelfport
	log.SystemID = serviceConfig.SystemID

	log.InitGelf()
}

func initRegistry() {
	//register to consul, if configured
	consulConfig := consulApi.DefaultConfig()
	consulURL, err := url.Parse(serviceConfig.RegistryURL)
	consulConfig.Scheme = consulURL.Scheme
	consulConfig.Address = fmt.Sprintf("%s:%s", consulURL.Hostname(), consulURL.Port())
	consulClient, err := consulApi.NewClient(consulConfig)
	if err != nil {
		log.Alertf("can't connect to consul. %v", err)
	}
	consulAgent = consulClient.Agent()

	check := new(consulApi.AgentServiceCheck)
	check.HTTP = fmt.Sprintf("%s/health/health", serviceConfig.ServiceURL)
	check.Timeout = (time.Minute * 1).String()
	check.Interval = (time.Second * 30).String()
	check.TLSSkipVerify = true
	serviceDef := &consulApi.AgentServiceRegistration{
		Name:  servicename,
		Check: check,
	}

	err = consulAgent.ServiceRegister(serviceDef)

	if err != nil {
		log.Alertf("can't register to consul. %s", err)
		time.Sleep(time.Second * 60)
	}

}

func initConfig() {
	if port > 0 {
		serviceConfig.Port = port
	}
	if sslport > 0 {
		serviceConfig.Sslport = sslport
	}
	if system != "" {
		serviceConfig.SystemID = system
	}
	if serviceURL != "" {
		serviceConfig.ServiceURL = serviceURL
	}
}

func getApikey() string {
	value := fmt.Sprintf("%s_%s", servicename, serviceConfig.SystemID)
	apikey := fmt.Sprintf("%x", md5.Sum([]byte(value)))
	return strings.ToLower(apikey)
}

func initAutoRest() {
	backendPath := serviceConfig.BackendPath

	var files []string
	err := filepath.Walk(backendPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if strings.HasSuffix(info.Name(), ".yaml") {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	for _, value := range files {
		data, err := ioutil.ReadFile(value)
		bemodel := model.Backend{}
		bemodel.DataSources = make([]model.DataSource, 0)
		err = yaml.Unmarshal(data, &bemodel)
		if err != nil {
			log.Alertf("%v", err)
			break
		}
		for i, dataSource := range bemodel.DataSources {
			configJson, err := json.Marshal(dataSource.Config)
			switch dataSource.Type {
			case "mqtt":
				var config model.DataSourceConfigMQTT
				if err = json.Unmarshal(configJson, &config); err != nil {
					log.Alertf("%v", err)
				}
				bemodel.DataSources[i].Config = config
			default:
				log.Alertf("unknown message type: %q", dataSource.Type)
				break
			}
		}
		err = worker.ValidateBackend(bemodel)
		if err != nil {
			log.Alertf("validating backend %s in %s, getting error: %v", bemodel.Backendname, value, err)
			break
		}
		jsonData, err := json.Marshal(bemodel)
		fmt.Println(string(jsonData))
		fmt.Println("--------------------------------------------------------------------------------")
		err = worker.RegisterBackend(bemodel)
		if err != nil {
			log.Alertf("error registering backend %s successfully. %v", bemodel.Backendname, err)
			fmt.Println(err)
		} else {
			backendName := model.BackendList.Add(bemodel)
			log.Infof("registering backend %s successfully.", backendName)
		}
	}
}
