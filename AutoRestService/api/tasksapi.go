package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/worker"
)

//AdminRoutes getting all routes for the config endpoint
func TasksRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.With(RoleCheck([]string{"admin"})).Get("/", GetAdminTasksHandler)
	return router
}

// GetAdminTasksHandler getting server info
func GetAdminTasksHandler(response http.ResponseWriter, request *http.Request) {
	log.Infof("GET: path: %s", request.URL.Path)
	route := worker.GetTaskRoute()

	log.Infof("GET many: path: %s, route: %s", request.URL.Path, route.String())

	query := ""

	n, models, err := worker.Query(route, query, 0, 0)
	if err != nil {
		if err == dao.ErrNotImplemented {
			render.Render(response, request, ErrNotImplemted)
			return
		}
		render.Render(response, request, ErrInternalServer(err))
		return
	}

	m := make(map[string]interface{})
	m["data"] = models
	m["found"] = n
	m["count"] = len(models)
	m["query"] = query
	m["offset"] = 0
	m["limit"] = 0

	render.JSON(response, request, m)
}
