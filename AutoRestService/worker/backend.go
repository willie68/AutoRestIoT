package worker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/model"
)

func init() {
}

//ValidateBackend validate if a backend definition is valid
func ValidateBackend(be model.Backend) error {
	// checking backendname format
	// checking models
	// checking indexes
	return nil
}

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
		clientID := fmt.Sprintf("autorestIoT.%s", datasource.Name)
		err := mqttRegisterTopic(clientID, backendname, datasource)
		if err != nil {
			return err
		}
	default:
		log.Alertf("type \"%s\" is not availble as data source type", datasource.Type)
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
