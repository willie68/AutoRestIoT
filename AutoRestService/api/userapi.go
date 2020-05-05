package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/model"
)

// UsersRoutes routes to the user interface
func UsersRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/me", GetMeEndpoint)
	router.With(RoleCheck([]string{"admin"})).Get("/", GetUsersEndpoint)
	router.With(RoleCheck([]string{"admin"})).Post("/", PostUserEndpoint)
	router.With(RoleCheck([]string{"admin"})).Get("/{username}", GetUserEndpoint)
	router.With(RoleCheck([]string{"admin"})).Put("/{username}", PutUserEndpoint)
	router.With(RoleCheck([]string{"admin"})).Delete("/{username}", DeleteUserEndpoint)
	return router
}

//GetMeEndpoint getting all user infos
func GetMeEndpoint(response http.ResponseWriter, request *http.Request) {
	username, _, _ := request.BasicAuth()
	user, ok := dao.GetStorage().GetUser(username)
	if !ok {
		render.Render(response, request, ErrInternalServer(errors.New("")))
		return
	}
	user.Password = ""
	user.NewPassword = ""
	user.Salt = []byte{}
	render.JSON(response, request, user)
}

//GetUsersEndpoint getting all user infos
func GetUsersEndpoint(response http.ResponseWriter, request *http.Request) {
	users, err := dao.GetStorage().GetUsers()
	if err != nil {
		render.Render(response, request, ErrInternalServer(err))
		return
	}
	myUsers := make([]model.User, 0)
	for _, user := range users {
		user.Password = ""
		user.NewPassword = ""
		user.Salt = []byte{}
		myUsers = append(myUsers, user)
	}
	render.JSON(response, request, myUsers)
}

//GetUserEndpoint getting a user info
func GetUserEndpoint(response http.ResponseWriter, request *http.Request) {
	username := chi.URLParam(request, "username")
	user, ok := dao.GetStorage().GetUser(username)
	user.Password = ""
	user.NewPassword = ""
	user.Salt = []byte{}
	if !ok {
		render.Render(response, request, ErrInternalServer(errors.New("")))
		return
	}
	render.JSON(response, request, user)
}

//PostUserEndpoint adding a new user
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
	idm := dao.GetIDM()

	user, err = idm.AddUser(user)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	user.Salt = []byte{}
	user.Password = "#####"
	user.NewPassword = ""
	render.Status(request, http.StatusCreated)
	render.JSON(response, request, user)
}

//PutUserEndpoint change password of user
func PutUserEndpoint(response http.ResponseWriter, request *http.Request) {
	username := chi.URLParam(request, "username")
	var user model.User
	err := render.DefaultDecoder(request, &user)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	if username != user.Name {
		render.Render(response, request, ErrInvalidRequest(errors.New("username should be identically")))
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
	idm := dao.GetIDM()

	err = idm.ChangePWD(username, user.NewPassword, user.Password)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	user.Salt = []byte{}
	user.Password = "#####"
	user.NewPassword = ""
	render.JSON(response, request, user)
}

//DeleteUserEndpoint deleting a user
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
	idm := dao.GetIDM()

	err := idm.DeleteUser(username)
	if err != nil {
		render.Render(response, request, ErrInvalidRequest(err))
		return
	}
	return
}
