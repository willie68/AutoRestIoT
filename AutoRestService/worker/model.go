package worker

/*
 the worker middleware is for doing some part of tranformtion on the business object side.
 Here you will find functions for validating the model validating and storage and retrieval in an storage technologie indipendent way.
*/

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
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

type ErrValidationError struct {
	message string
}

func (p ErrValidationError) Error() string {
	return p.message
}

var modelCount = make(map[string]int)
var modelCountMutex = &sync.Mutex{}

func GetModelCount() map[string]int {
	internaleCountMap := make(map[string]int)
	modelCountMutex.Lock()
	for k, v := range modelCount {
		internaleCountMap[k] = v
	}
	modelCountMutex.Unlock()
	return internaleCountMap
}

func IncModelCounter(name string, count int) {
	modelCountMutex.Lock()
	modelCount[name] = modelCount[name] + count
	modelCountMutex.Unlock()
}

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

//Validate validates the model against the definition, and convert attributes, if there is something to convert (like dateTime attributes)
func Validate(route model.Route, data model.JSONMap) (model.JSONMap, error) {
	// return false, dao.ErrNotImplemented
	err := CheckRoute(route)
	if err != nil {
		return nil, err
	}
	modelDefinition, ok := model.BackendList.GetModel(route)
	if !ok || !config.Get().AllowAnonymousBackend {
		return nil, ErrBackendModelNotFound
	}
	log.Info(modelDefinition.Name)
	// check fieldtypes, eventually convert
	//TODO check field values
	for key, value := range data {
		field, ok := modelDefinition.GetField(key)
		if !ok {
			continue
		}
		if field.Mandatory {
			if isEmpty(value) {
				return nil, ErrValidationError{
					message: fmt.Sprintf("mandatory field \"%s\" is empty", key),
				}
			}
		}

		if field.Collection {
			if !isEmpty(value) {
				//TODO check and convert every array entry
				if reflect.TypeOf(value).Kind() != reflect.Slice {
					return nil, ErrValidationError{
						message: fmt.Sprintf("collection field \"%s\" is not a collection", key),
					}
				}
			}
		}

		if field.Type == model.FieldTypeBool {
			switch v := value.(type) {
			case float64:
				data[key] = v > 0
			case int:
				data[key] = v > 0
			case bool:
				data[key] = v
			}
		}

		if field.Type == model.FieldTypeTime {
			switch v := value.(type) {
			case string:
				layout := "2006-01-02T15:04:05.000Z07:00"
				time, err := time.Parse(layout, v)
				if err != nil {
					return nil, ErrValidationError{
						message: fmt.Sprintf("wrong time format: key: \"%s\", err: %v", key, err),
					}
				}
				data[key] = time
			case float64:
				data[key] = time.Unix(0, int64(v)*int64(time.Millisecond))
			}
		}
	}

	//check mandatory fields
	for _, field := range modelDefinition.Fields {
		if field.Mandatory {
			if isEmpty(data[field.Name]) {
				return nil, ErrValidationError{
					message: fmt.Sprintf("mandatory field \"%s\" is empty", field.Name),
				}
			}
		}
	}
	return data, nil
}

func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return v == ""
	}
	return false
}

//Store create a new model
func Store(route model.Route, data model.JSONMap) (model.JSONMap, error) {
	err := CheckRoute(route)
	if err != nil {
		return nil, err
	}

	// adding system attributes
	data[internal.AttributeOwner] = route.Username
	data[internal.AttributeCreated] = time.Now()
	data[internal.AttributeModified] = time.Now()

	modelid, err := dao.GetStorage().CreateModel(route, data)
	if err != nil {
		return nil, err
	}
	IncModelCounter(route.GetRouteName(), 1)
	route.Identity = modelid
	modelData, err := Get(route)
	if err != nil {
		return nil, err
	}
	return modelData, nil
}

//StoreMany create a bunch of new model
func StoreMany(route model.Route, datas []model.JSONMap) ([]string, error) {
	err := CheckRoute(route)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for _, data := range datas {

		// adding system attributes
		data[internal.AttributeOwner] = route.Username
		data[internal.AttributeCreated] = now
		data[internal.AttributeModified] = now

	}

	modelids, err := dao.GetStorage().CreateModels(route, datas)
	if err != nil {
		return nil, err
	}
	return modelids, nil
}

//Get getting one model
func Get(route model.Route) (model.JSONMap, error) {
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
func Update(route model.Route, data model.JSONMap) (model.JSONMap, error) {
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
		fmt.Printf("%v\n", err)
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

//Query query for existing models
func Query(route model.Route, query string, offset int, limit int) (int, []model.JSONMap, error) {
	err := CheckRoute(route)
	if err != nil {
		return 0, nil, err
	}

	n, dataModels, err := dao.GetStorage().QueryModel(route, query, offset, limit)
	if err != nil {
		return 0, nil, err
	}

	return n, dataModels, nil
}

//GetCount query for existing models
func GetCount(route model.Route) (int, error) {
	err := CheckRoute(route)
	if err != nil {
		return 0, err
	}

	n, err := dao.GetStorage().CountModel(route)
	if err != nil {
		return 0, err
	}

	return n, nil
}
