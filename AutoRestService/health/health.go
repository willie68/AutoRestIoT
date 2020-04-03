package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/willie68/AutoRestIoT/logging"
)

var myhealthy bool
var log logging.ServiceLogger

/*
This is the healtchcheck you will have to provide.
*/
func check() (bool, string) {
	// TODO implement here your healthcheck.
	myhealthy = !myhealthy
	message := ""
	if myhealthy {
		log.Info("healthy")
	} else {
		log.Info("not healthy")
		message = "ungesund"
	}
	return myhealthy, message
}

//##### template internal functions for processing the healthchecks #####
var healthmessage string
var healthy bool
var lastChecked time.Time
var period int

// CheckConfig configuration for the healthcheck system
type CheckConfig struct {
	Period int
}

// InitHealthSystem initialise the complete health system
func InitHealthSystem(config CheckConfig) {
	period = config.Period
	log.Infof("healthcheck starting with period: %d seconds", period)
	healthmessage = "service starting"
	healthy = false
	doCheck()
	go func() {
		background := time.NewTicker(time.Second * time.Duration(period))
		for _ = range background.C {
			doCheck()
		}
	}()
}

/*
internal function to process the health check
*/
func doCheck() {
	var msg string
	healthy, msg = check()
	if !healthy {
		healthmessage = msg
	} else {
		healthmessage = ""
	}
	lastChecked = time.Now()
}

/*
Routes getting all routes for the health endpoint
*/
func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/health", GetHealthyEndpoint)
	router.Get("/readiness", GetReadinessEndpoint)
	return router
}

/*
GetHealthyEndpoint is this service healthy
*/
func GetHealthyEndpoint(response http.ResponseWriter, req *http.Request) {
	t := time.Now()
	if t.Sub(lastChecked) > (time.Second * time.Duration(2*period)) {
		healthy = false
		healthmessage = "Healthcheck not running"
	}
	response.Header().Add("Content-Type", "application/json")
	if healthy {
		response.WriteHeader(http.StatusOK)
		message := fmt.Sprintf(`{ "message": "service up and running", "lastCheck": "%s" }`, lastChecked.String())
		response.Write([]byte(message))
	} else {
		response.WriteHeader(http.StatusServiceUnavailable)
		message := fmt.Sprintf(`{ "message": "service is unavailable: %s", "lastCheck": "%s" }`, healthmessage, lastChecked.String())
		response.Write([]byte(message))
	}
}

/*
GetReadinessEndpoint is this service ready for taking requests
*/
func GetReadinessEndpoint(response http.ResponseWriter, req *http.Request) {
	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(`{ "message": "service started" }`))
}
