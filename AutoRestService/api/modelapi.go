package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
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
	model := chi.URLParam(req, "model")
	fmt.Printf("POST: path: %s, be: %s, model: %s \n", req.URL.Path, backend, model)
	tenant := getTenant(req)
	if tenant == "" {
		Msg(response, http.StatusBadRequest, "tenant not set")
		return
	}
	log.Printf("create store for tenant %s", tenant)
	render.Status(req, http.StatusCreated)
	render.JSON(response, req, tenant)
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
	model := chi.URLParam(req, "model")
	fmt.Printf("GET many: path: %s, be: %s, model: %s \n", req.URL.Path, backend, model)
	tenant := getTenant(req)
	if tenant == "" {
		Msg(response, http.StatusBadRequest, "tenant not set")
		return
	}
	c := ConfigDescription{
		StoreID:  "myNewStore",
		TenantID: tenant,
		Size:     1234567,
	}
	render.JSON(response, req, c)
}

/*
GetManyModelsEndpoint getting if a store for a tenant is initialised
because of the automatic store creation, the value is more likely that data is stored for this tenant
*/
func GetModelEndpoint(response http.ResponseWriter, req *http.Request) {
	backend := chi.URLParam(req, "bename")
	model := chi.URLParam(req, "model")
	modelid := chi.URLParam(req, "modelid")
	fmt.Printf("GET: path: %s, be: %s, model: %s, modelid: %s  \n", req.URL.Path, backend, model, modelid)
	tenant := getTenant(req)
	if tenant == "" {
		Msg(response, http.StatusBadRequest, "tenant not set")
		return
	}
	c := ConfigDescription{
		StoreID:  "myNewStore",
		TenantID: tenant,
		Size:     1234567,
	}
	render.JSON(response, req, c)
}

/*
GetConfigSizeEndpoint size of the store for a tenant
*/
func PutModelEndpoint(response http.ResponseWriter, req *http.Request) {
	backend := chi.URLParam(req, "bename")
	model := chi.URLParam(req, "model")
	modelid := chi.URLParam(req, "modelid")
	fmt.Printf("PUT: path: %s, be: %s, model: %s, modelid: %s  \n", req.URL.Path, backend, model, modelid)
	tenant := getTenant(req)
	if tenant == "" {
		Msg(response, http.StatusBadRequest, "tenant not set")
		return
	}

	render.JSON(response, req, tenant)
}

/*
DeleteConfigEndpoint deleting store for a tenant, this will automatically delete all data in the store
*/
func DeleteModelEndpoint(response http.ResponseWriter, req *http.Request) {
	backend := chi.URLParam(req, "bename")
	model := chi.URLParam(req, "model")
	modelid := chi.URLParam(req, "modelid")
	deleteRef := isDeleteRef(req)
	fmt.Printf("DELETE: path: %s, be: %s, model: %s, modelid: %s, delRef: %t  \n", req.URL.Path, backend, model, modelid, deleteRef)
	tenant := getTenant(req)
	if tenant == "" {
		Msg(response, http.StatusBadRequest, "tenant not set")
		return
	}
	render.JSON(response, req, tenant)
}

/*
getTenant getting the tenant from the request
*/
func getTenant(req *http.Request) string {
	return req.Header.Get(TenantHeader)
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
