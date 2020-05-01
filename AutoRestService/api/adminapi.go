package api

import (
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
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/", BackendsPrefix), GetAdminBackendsHandler)
	router.With(RoleCheck([]string{"admin"})).Post(fmt.Sprintf("/%s/", BackendsPrefix), PostAdminBackendHandler)
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/{bename}/", BackendsPrefix), GetAdminBackendHandler)
	router.With(RoleCheck([]string{"admin"})).Delete(fmt.Sprintf("/%s/{bename}/dropdata", BackendsPrefix), DeleteAdminBackendEndpoint)
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/{bename}/models", BackendsPrefix), GetAdminModelsHandler)
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/{bename}/datasources", BackendsPrefix), GetAdminDatasourcesHandler)
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/{bename}/rules", BackendsPrefix), GetAdminRulesHandler)
	router.With(RoleCheck([]string{"admin"})).Get(fmt.Sprintf("/%s/{bename}/rules/{rulename}", BackendsPrefix), GetAdminRulesRuleHandler)
	router.With(RoleCheck([]string{"admin"})).Post(fmt.Sprintf("/%s/{bename}/rules/{rulename}/test", BackendsPrefix), PostAdminRulesRuleTestHandler)
	return router
}

type BackendInfo struct {
	Name        string
	Description string
	URL         string
}

// GetAdminBackendsHandler create a new backend
func GetAdminBackendsHandler(response http.ResponseWriter, request *http.Request) {
	log.Infof("GET: path: %s", request.URL.Path)
	names := model.BackendList.Names()
	backendInfos := make([]BackendInfo, 0)
	myconfig := config.Get()
	for _, name := range names {
		backend, _ := model.BackendList.Get(name)
		backendInfos = append(backendInfos, BackendInfo{
			Name:        name,
			Description: backend.Description,
			URL:         fmt.Sprintf("%s%s%s/", myconfig.ServiceURL, request.URL.Path, name),
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

// PostAdminBackendHandler create a new backend
func PostAdminBackendHandler(response http.ResponseWriter, request *http.Request) {
	log.Infof("POST: path: %s", request.URL.Path)
	
	render.Render(response, request, ErrNotImplemted)
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
