package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/internal"
	"github.com/willie68/AutoRestIoT/model"
	"github.com/willie68/AutoRestIoT/worker"
)

const DeleteRefHeader = "X-mcs-deleteref"

/*
ModelRoutes getting all routes for the config endpoint
*/
func ModelRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.With(RoleCheck([]string{"edit"})).Post("/{bename}/{model}/", PostModelEndpoint)
	router.With(RoleCheck([]string{"edit", "read"})).With(Paginate).Get("/{bename}/{model}/", GetManyModelsEndpoint)
	router.With(RoleCheck([]string{"edit", "read"})).Get("/{bename}/{model}/{modelid}", GetModelEndpoint)
	router.With(RoleCheck([]string{"edit"})).Put("/{bename}/{model}/{modelid}", PutModelEndpoint)
	router.With(RoleCheck([]string{"edit"})).Delete("/{bename}/{model}/{modelid}", DeleteModelEndpoint)
	return router
}

//PostModelEndpoint , this method will always return 201
func PostModelEndpoint(res http.ResponseWriter, req *http.Request) {
	backend := chi.URLParam(req, "bename")
	mymodel := chi.URLParam(req, "model")
	route := model.Route{
		Backend: backend,
		Model:   mymodel,
	}
	route = enrichRouteInformation(req, route)
	data := &model.JsonMap{}

	if err := render.Decode(req, data); err != nil {
		render.Render(res, req, ErrInvalidRequest(err))
		return
	}

	bemodel := *data

	fmt.Printf("POST: path: %s, route: %s \n", req.URL.Path, route.String())
	valid, err := worker.Validate(route, bemodel)
	if err != nil {
		if err == dao.ErrNotImplemented {
			Msg(res, http.StatusNotImplemented, err.Error())
			return
		}
		Msg(res, http.StatusInternalServerError, err.Error())
		return
	}
	if !valid {
		Msg(res, http.StatusBadRequest, "data model not valid")
		return
	}
	bemodel, err = worker.Store(route, bemodel)
	if err != nil {
		if err == dao.ErrNotImplemented {
			Msg(res, http.StatusNotImplemented, err.Error())
			return
		}
		Msg(res, http.StatusInternalServerError, err.Error())
		return
	}

	route.Identity = bemodel[internal.AttributeID].(string)

	buildLocationHeader(res, req, route)

	render.Status(req, http.StatusCreated)
	render.JSON(res, req, bemodel)
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
func GetModelEndpoint(res http.ResponseWriter, req *http.Request) {
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

	model, err := worker.Get(route)
	if err != nil {
		if err == dao.ErrNotImplemented {
			Msg(res, http.StatusNotImplemented, err.Error())
			return
		}
		if err == dao.ErrNoDocument {
			render.Render(res, req, ErrNotFound)
			return
		}
		Msg(res, http.StatusInternalServerError, err.Error())
		return
	}
	render.JSON(res, req, model)
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
	username, _, _ := req.BasicAuth()
	if username != "" {
		route.Username = username
	}
	return route
}

func buildLocationHeader(res http.ResponseWriter, req *http.Request, route model.Route) {

	loc := fmt.Sprintf("%s/%s", req.URL.Path, route.Identity)
	res.Header().Add("Location", loc)
}
