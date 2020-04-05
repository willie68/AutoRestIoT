package api

import (
	"net/http"

	"github.com/willie68/AutoRestIoT/dao"
)

// BasicAuth implements a simple middleware handler for adding basic http auth to a route.
func RoleCheck(allowedRoles []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, _, _ := r.BasicAuth()
			ok := dao.GetStorage().UserInRoles(user, allowedRoles)
			if !ok {
				Msg(w, http.StatusForbidden, "endpoint not allowed")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
