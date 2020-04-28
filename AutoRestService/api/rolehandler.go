package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/dao"
)

//RoleCheck implements a simple middleware handler for adding a role check to a route.
func RoleCheck(allowedRoles []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			user, _, _ := request.BasicAuth()
			idm := dao.GetIDM()
			ok := idm.UserInRoles(user, allowedRoles)
			if !ok {
				render.Render(response, request, ErrForbidden)
				return
			}
			next.ServeHTTP(response, request)
		})
	}
}
