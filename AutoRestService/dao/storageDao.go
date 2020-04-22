package dao

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"

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

	CheckUser(username string, password string) bool
	GetUser(username string) (model.User, bool)
	UserInRoles(username string, roles []string) bool

	GetUsers() ([]model.User, error)
	AddUser(user model.User) error
	DeleteUser(username string) error
	ChangePWD(username string, newpassword string, oldpassword string) error

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

//BuildPasswordHash building a hash value of the password
func BuildPasswordHash(password string) string {
	if !strings.HasPrefix(password, "md5:") {
		hash := md5.Sum([]byte(password))
		password = fmt.Sprintf("md5:%x", hash)
	}
	return password
}
