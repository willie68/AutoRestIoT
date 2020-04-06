package worker

import (
	"errors"
	"time"

	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/internal"
	"github.com/willie68/AutoRestIoT/model"
)

var ErrMissingID = errors.New("Missing _id")

func Validate(route model.Route, data model.JsonMap) (bool, error) {
	// return false, dao.ErrNotImplemented
	return true, nil
}

//Store create a new model, or update an existing model
func Store(route model.Route, data model.JsonMap) (model.JsonMap, error) {

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
	model, err := dao.GetStorage().GetModel(route)
	if err != nil {
		return nil, err
	}
	return model, nil
}

//Update create a new model, or update an existing model
func Update(route model.Route, data model.JsonMap) (model.JsonMap, error) {
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

	modelid, err := dao.GetStorage().UpdateModel(route, data)
	if err != nil {
		return nil, err
	}
	modelData, err := Get(route)
	if err != nil {
		return nil, err
	}
	return modelData, nil
}
