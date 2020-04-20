package dao

import "errors"

var ErrNoDocument = errors.New("Document not found")

var ErrUniqueIndexError = errors.New("Unique index error")

var ErrNotImplemented = errors.New("Not implemented")

var ErrUnknownError = errors.New("Unknown server error")
