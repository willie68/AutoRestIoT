package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/model"
)

// TagsRoutes getting all routes for the tags endpoint
func UsersRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", GetUserEndpoint)
	router.Post("/", PostUserEndpoint)
	router.Put("/{username}", PutUserEndpoint)
	router.Delete("/{username}", DeleteUserEndpoint)
	return router
}

//PutUserEndpoint getting all tags back. No paging...
func GetUserEndpoint(response http.ResponseWriter, req *http.Request) {
	username, _, _ := req.BasicAuth()
	user, ok := dao.GetStorage().GetUser(username)
	user.Password = ""
	user.NewPassword = ""
	if !ok {
		Msg(response, http.StatusInternalServerError, "")
		return
	}
	render.JSON(response, req, user)
}

//PutUserEndpoint getting all tags back. No paging...
func PostUserEndpoint(response http.ResponseWriter, req *http.Request) {
	var user model.User
	err := render.DefaultDecoder(req, &user)
	if err != nil {
		Msg(response, http.StatusBadRequest, err.Error())
		return
	}

	adminusername, _, _ := req.BasicAuth()
	admin, ok := dao.GetStorage().GetUser(adminusername)
	if !ok {
		Msg(response, http.StatusInternalServerError, "")
		return
	}
	if !admin.Admin {
		Msg(response, http.StatusForbidden, "permission denied")
		return
	}

	err = dao.GetStorage().AddUser(user)
	if err != nil {
		Msg(response, http.StatusBadRequest, err.Error())
		return
	}
	Msg(response, http.StatusCreated, fmt.Sprintf("user \"%s\" created sucessfully", user.Name))
}

//PutUserEndpoint getting all tags back. No paging...
func PutUserEndpoint(response http.ResponseWriter, req *http.Request) {
	username := chi.URLParam(req, "username")
	var user model.User
	err := render.DefaultDecoder(req, &user)
	if err != nil {
		Msg(response, http.StatusBadRequest, err.Error())
		return
	}
	if username != user.Name {
		Msg(response, http.StatusBadRequest, "username should be identically")
		return
	}
	adminusername, _, _ := req.BasicAuth()
	admin, ok := dao.GetStorage().GetUser(adminusername)
	if !ok {
		Msg(response, http.StatusInternalServerError, "")
		return
	}
	if (adminusername != username) && !admin.Admin {
		Msg(response, http.StatusForbidden, "permission denied")
		return
	}

	err = dao.GetStorage().ChangePWD(username, user.NewPassword, user.Password)
	if err != nil {
		Msg(response, http.StatusBadRequest, err.Error())
		return
	}
	return
}

//PutUserEndpoint getting all tags back. No paging...
func DeleteUserEndpoint(response http.ResponseWriter, req *http.Request) {
	username := chi.URLParam(req, "username")
	adminusername, _, _ := req.BasicAuth()
	admin, ok := dao.GetStorage().GetUser(adminusername)
	if !ok {
		Msg(response, http.StatusInternalServerError, "")
		return
	}
	if !admin.Admin {
		Msg(response, http.StatusForbidden, "permission denied")
		return
	}

	err := dao.GetStorage().DeleteUser(username)
	if err != nil {
		Msg(response, http.StatusBadRequest, err.Error())
		return
	}
	return
}
