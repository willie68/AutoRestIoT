package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/yaml.v2"
)

const DataSourceTypeMQTT = "mqtt"
const DataSourceTypeREST = "rest"

const DataSinkTypeMQTT = "mqtt"

var BackendStorageRoute model.Route

func init() {
}

//ValidateBackend validate if a backend definition is valid
func ValidateBackend(be model.Backend) error {
	// checking backendname format
	// checking models
	// checking indexes
	// checking datasources
	// checking rules
	// checking destinations
	return nil
}

//PrepareBackend will mmaily prepare the configuration of the datasources and destination to the right config type
func PrepareBackend(backend model.Backend) (model.Backend, error) {
	for i, dataSource := range backend.DataSources {
		configJSON, err := json.Marshal(dataSource.Config)
		switch dataSource.Type {
		case DataSourceTypeMQTT:
			var config model.DataSourceConfigMQTT
			if err = json.Unmarshal(configJSON, &config); err != nil {
				return backend, errors.New(fmt.Sprintf("backend: %s, unmarshall mqtt config: %q", backend.Backendname, dataSource.Type))
			}
			backend.DataSources[i].Config = config
		case DataSourceTypeREST:
			var config model.DataSourceConfigREST
			if err = json.Unmarshal(configJSON, &config); err != nil {
				return backend, errors.New(fmt.Sprintf("backend: %s, unmarshall mqtt config: %q", backend.Backendname, dataSource.Type))
			}
			backend.DataSources[i].Config = config
		default:
			return backend, errors.New(fmt.Sprintf("backend: %s, unknown datasource type: %q", backend.Backendname, dataSource.Type))
		}
	}
	for i, destination := range backend.Destinations {
		configJSON, err := json.Marshal(destination.Config)
		switch destination.Type {
		case DataSinkTypeMQTT:
			var config model.DataSourceConfigMQTT
			if err = json.Unmarshal(configJSON, &config); err != nil {
				return backend, errors.New(fmt.Sprintf("backend: %s, unmarshall mqtt config: %q", backend.Backendname, destination.Type))
			}
			backend.Destinations[i].Config = config
		default:
			return backend, errors.New(fmt.Sprintf("backend: %s, unknown destination type: %q", backend.Backendname, destination.Type))
		}
	}
	return backend, nil
}

//RegisterBackend will create the needed indexes for the models and create the datasources, rules and destinations
func RegisterBackend(backend model.Backend) error {
	// create indexes if missing
	models := backend.Models
	for _, bemodel := range models {
		err := createIndex(bemodel, backend.Backendname)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
	// creating source plugins
	for _, datasource := range backend.DataSources {
		ok := false
		for !ok {
			err := createDatasource(datasource, backend.Backendname)
			if err != nil {
				log.Fatalf("%v", err)
				time.Sleep(10 * time.Second)
				continue
			}
			ok = true
		}
	}

	for _, rule := range backend.Rules {
		ok := false
		for !ok {
			err := createRule(rule, backend.Backendname)
			if err != nil {
				log.Fatalf("%v", err)
				time.Sleep(10 * time.Second)
				continue
			}
			ok = true
		}
	}

	for _, destination := range backend.Destinations {
		ok := false
		for !ok {
			err := Destinations.Register(backend.Backendname, destination)
			if err != nil {
				log.Fatalf("%v", err)
				time.Sleep(10 * time.Second)
				continue
			}
			ok = true
		}
	}

	return nil
}

func createDatasource(datasource model.DataSource, backendname string) error {
	switch datasource.Type {
	case "mqtt":
		clientID := fmt.Sprintf("autorestIoT.%s.%s", backendname, datasource.Name)
		err := mqttRegisterTopic(clientID, backendname, datasource)
		if err != nil {
			return err
		}
	default:
		log.Alertf("type \"%s\" is not availble as data source type", datasource.Type)
	}
	return nil
}

func destroyDatasource(datasource model.DataSource, backendname string) error {
	switch datasource.Type {
	case "mqtt":
		clientID := fmt.Sprintf("autorestIoT.%s.%s", backendname, datasource.Name)
		err := mqttDeregisterTopic(clientID, backendname, datasource)
		if err != nil {
			return err
		}
	}
	return nil
}

func createRule(rule model.Rule, backendname string) error {
	json, err := json.Marshal(rule.Transform)
	if err != nil {
		return err
	}
	err = Rules.Register(backendname, rule.Name, string(json))
	if err != nil {
		return err
	}
	return nil
}

func destroyRule(rule model.Rule, backendname string) error {
	err := Rules.Deregister(backendname, rule.Name)
	if err != nil {
		return err
	}
	return nil
}

func createIndex(bemodel model.Model, backendname string) error {
	indexes := bemodel.Indexes
	// define stardard fulltext index
	_, ok := bemodel.GetIndex(dao.FulltextIndexName)
	if !ok {
		fulltextIndex := model.Index{
			Name:   dao.FulltextIndexName,
			Fields: bemodel.GetFieldNames(),
		}
		indexes = append(indexes, fulltextIndex)
	}
	// define stardard indexes
	for _, field := range bemodel.Fields {
		_, ok := bemodel.GetIndex(dao.FulltextIndexName)
		if !ok {
			index := model.Index{
				Name:   field.Name,
				Fields: []string{field.Name},
			}
			indexes = append(indexes, index)
		}
	}
	// Delete unused indexes
	route := model.Route{
		Backend: backendname,
		Model:   bemodel.Name,
	}
	names, err := dao.GetStorage().GetIndexNames(route)
	if err != nil {
		return err
	}
	for _, idxName := range names {
		found := false
		for _, index := range indexes {
			if idxName == index.Name {
				found = true
				break
			}
		}
		if !found {
			err = dao.GetStorage().DeleteIndex(route, idxName)
		}
	}
	for _, index := range indexes {
		err := dao.GetStorage().UpdateIndex(route, index)
		if err != nil {
			return err
		}
	}
	return nil
}

//DeregisterBackend will destroy all datasources, Rules and destinations and will remove the backend from the internal backendlist.
func DeregisterBackend(backendname string) error {
	backend, ok := model.BackendList.Get(backendname)
	if ok {
		for _, datasource := range backend.DataSources {
			ok := false
			for !ok {
				err := destroyDatasource(datasource, backend.Backendname)
				if err != nil {
					log.Fatalf("%v", err)
					return err
				}
				ok = true
			}
		}

		for _, rule := range backend.Rules {
			ok := false
			for !ok {
				err := destroyRule(rule, backend.Backendname)
				if err != nil {
					log.Fatalf("%v", err)
					return err
				}
				ok = true
			}
		}

		for _, destination := range backend.Destinations {
			ok := false
			for !ok {
				err := Destinations.Deregister(backend.Backendname, destination)
				if err != nil {
					log.Fatalf("%v", err)
					return err
				}
				ok = true
			}
		}

		model.BackendList.Remove(backendname)
	}
	return nil
}

//StoreBackend will save the backend definition to the storage. If its already there, it will be updated
func StoreBackend(backend model.Backend) (string, error) {
	update := false
	id := ""
	query := fmt.Sprintf("{\"backendname\": \"%s\"}", backend.Backendname)
	count, bemodels, err := dao.GetStorage().QueryModel(BackendStorageRoute, query, 0, 10)
	if err != nil {
		log.Alertf("%v", err)
		return "", err
	}
	if count > 0 {
		update = true
		bemodel := model.JSONMap(bemodels[0])
		id = bemodel["_id"].(primitive.ObjectID).Hex()
		log.Infof("found backend with id: %s", id)
	}
	jsonString, err := json.Marshal(backend)
	if err != nil {
		return "", err
	}

	jsonModel := model.JSONMap{}
	err = yaml.Unmarshal(jsonString, &jsonModel)
	if err != nil {
		return "", err
	}
	if update {
		route := model.Route{
			Backend:  BackendStorageRoute.Backend,
			Apikey:   BackendStorageRoute.Apikey,
			Identity: id,
			Model:    BackendStorageRoute.Model,
			SystemID: BackendStorageRoute.SystemID,
			Username: BackendStorageRoute.Username,
		}
		_, err = dao.GetStorage().UpdateModel(route, jsonModel)
		if err != nil {
			return "", err
		}
		log.Infof("backend updated: %s", id)

	} else {
		id, err = dao.GetStorage().CreateModel(BackendStorageRoute, jsonModel)
		if err != nil {
			return "", err
		}
		log.Infof("backend created: %s", id)
	}
	return id, nil
}

//DeleteBackend deleting the backend from the storage, no data will be deleted
func DeleteBackend(backendname string) error {
	query := fmt.Sprintf("{\"backendname\": \"%s\"}", backendname)
	count, bemodels, err := dao.GetStorage().QueryModel(BackendStorageRoute, query, 0, 10)
	if err != nil {
		log.Alertf("%v", err)
		return err
	}
	if count > 0 {
		bemodel := model.JSONMap(bemodels[0])
		id := bemodel["_id"].(primitive.ObjectID).Hex()
		log.Infof("found backend with id: %s", id)
		route := model.Route{
			Backend:  BackendStorageRoute.Backend,
			Apikey:   BackendStorageRoute.Apikey,
			Identity: id,
			Model:    BackendStorageRoute.Model,
			SystemID: BackendStorageRoute.SystemID,
			Username: BackendStorageRoute.Username,
		}
		err = dao.GetStorage().DeleteModel(route)
		if err != nil {
			return err
		}
		log.Infof("backend deleted: %s", id)
	}
	return nil
}
