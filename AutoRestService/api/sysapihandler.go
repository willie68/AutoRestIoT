package api

import (
	"context"
	"log"
	"net/http"
	"strconv"
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

var (
	contextKeyOffset = contextKey("offset")
	contextKeyLimit  = contextKey("limit")
)

// Paginate is a middleware logic for populating the context with offset and limit values
func Paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		offsetStr := r.URL.Query().Get("offset")
		limitStr := r.URL.Query().Get("limit")
		if offsetStr != "" {
			offset, err := strconv.Atoi(offsetStr)
			if err != nil {
				Msg(w, http.StatusBadRequest, err.Error())
				return
			}
			ctx = context.WithValue(ctx, contextKeyOffset, offset)
		} else {
			ctx = context.WithValue(ctx, contextKeyOffset, 0)
		}
		if limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				Msg(w, http.StatusBadRequest, err.Error())
				return
			}
			ctx = context.WithValue(ctx, contextKeyLimit, limit)
		} else {
			ctx = context.WithValue(ctx, contextKeyLimit, 0)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type contextKey string

func (c contextKey) String() string {
	return "api" + string(c)
}
