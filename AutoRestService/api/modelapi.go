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

//DeleteRefHeader header key for the delete reference options
const DeleteRefHeader = "X-mcs-deleteref"

/*
ModelRoutes getting all routes for the config endpoint
*/
func ModelRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.With(RoleCheck([]string{"edit"})).Post("/{bename}/{model}/", PostModelEndpoint)
	//TODO insert bulk import api
	router.With(RoleCheck([]string{"edit", "read"})).Get("/{bename}/{model}/count", GetModelCountEndpoint)
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

	log.Infof("POST: path: %s, route: %s \n", request.URL.Path, route.String())
	validModel, err := worker.Validate(route, bemodel)
	if err != nil {
		if pe, ok := err.(worker.ErrValidationError); ok {
			render.Render(response, request, ErrInvalidRequest(pe))
			return
		}
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
	if validModel == nil {
		render.Render(response, request, ErrInvalidRequest(errors.New("data model not valid")))
		return
	}
	validModel, err = worker.Store(route, validModel)
	if err != nil {
		if err == dao.ErrNotImplemented {
			render.Render(response, request, ErrNotImplemted)
			return
		}
		if err == dao.ErrUniqueIndexError {
			render.Render(response, request, ErrUniqueIndexError)
			return
		}
		render.Render(response, request, ErrInternalServer(err))
		return
	}

	route.Identity = validModel[internal.AttributeID].(string)

	buildLocationHeader(response, request, route)

	render.Status(request, http.StatusCreated)
	render.JSON(response, request, validModel)
}

//GetManyModelsEndpoint searching for a list of model instances
func GetManyModelsEndpoint(response http.ResponseWriter, request *http.Request) {
	offset := request.Context().Value(contextKeyOffset)
	limit := request.Context().Value(contextKeyLimit)
	log.Infof("GET many: offset: %d, limit: %d", offset, limit)

	backend := chi.URLParam(request, "bename")
	mymodel := chi.URLParam(request, "model")
	route := model.Route{
		Backend: backend,
		Model:   mymodel,
	}
	route = enrichRouteInformation(request, route)
	log.Infof("GET many: path: %s, route: %s", request.URL.Path, route.String())

	query := request.URL.Query().Get("query")

	log.Infof("query: %s, offset: %d, limit: %d", query, offset, limit)
	//owner, _, _ := request.BasicAuth()

	n, models, err := worker.Query(route, query, offset.(int), limit.(int))
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
	m["offset"] = offset
	m["limit"] = limit

	render.JSON(response, request, m)
}

//GetModelCountEndpoint getting a model with an identifier
func GetModelCountEndpoint(response http.ResponseWriter, request *http.Request) {
	backend := chi.URLParam(request, "bename")
	mymodel := chi.URLParam(request, "model")
	route := model.Route{
		Backend: backend,
		Model:   mymodel,
	}
	route = enrichRouteInformation(request, route)
	log.Infof("GET count: path: %s, route: %s", request.URL.Path, route.String())

	modelCount, err := worker.GetCount(route)
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
	m := make(map[string]interface{})
	m["found"] = modelCount

	render.JSON(response, request, m)
}

//GetModelEndpoint getting a model with an identifier
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
	log.Infof("GET: path: %s, route: %s", request.URL.Path, route.String())

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

//PutModelEndpoint change a model
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
	log.Infof("PUT: path: %s, route: %s", request.URL.Path, route.String())
	data := &model.JsonMap{}

	if err := render.Decode(request, data); err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}

	bemodel := *data
	validModel, err := worker.Validate(route, bemodel)
	if err != nil {
		if err == dao.ErrNotImplemented {
			render.Render(response, request, ErrNotImplemted)
			return
		}
		render.Render(response, request, ErrInternalServer(err))
		return
	}
	if validModel == nil {
		render.Render(response, request, ErrInvalidRequest(errors.New("data model not valid")))
		return
	}

	validModel, err = worker.Update(route, validModel)
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

	route.Identity = validModel[internal.AttributeID].(string)

	buildLocationHeader(response, request, route)

	render.Status(request, http.StatusCreated)
	render.JSON(response, request, validModel)
}

//DeleteModelEndpoint deleting a model
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
	log.Infof("DELETE: path: %s,  route: %s, delRef: %t", request.URL.Path, route.String(), deleteRef)

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
