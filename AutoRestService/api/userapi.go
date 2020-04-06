package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/model"
)

// TagsRoutes getting all routes for the tags endpoint
func UsersRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.With(RoleCheck([]string{"admin"})).Get("/", GetUsersEndpoint)
	router.With(RoleCheck([]string{"admin"})).Post("/", PostUserEndpoint)
	router.With(RoleCheck([]string{"admin"})).Get("/{username}", GetUserEndpoint)
	router.With(RoleCheck([]string{"admin"})).Put("/{username}", PutUserEndpoint)
	router.With(RoleCheck([]string{"admin"})).Delete("/{username}", DeleteUserEndpoint)
	return router
}

//GetUsersEndpoint getting all tags back. No paging...
func GetUsersEndpoint(response http.ResponseWriter, request *http.Request) {
	users, err := dao.GetStorage().GetUsers()
	if err != nil {
		render.Render(response, request, ErrInternalServer(err))
		return
	}
	render.JSON(response, request, users)
}

//GetUserEndpoint getting all tags back. No paging...
func GetUserEndpoint(response http.ResponseWriter, request *http.Request) {
	username := chi.URLParam(request, "username")
	user, ok := dao.GetStorage().GetUser(username)
	user.Password = ""
	user.NewPassword = ""
	if !ok {
		render.Render(response, request, ErrInternalServer(errors.New("")))
		return
	}
	render.JSON(response, request, user)
}

//PostUserEndpoint getting all tags back. No paging...
func PostUserEndpoint(response http.ResponseWriter, request *http.Request) {
	var user model.User
	err := render.DefaultDecoder(request, &user)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}

	adminusername, _, _ := request.BasicAuth()
	admin, ok := dao.GetStorage().GetUser(adminusername)
	if !ok {
		render.Render(response, request, ErrInternalServer(errors.New("")))
		return
	}
	if !admin.Admin {
		render.Render(response, request, ErrForbidden)
		return
	}

	err = dao.GetStorage().AddUser(user)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	user.Password = "#####"
	user.NewPassword = ""
	render.Status(request, http.StatusCreated)
	render.JSON(response, request, user)
}

//PutUserEndpoint getting all tags back. No paging...
func PutUserEndpoint(response http.ResponseWriter, request *http.Request) {
	username := chi.URLParam(request, "username")
	var user model.User
	err := render.DefaultDecoder(request, &user)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	if username != user.Name {
		render.Render(response, request, ErrInvalidRequest(errors.New("username should be identically.")))
		return
	}
	adminusername, _, _ := request.BasicAuth()
	admin, ok := dao.GetStorage().GetUser(adminusername)
	if !ok {
		render.Render(response, request, ErrInternalServer(errors.New("")))
		return
	}
	if (adminusername != username) && !admin.Admin {
		render.Render(response, request, ErrForbidden)
		return
	}

	err = dao.GetStorage().ChangePWD(username, user.NewPassword, user.Password)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	user.Password = "#####"
	user.NewPassword = ""
	render.JSON(response, request, user)
}

//PutUserEndpoint getting all tags back. No paging...
func DeleteUserEndpoint(response http.ResponseWriter, request *http.Request) {
	username := chi.URLParam(request, "username")
	adminusername, _, _ := request.BasicAuth()
	admin, ok := dao.GetStorage().GetUser(adminusername)
	if !ok {
		render.Render(response, request, ErrInternalServer(errors.New("")))
		return
	}
	if !admin.Admin {
		render.Render(response, request, ErrForbidden)
		return
	}

	err := dao.GetStorage().DeleteUser(username)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	return
}
