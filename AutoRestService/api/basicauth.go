package api

import (
	"fmt"
	"net/http"

	"github.com/willie68/AutoRestIoT/dao"
)

// BasicAuth implements a simple middleware handler for adding basic http auth to a route.
func BasicAuth(realm string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok {
				basicAuthFailed(w, realm)
				return
			}
			salt, ok := dao.GetStorage().GetSalt(user)
			if !ok {
				basicAuthFailed(w, realm)
				return
			}
			pass = dao.BuildPasswordHash(pass, salt)
			if !dao.GetStorage().CheckUser(user, pass) {
				basicAuthFailed(w, realm)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func basicAuthFailed(w http.ResponseWriter, realm string) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}
