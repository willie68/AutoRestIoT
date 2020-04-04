package dao

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"

	"github.com/willie68/AutoRestIoT/model"
)

/*
StorageDao this is the interface which all method implementation of a storage engine has to fulfill
*/
type StorageDao interface {
	AddFile(filename string, reader io.Reader) (string, error)
	GetFilename(fileid string) (string, error)
	GetFile(fileid string, stream io.Writer) error

	CheckUser(username string, password string) bool
	GetUser(username string) (model.User, bool)

	AddUser(user model.User) error
	DeleteUser(username string) error
	ChangePWD(username string, newpassword string, oldpassword string) error

	DropAll()
	Ping() error

	Stop()
}

var Storage StorageDao

func GetStorage() StorageDao {
	return Storage
}

func BuildPasswordHash(password string) string {
	if !strings.HasPrefix(password, "md5:") {
		hash := md5.Sum([]byte(password))
		password = fmt.Sprintf("md5:%x", hash)
	}
	return password
}
