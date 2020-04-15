package worker

import (
	"github.com/willie68/AutoRestIoT/dao"
	"github.com/willie68/AutoRestIoT/model"
)

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
			Backend: backend.Backendname,
			Model:   bemodel.Name,
		}
		names, err := dao.GetStorage().GetIndexNames(route)
		if err != nil {
			log.Fatalf("%v", err)
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
				if err != nil {
					log.Fatalf("%v", err)
				}
			}
		}
		for _, index := range indexes {
			err := dao.GetStorage().UpdateIndex(route, index)
			if err != nil {
				log.Fatalf("%v", err)
			}
		}
	}
	return nil
}
