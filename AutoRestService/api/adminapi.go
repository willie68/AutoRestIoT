package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/logging"
	"github.com/willie68/AutoRestIoT/model"
)

var log logging.ServiceLogger

//AdminRoutes getting all routes for the config endpoint
func AdminRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.With(RoleCheck([]string{"admin"})).Get("/{bename}/", GetAdminHandler)
	router.With(RoleCheck([]string{"admin"})).Post("/{bename}/", PostAdminHandler)
	router.With(RoleCheck([]string{"admin"})).Delete("/{bename}/dropdata", DeleteAdminEndpoint)
	return router
}

// GetAdminHandler create a new backend
func GetAdminHandler(response http.ResponseWriter, request *http.Request) {
	backendName := chi.URLParam(request, "bename")
	log.Infof("GET: path: %s, be: %s", request.URL.Path, backendName)
	backend, ok := model.BackendList.Get(backendName)
	if !ok {
		render.Render(response, request, ErrNotFound)
		return
	}
	render.JSON(response, request, backend)
}

// PostAdminHandler create a new backend
func PostAdminHandler(response http.ResponseWriter, request *http.Request) {
	backend := chi.URLParam(request, "bename")
	log.Infof("POST: path: %s, be: %s", request.URL.Path, backend)
	render.Render(response, request, ErrNotImplemted)
}

//DeleteAdminEndpoint delete a backend with data
func DeleteAdminEndpoint(response http.ResponseWriter, request *http.Request) {
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
