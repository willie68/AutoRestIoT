package api

import (
	"errors"
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
func PostModelEndpoint(response http.ResponseWriter, request *http.Request) {
	backend := chi.URLParam(request, "bename")
	mymodel := chi.URLParam(request, "model")
	route := model.Route{
		Backend: backend,
		Model:   mymodel,
	}
	route = enrichRouteInformation(request, route)
	data := &model.JsonMap{}

	if err := render.Decode(request, data); err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}

	bemodel := *data

	fmt.Printf("POST: path: %s, route: %s \n", request.URL.Path, route.String())
	valid, err := worker.Validate(route, bemodel)
	if err != nil {
		if err == dao.ErrNotImplemented {
			render.Render(response, request, ErrNotImplemted)
			return
		}
		if err == worker.ErrBackendNotFound || err == worker.ErrBackendModelNotFound {
			render.Render(response, request, ErrNotFound)
			return
		}
		render.Render(response, request, ErrInternalServer(err))
		return
	}
	if !valid {
		render.Render(response, request, ErrInvalidRequest(errors.New("data model not valid")))
		return
	}
	bemodel, err = worker.Store(route, bemodel)
	if err != nil {
		if err == dao.ErrNotImplemented {
			render.Render(response, request, ErrNotImplemted)
			return
		}
		render.Render(response, request, ErrInternalServer(err))
		return
	}

	route.Identity = bemodel[internal.AttributeID].(string)

	buildLocationHeader(response, request, route)

	render.Status(request, http.StatusCreated)
	render.JSON(response, request, bemodel)
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
func GetModelEndpoint(response http.ResponseWriter, request *http.Request) {
	backend := chi.URLParam(request, "bename")
	mymodel := chi.URLParam(request, "model")
	modelid := chi.URLParam(request, "modelid")
	route := model.Route{
		Backend:  backend,
		Model:    mymodel,
		Identity: modelid,
	}
	route = enrichRouteInformation(request, route)
	fmt.Printf("GET: path: %s, route: %s \n", request.URL.Path, route.String())

	model, err := worker.Get(route)
	if err != nil {
		if err == dao.ErrNotImplemented {
			render.Render(response, request, ErrNotImplemted)
			return
		}
		if err == dao.ErrNoDocument {
			render.Render(response, request, ErrNotFound)
			return
		}
		render.Render(response, request, ErrInternalServer(err))
		return
	}
	render.JSON(response, request, model)
}

/*
GetConfigSizeEndpoint size of the store for a tenant
*/
func PutModelEndpoint(response http.ResponseWriter, request *http.Request) {
	backend := chi.URLParam(request, "bename")
	mymodel := chi.URLParam(request, "model")
	modelid := chi.URLParam(request, "modelid")
	route := model.Route{
		Backend:  backend,
		Model:    mymodel,
		Identity: modelid,
	}
	route = enrichRouteInformation(request, route)
	fmt.Printf("PUT: path: %s, route: %s \n", request.URL.Path, route.String())
	data := &model.JsonMap{}

	if err := render.Decode(request, data); err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}

	bemodel := *data
	valid, err := worker.Validate(route, bemodel)
	if err != nil {
		if err == dao.ErrNotImplemented {
			render.Render(response, request, ErrNotImplemted)
			return
		}
		render.Render(response, request, ErrInternalServer(err))
		return
	}
	if !valid {
		render.Render(response, request, ErrInvalidRequest(errors.New("data model not valid.")))
		return
	}

	bemodel, err = worker.Update(route, bemodel)
	if err != nil {
		if err == dao.ErrNoDocument {
			render.Render(response, request, ErrNotFound)
			return
		}
		if err == dao.ErrNotImplemented {
			render.Render(response, request, ErrNotImplemted)
			return
		}
		render.Render(response, request, ErrInternalServer(err))
		return
	}

	route.Identity = bemodel[internal.AttributeID].(string)

	buildLocationHeader(response, request, route)

	render.Status(request, http.StatusCreated)
	render.JSON(response, request, bemodel)
}

/*
DeleteConfigEndpoint deleting store for a tenant, this will automatically delete all data in the store
*/
func DeleteModelEndpoint(response http.ResponseWriter, request *http.Request) {
	backend := chi.URLParam(request, "bename")
	mymodel := chi.URLParam(request, "model")
	modelid := chi.URLParam(request, "modelid")
	route := model.Route{
		Backend:  backend,
		Model:    mymodel,
		Identity: modelid,
	}
	route = enrichRouteInformation(request, route)

	deleteRef := isDeleteRef(request)
	fmt.Printf("DELETE: path: %s,  route: %s, delRef: %t  \n", request.URL.Path, route.String(), deleteRef)

	err := worker.Delete(route, deleteRef)
	if err != nil {
		if err == dao.ErrNoDocument {
			render.Render(response, request, ErrNotFound)
			return
		}
		if err == dao.ErrNotImplemented {
			render.Render(response, request, ErrNotImplemted)
			return
		}
		render.Render(response, request, ErrInternalServer(err))
		return
	}
	MsgResponse(response, http.StatusOK, "model deleted.")
}

func isDeleteRef(request *http.Request) bool {
	deleteRef := true
	if request.Header.Get(DeleteRefHeader) != "" {
		b, err := strconv.ParseBool(request.Header.Get(DeleteRefHeader))
		if err == nil {
			deleteRef = b
		}
	}
	return deleteRef
}

func enrichRouteInformation(request *http.Request, route model.Route) model.Route {
	if request.Header.Get(APIKeyHeader) != "" {
		route.Apikey = request.Header.Get(APIKeyHeader)
	}
	if request.Header.Get(SystemHeader) != "" {
		route.SystemID = request.Header.Get(SystemHeader)
	}
	username, _, _ := request.BasicAuth()
	if username != "" {
		route.Username = username
	}
	return route
}

func buildLocationHeader(response http.ResponseWriter, request *http.Request, route model.Route) {

	loc := fmt.Sprintf("%s/%s", request.URL.Path, route.Identity)
	response.Header().Add("Location", loc)
}
