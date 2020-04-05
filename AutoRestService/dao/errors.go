package dao

import "errors"

var ErrNoDocument = errors.New("Document not found")

var ErrNotImplemented = errors.New("Not implemented")

var ErrUnknownError = errors.New("Unknown server error")
