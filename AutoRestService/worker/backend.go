package worker

import "github.com/willie68/AutoRestIoT/model"

//ValidateBackend validate if a backend definition is valid
func ValidateBackend(be model.Backend) error {
	// checking backendname format
	// checking models
	// checking indexes
	return nil
}

func RegisterBackend(backend model.Backend) error {
	// create indexes if missing
	model.BackendList.Get
	return nil
}
