package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/model"
)

const DeleteRefHeader = "X-mcs-deleteref"

/*
ModelRoutes getting all routes for the config endpoint
*/
func ModelRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/{bename}/{model}/", PostModelEndpoint)
	router.With(Paginate).Get("/{bename}/{model}/", GetManyModelsEndpoint)
	router.Get("/{bename}/{model}/{modelid}", GetModelEndpoint)
	router.Put("/{bename}/{model}/{modelid}", PutModelEndpoint)
	router.Delete("/{bename}/{model}/{modelid}", DeleteModelEndpoint)
	return router
}

//PostModelEndpoint , this method will always return 201
func PostModelEndpoint(response http.ResponseWriter, req *http.Request) {
	backend := chi.URLParam(req, "bename")
	mymodel := chi.URLParam(req, "model")
	route := model.Route{
		Backend: backend,
		Model:   mymodel,
	}
	route = enrichRouteInformation(req, route)
	fmt.Printf("POST: path: %s, route: %s \n", req.URL.Path, route.String())
	render.Status(req, http.StatusCreated)
	render.JSON(response, req, route)
}

/*
GetManyModelsEndpoint getting if a store for a tenant is initialised
because of the automatic store creation, the value is more likely that data is stored for this tenant
*/
func GetManyModelsEndpoint(response http.ResponseWriter, req *http.Request) {
	offset := req.Context().Value(contextKeyOffset)
	limit := req.Context().Value(contextKeyLimit)
	fmt.Printf("GET many: offset: %d, limit: %d\n", offset, limit)

	backend := chi.URLParam(req, "bename")
	mymodel := chi.URLParam(req, "model")
	route := model.Route{
		Backend: backend,
		Model:   mymodel,
	}
	route = enrichRouteInformation(req, route)
	fmt.Printf("GET many: path: %s, route: %s \n", req.URL.Path, route.String())

	render.JSON(response, req, route)
}

/*
GetManyModelsEndpoint getting if a store for a tenant is initialised
because of the automatic store creation, the value is more likely that data is stored for this tenant
*/
func GetModelEndpoint(response http.ResponseWriter, req *http.Request) {
	backend := chi.URLParam(req, "bename")
	mymodel := chi.URLParam(req, "model")
	modelid := chi.URLParam(req, "modelid")
	route := model.Route{
		Backend:  backend,
		Model:    mymodel,
		Identity: modelid,
	}
	route = enrichRouteInformation(req, route)
	fmt.Printf("GET: path: %s, route: %s \n", req.URL.Path, route.String())

	render.JSON(response, req, route)
}

/*
GetConfigSizeEndpoint size of the store for a tenant
*/
func PutModelEndpoint(response http.ResponseWriter, req *http.Request) {
	backend := chi.URLParam(req, "bename")
	mymodel := chi.URLParam(req, "model")
	modelid := chi.URLParam(req, "modelid")
	route := model.Route{
		Backend:  backend,
		Model:    mymodel,
		Identity: modelid,
	}
	route = enrichRouteInformation(req, route)
	fmt.Printf("PUT: path: %s, route: %s \n", req.URL.Path, route.String())
	render.JSON(response, req, route)
}

/*
DeleteConfigEndpoint deleting store for a tenant, this will automatically delete all data in the store
*/
func DeleteModelEndpoint(response http.ResponseWriter, req *http.Request) {
	backend := chi.URLParam(req, "bename")
	mymodel := chi.URLParam(req, "model")
	modelid := chi.URLParam(req, "modelid")
	route := model.Route{
		Backend:  backend,
		Model:    mymodel,
		Identity: modelid,
	}
	route = enrichRouteInformation(req, route)
	deleteRef := isDeleteRef(req)
	fmt.Printf("DELETE: path: %s,  route: %s, delRef: %t  \n", req.URL.Path, route.String(), deleteRef)
	render.JSON(response, req, route)
}

func isDeleteRef(req *http.Request) bool {
	deleteRef := true
	if req.Header.Get(DeleteRefHeader) != "" {
		b, err := strconv.ParseBool(req.Header.Get(DeleteRefHeader))
		if err == nil {
			deleteRef = b
		}
	}
	return deleteRef
}

func enrichRouteInformation(req *http.Request, route model.Route) model.Route {
	if req.Header.Get(APIKeyHeader) != "" {
		route.Apikey = req.Header.Get(APIKeyHeader)
	}
	if req.Header.Get(SystemHeader) != "" {
		route.SystemID = req.Header.Get(SystemHeader)
	}
	return route
}
