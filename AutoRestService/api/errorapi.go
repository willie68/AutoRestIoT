package api

import (
	"net/http"

	"github.com/go-chi/render"
)

// ErrResponse renderer type for handling all sorts of errors.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

//Render render the error automaticaly to the response
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

//ErrInvalidRequest creates a new Invalid request error response
func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

//ErrRender creates a wrapper for any error when an output can not be rendered
func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusUnprocessableEntity,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

//ErrInternalServer render a internal server error, with another error as source
func ErrInternalServer(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "Internal server error.",
		ErrorText:      err.Error(),
	}
}

//ErrValidationError error on validating objects
func ErrValidationError(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Validation error.",
		ErrorText:      err.Error(),
	}
}

//ErrNotFound requested resource was not found
var ErrNotFound = &ErrResponse{HTTPStatusCode: http.StatusNotFound, StatusText: "Resource not found."}

//ErrNotImplemted this feature/function/methode is not implemented yet
var ErrNotImplemted = &ErrResponse{HTTPStatusCode: http.StatusNotImplemented, StatusText: "Not im plemented yet."}

//ErrUniqueIndexError index violation error
var ErrUniqueIndexError = &ErrResponse{HTTPStatusCode: http.StatusBadRequest, StatusText: "Unique index violation."}

//ErrForbidden not enough rights for doing this
var ErrForbidden = &ErrResponse{HTTPStatusCode: http.StatusForbidden, StatusText: "endpoint not permitted."}
