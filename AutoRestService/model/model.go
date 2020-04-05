package model

import (
	"net/http"
)

type JsonMap map[string]interface{}

type BEModelRequest struct {
	JsonMap
}

func (a *BEModelRequest) Bind(r *http.Request) error {
	// a.Article is nil if no Article fields are sent in the request. Return an
	// error to avoid a nil pointer dereference.
	//	if a.JsonMap == nil {
	//		return errors.New("missing required Model fields.")
	//	}

	// a.User is nil if no Userpayload fields are sent in the request. In this app
	// this won't cause a panic, but checks in this Bind method may be required if
	// a.User or futher nested fields like a.User.Name are accessed elsewhere.

	// just a post-process after a decode..
	return nil
}
