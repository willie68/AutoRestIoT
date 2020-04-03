package api

import (
	"log"
	"net/http"
	"strings"
)

// APIKeyHeader in this header thr right api key should be inserted
const APIKeyHeader = "X-mcs-apikey"

// SystemHeader in this header thr right system should be inserted
const SystemHeader = "X-mcs-system"

/*
SysAPIKey defining a handler for checking system id and api key
*/
type SysAPIKey struct {
	log      *log.Logger
	SystemID string
	Apikey   string
}

/*
NewSysAPIHandler creates a new SysApikeyHandler
*/
func NewSysAPIHandler(systemID string, apikey string) *SysAPIKey {
	c := &SysAPIKey{
		SystemID: systemID,
		Apikey:   apikey,
	}
	return c
}

/*
Handler the handler checks systemid and apikey headers
*/
func (s *SysAPIKey) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSuffix(r.URL.Path, "/")
		if !strings.HasPrefix(path, "/health") {
			if s.SystemID != r.Header.Get(SystemHeader) {
				Msg(w, http.StatusBadRequest, "either system id or apikey not correct")
				return
			}
			if s.Apikey != strings.ToLower(r.Header.Get(APIKeyHeader)) {
				Msg(w, http.StatusBadRequest, "either system id or apikey not correct")
				return
			}
		}
		next.ServeHTTP(w, r)
	})

}

/*
AddHeader adding gefault header for system and apikey
*/
func AddHeader(response http.ResponseWriter, apikey string, system string) {
	response.Header().Add(APIKeyHeader, apikey)
	response.Header().Add(SystemHeader, system)
}
