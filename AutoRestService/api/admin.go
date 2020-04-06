package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/dao"
)

//AdminRoutes getting all routes for the config endpoint
func AdminRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.With(RoleCheck([]string{"admin"})).Post("/{bename}/", PostAdminHandler)
	router.With(RoleCheck([]string{"admin"})).Delete("/{bename}/", DeleteAdminEndpoint)
	return router
}

// PostAdminHandler create a new backend
func PostAdminHandler(response http.ResponseWriter, request *http.Request) {
	backend := chi.URLParam(request, "bename")
	fmt.Printf("POST: path: %s, be: %s\n", request.URL.Path, backend)
	render.Render(response, request, ErrNotImplemted)
}

//DeleteAdminEndpoint delete a backend with data
func DeleteAdminEndpoint(response http.ResponseWriter, request *http.Request) {
	backend := chi.URLParam(request, "bename")
	fmt.Printf("DELETE: path: %s, be: %s\n", request.URL.Path, backend)

	err := dao.GetStorage().DeleteBackend(backend)
	if err != nil {
		fmt.Printf("%v\n", err)
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}

	m := make(map[string]interface{})
	m["backend"] = backend
	m["msg"] = fmt.Sprintf("backend %s deleted. All data destroyed.", backend)

	render.JSON(response, request, m)
}
