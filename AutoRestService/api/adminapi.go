package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/config"
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/logging"
	"github.com/willie68/AutoRestIoT/model"
	"github.com/willie68/AutoRestIoT/worker"
)

var log logging.ServiceLogger

const AdminPrefix = "admin"
const BackendsPrefix = "backends"

//AdminRoutes getting all routes for the config endpoint
func AdminRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Mount("/tasks", TasksRoutes())
	router.With(RoleCheck([]string{"admin"})).Get("/info", GetAdminInfoHandler)
	router.With(RoleCheck([]string{"admin"})).Post("/test/rule", PostAdminTestRuleHandler)
	router.With(RoleCheck([]string{"admin", "edit", "read"})).Get(fmt.Sprintf("/%s/", BackendsPrefix), GetAdminBackendsHandler)
	router.With(RoleCheck([]string{"admin"})).Post(fmt.Sprintf("/%s/", BackendsPrefix), PostAdminBackendHandler)
	router.With(RoleCheck([]string{"admin", "edit", "read"})).Get(fmt.Sprintf("/%s/{bename}", BackendsPrefix), GetAdminBackendHandler)
	router.With(RoleCheck([]string{"admin"})).Delete(fmt.Sprintf("/%s/{bename}", BackendsPrefix), DeleteAdminBackendHandler)
	router.With(RoleCheck([]string{"admin"})).Delete(fmt.Sprintf("/%s/{bename}/dropdata", BackendsPrefix), DeleteAdminBackendEndpoint)
	router.With(RoleCheck([]string{"admin", "edit", "read"})).Get(fmt.Sprintf("/%s/{bename}/models", BackendsPrefix), GetAdminModelsHandler)
	router.With(RoleCheck([]string{"admin", "edit", "read"})).Get(fmt.Sprintf("/%s/{bename}/models/{model}", BackendsPrefix), GetAdminModelHandler)
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/{bename}/rules", BackendsPrefix), GetAdminRulesHandler)
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/{bename}/rules/{rulename}", BackendsPrefix), GetAdminRulesRuleHandler)
	router.With(RoleCheck([]string{"admin"})).Post(fmt.Sprintf("/%s/{bename}/rules/{rulename}/test", BackendsPrefix), PostAdminRulesRuleTestHandler)
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/{bename}/datasources", BackendsPrefix), GetAdminDatasourcesHandler)
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/{bename}/datasources/{datasource}", BackendsPrefix), GetAdminDatasourceHandler)
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/{bename}/destinations", BackendsPrefix), GetAdminDestinationsHandler)
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/{bename}/destinations/{destination}", BackendsPrefix), GetAdminDestinationHandler)
	return router
}

type BackendInfo struct {
	Name         string
	Description  string
	URL          string
	Models       []string
	Rules        []string
	Datasources  []string
	Destinations []string
}

// GetAdminInfoHandler getting server info
func GetAdminInfoHandler(response http.ResponseWriter, request *http.Request) {
	log.Infof("GET: path: %s", request.URL.Path)
	info := model.JSONMap{}

	info["backends"] = model.BackendList.Names()
	info["modelcounter"] = worker.GetModelCount()
	info["rules"] = worker.Rules.GetRulelist()
	info["mqttClients"] = worker.GetMQTTClients()

	render.JSON(response, request, info)
}

// GetAdminBackendsHandler create a new backend
func GetAdminBackendsHandler(response http.ResponseWriter, request *http.Request) {
	log.Infof("GET: path: %s", request.URL.Path)
	names := model.BackendList.Names()
	backendInfos := make([]BackendInfo, 0)
	myconfig := config.Get()
	for _, name := range names {
		backend, _ := model.BackendList.Get(name)
		modelNames := make([]string, 0)
		for _, model := range backend.Models {
			modelNames = append(modelNames, model.Name)
		}
		ruleNames := make([]string, 0)
		for _, rule := range backend.Rules {
			ruleNames = append(ruleNames, rule.Name)
		}
		datasourceNames := make([]string, 0)
		for _, datasource := range backend.DataSources {
			datasourceNames = append(datasourceNames, datasource.Name)
		}
		destinationNames := make([]string, 0)
		for _, destination := range backend.Destinations {
			destinationNames = append(destinationNames, destination.Name)
		}
		backendInfos = append(backendInfos, BackendInfo{
			Name:         name,
			Description:  backend.Description,
			URL:          fmt.Sprintf("%s%s%s/", myconfig.ServiceURL, request.URL.Path, name),
			Models:       modelNames,
			Rules:        ruleNames,
			Datasources:  datasourceNames,
			Destinations: destinationNames,
		})
	}
	render.JSON(response, request, backendInfos)
}

// GetAdminBackendHandler create a new backend
func GetAdminBackendHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	log.Infof("GET: path: %s, be: %s", request.URL.Path, backendName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	render.JSON(response, request, backend)
}

// DeleteAdminBackendHandler create a new backend
func DeleteAdminBackendHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	log.Infof("DELETE: path: %s, be: %s", request.URL.Path, backendName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	err := worker.DeregisterBackend(backendName)
	if err != nil {
		log.Alertf("%v\n", err)
		render.Render(response, request, ErrInternalServer(err))
		return
	}

	worker.DeleteBackend(backendName)

	m := make(map[string]interface{})
	m["backend"] = backend
	m["msg"] = fmt.Sprintf("backend %s definition deleted. No data destroyed.", backendName)

	render.JSON(response, request, m)
}

// PostAdminBackendHandler create a new backend
func PostAdminBackendHandler(response http.ResponseWriter, request *http.Request) {
	log.Infof("POST: path: %s", request.URL.Path)

	data := &model.Backend{}
	if err := render.Decode(request, data); err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}

	bemodel := *data
	if !model.BackendList.Contains(bemodel.Backendname) {
		render.Status(request, http.StatusCreated)
	} else {
		worker.DeregisterBackend(bemodel.Backendname)
	}
	bemodel, err := worker.PrepareBackend(bemodel)
	if err != nil {
		log.Alertf("%v\n", err)
		render.Render(response, request, ErrInternalServer(err))
		return
	}
	_, err = worker.StoreBackend(bemodel)
	if err != nil {
		log.Alertf("%v\n", err)
		render.Render(response, request, ErrInternalServer(err))
		return
	}
	err = worker.RegisterBackend(bemodel)
	if err != nil {
		log.Alertf("%v\n", err)
		render.Render(response, request, ErrInternalServer(err))
		return
	}
	render.JSON(response, request, bemodel)
}

//DeleteAdminBackendEndpoint delete a backend with data
func DeleteAdminBackendEndpoint(response http.ResponseWriter, request *http.Request) {
	backend := chi.URLParam(request, "bename")
	log.Infof("DELETE: path: %s, be: %s", request.URL.Path, backend)

	err := dao.GetStorage().DeleteBackend(backend)
	if err != nil {
		log.Alertf("%v\n", err)
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}

	m := make(map[string]interface{})
	m["backend"] = backend
	m["msg"] = fmt.Sprintf("backend %s deleted. All data destroyed.", backend)

	render.JSON(response, request, m)
}

// GetAdminModelsHandler getting all models of a backend
func GetAdminModelsHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	log.Infof("GET: path: %s, be: %s", request.URL.Path, backendName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	models := backend.Models
	render.JSON(response, request, models)
}

// GetAdminModelHandler getting all models of a backend
func GetAdminModelHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	modelName := chi.URLParam(request, "model")
	log.Infof("GET: path: %s, be: %s, md: %s", request.URL.Path, backendName, modelName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	myModel, ok := backend.GetModel(modelName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	render.JSON(response, request, myModel)
}

// GetAdminRulesHandler getting all rules of a backend
func GetAdminRulesHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	log.Infof("GET: path: %s, be: %s", request.URL.Path, backendName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	rules := backend.Rules
	render.JSON(response, request, rules)
}

//GetAdminRulesRuleHandler getting all rules of a backend
func GetAdminRulesRuleHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	ruleName := chi.URLParam(request, "rulename")
	log.Infof("GET: path: %s, be: %s", request.URL.Path, backendName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	rule, ok := backend.GetRule(ruleName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	render.JSON(response, request, rule)
}

//PostAdminRulesRuleTestHandler getting all rules of a backend
func PostAdminRulesRuleTestHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	ruleName := chi.URLParam(request, "rulename")
	log.Infof("POST: path: %s, be: %s", request.URL.Path, backendName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	_, ok = backend.GetRule(ruleName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	transformedJson, err := worker.Rules.TransformJSON(backendName, ruleName, body)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	response.Header().Add("Content-Type", "application/json")
	response.Write(transformedJson)
}

// GetAdminDatasourcesHandler getting all datasources of a backend
func GetAdminDatasourcesHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	log.Infof("GET: path: %s, be: %s", request.URL.Path, backendName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	datasources := backend.DataSources
	render.JSON(response, request, datasources)
}

// GetAdminDatasourceHandler getting all datasources of a backend
func GetAdminDatasourceHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	datasourceName := chi.URLParam(request, "datasource")
	log.Infof("GET: path: %s, be: %s, src: %s", request.URL.Path, backendName, datasourceName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	datasource, ok := backend.GetDatasource(datasourceName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	render.JSON(response, request, datasource)
}

// GetAdminDestinationsHandler getting all datasources of a backend
func GetAdminDestinationsHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	log.Infof("GET: path: %s, be: %s", request.URL.Path, backendName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	destinations := backend.Destinations
	render.JSON(response, request, destinations)
}

// GetAdminDestinationHandler getting all datasources of a backend
func GetAdminDestinationHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	destinationName := chi.URLParam(request, "destination")
	log.Infof("GET: path: %s, be: %s, src: %s", request.URL.Path, backendName, destinationName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	destination, ok := backend.GetDestination(destinationName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	render.JSON(response, request, destination)
}

// PostAdminTestRuleHandler test a new rule with data
func PostAdminTestRuleHandler(response http.ResponseWriter, request *http.Request) {
	log.Infof("POST: path: %s", request.URL.Path)

	data := &model.TestRuleDTO{}
	if err := render.Decode(request, data); err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	jsonStr, err := json.Marshal(data.Rule.Transform)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	k, err := worker.NewRule(string(jsonStr))
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	jsonStr, err = json.Marshal(data.Data)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	out, err := k.TransformInPlace(jsonStr)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	response.Header().Add("Content-Type", "application/json")
	response.Write(out)
	//	render.JSON(response, request, data)
}
