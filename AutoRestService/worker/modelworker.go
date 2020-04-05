package worker

import (
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/model"
)

func Validate(route model.Route, data model.JsonMap) (bool, error) {
	// return false, dao.ErrNotImplemented
	return true, nil
}

func Store(route model.Route, data model.JsonMap) (string, error) {
	modelid, err := dao.GetStorage().CreateModel(route, data)
	if err != nil {
		return "", err
	}
	return modelid, nil
}

func Get(route model.Route) (model.JsonMap, error) {
	model, err := dao.GetStorage().GetModel(route)
	if err != nil {
		return nil, err
	}
	return model, nil
}
