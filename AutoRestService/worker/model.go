package worker

/*
 the worker middleware is for doing some part of tranformtion on the business object side.
 Here you will find functions for validating the model validating and storage and retrieval in an storage technologie indipendent way.
*/

import (
	"errors"
	"fmt"
	"time"

	"github.com/willie68/AutoRestIoT/config"
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/internal"
	"github.com/willie68/AutoRestIoT/logging"
	"github.com/willie68/AutoRestIoT/model"
)

//ErrMissingID the id of the model is mandatory and not availble
var ErrMissingID = errors.New("Missing _id")

//ErrBackendNotFound backend with that name was not found
var ErrBackendNotFound = errors.New("Missing backend")

//ErrBackendModelNotFound backend model with that name was not found
var ErrBackendModelNotFound = errors.New("Missing backend model")

var log logging.ServiceLogger

//CheckRoute checking if the route inforamtion are ok
func CheckRoute(route model.Route) error {
	if !config.Get().AllowAnonymousBackend {
		backend, ok := model.BackendList.Get(route.Backend)
		if !ok {
			log.Alertf("backend not found: %s", route.Backend)
			return ErrBackendNotFound
		}
		_, ok = backend.GetModel(route.Model)
		if !ok {
			log.Alertf("backend model not found: %s.%s", route.Backend, route.Model)
			return ErrBackendModelNotFound
		}
	}
	return nil
}

//Validate validates the model against the definition
func Validate(route model.Route, data model.JsonMap) (bool, error) {
	// return false, dao.ErrNotImplemented
	err := CheckRoute(route)
	if err != nil {
		return false, err
	}

	return true, nil
}

//Store create a new model
func Store(route model.Route, data model.JsonMap) (model.JsonMap, error) {
	err := CheckRoute(route)
	if err != nil {
		return nil, err
	}

	if data[internal.AttributeID] != nil {
		//modelid := data["_id"]
		//		Get(route)
	}

	// adding system attributes
	data[internal.AttributeOwner] = route.Username
	data[internal.AttributeCreated] = time.Now()
	data[internal.AttributeModified] = time.Now()

	modelid, err := dao.GetStorage().CreateModel(route, data)
	if err != nil {
		return nil, err
	}
	route.Identity = modelid
	modelData, err := Get(route)
	if err != nil {
		return nil, err
	}
	return modelData, nil
}

func Get(route model.Route) (model.JsonMap, error) {
	err := CheckRoute(route)
	if err != nil {
		return nil, err
	}

	model, err := dao.GetStorage().GetModel(route)
	if err != nil {
		return nil, err
	}
	return model, nil
}

//Update update an existing model
func Update(route model.Route, data model.JsonMap) (model.JsonMap, error) {
	err := CheckRoute(route)
	if err != nil {
		return nil, err
	}

	if route.Identity == "" {
		return nil, ErrMissingID
	}

	dataModel, err := dao.GetStorage().GetModel(route)
	if err != nil {
		return nil, err
	}

	// adding system attributes
	data[internal.AttributeID] = route.Identity
	data[internal.AttributeOwner] = route.Username
	data[internal.AttributeCreated] = dataModel[internal.AttributeCreated]
	data[internal.AttributeModified] = time.Now()

	modelData, err := dao.GetStorage().UpdateModel(route, data)
	if err != nil {
		fmt.Printf("%v\n")
		return nil, err
	}
	return modelData, nil
}

//Delete delete an existing model
func Delete(route model.Route, deleteRef bool) error {
	err := CheckRoute(route)
	if err != nil {
		return err
	}

	if route.Identity == "" {
		return ErrMissingID
	}
	data, err := dao.GetStorage().GetModel(route)
	if err != nil {
		return err
	}

	beModel, ok := model.BackendList.Get(route.Backend)
	if ok {
		fmt.Printf("getting backend %s\n", beModel.Backendname)

		if beModel.IsValidDatamodel(route.Model, data) && deleteRef {
			files, err := beModel.GetReferencedFiles(route.Model, data)
			if err != nil {
				return err
			}
			for _, fileID := range files {
				err = dao.GetStorage().DeleteFile(route.Backend, fileID)
				if err != nil {
					return err
				}
			}
		}
	}
	err = dao.GetStorage().DeleteModel(route)
	if err != nil {
		return err
	}
	return nil
}
