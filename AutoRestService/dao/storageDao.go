package dao

import (
	"io"

	"github.com/willie68/AutoRestIoT/model"
)

//FulltextIndexName name of the index containing fulltext data
const FulltextIndexName = "$fulltext"

/*
StorageDao this is the interface which all method implementation of a storage engine has to fulfill
*/
type StorageDao interface {
	AddFile(backend string, filename string, reader io.Reader) (string, error)
	GetFilename(backend string, fileid string) (string, error)
	GetFile(backend string, fileid string, stream io.Writer) error
	DeleteFile(backend string, fileid string) error

	CreateModel(route model.Route, data model.JSONMap) (string, error)
	CreateModels(route model.Route, datas []model.JSONMap) ([]string, error)
	CountModel(route model.Route) (int, error)
	GetModel(route model.Route) (model.JSONMap, error)
	QueryModel(route model.Route, query string, offset int, limit int) (int, []model.JSONMap, error)
	UpdateModel(route model.Route, data model.JSONMap) (model.JSONMap, error)
	DeleteModel(route model.Route) error

	GetIndexNames(route model.Route) ([]string, error)
	DeleteIndex(route model.Route, name string) error
	UpdateIndex(route model.Route, index model.Index) error

	GetUsers() ([]model.User, error)
	GetUser(username string) (model.User, bool)
	AddUser(user model.User) (model.User, error)
	DeleteUser(username string) error
	ChangePWD(username string, newpassword string) (model.User, error)

	DeleteBackend(beackend string) error
	DropAll()
	Ping() error

	Stop()
}

var storageDao StorageDao

//GetStorage getting the actual storage dao
func GetStorage() StorageDao {
	return storageDao
}

//SetStorage setting the actual storage dad
func SetStorage(storage StorageDao) {
	storageDao = storage
}
