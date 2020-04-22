package dao

import "errors"

//ErrNoDocument the requested document was not found in the backend
var ErrNoDocument = errors.New("Document not found")

//ErrUniqueIndexError there is a unique index violation
var ErrUniqueIndexError = errors.New("Unique index error")

//ErrNotImplemented the desired feature/function/method is not implemented
var ErrNotImplemented = errors.New("Not implemented")

//ErrUnknownError i don't know what this error is about
var ErrUnknownError = errors.New("Unknown server error")
