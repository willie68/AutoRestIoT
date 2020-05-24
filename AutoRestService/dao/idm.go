package dao

/*
  This is the identity management system for the AutoRest service. Here you will find all methods regarding the identity of a user,
  authentication and authorisation.
*/

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/willie68/AutoRestIoT/internal/slicesutils"
	"github.com/willie68/AutoRestIoT/model"
	"golang.org/x/crypto/pbkdf2"
)

//IDM the idm cache struct with users and salts
type IDM struct {
	users map[string]string
	salts map[string][]byte
}

const hashPrefix = "hash:"

// time to reload all users
const userReloadPeriod = 1 * time.Hour

var idm IDM

//NewIDM creating a new idm object
func NewIDM() IDM {
	return IDM{
		users: make(map[string]string),
		salts: make(map[string][]byte),
	}
}

//GetIDM getting the actual idm object
func GetIDM() IDM {
	return idm
}

//SetIDM setting the actual idm object
func SetIDM(newIdm IDM) {
	idm = newIdm
}

//BuildPasswordHash building a hash value of the password
func BuildPasswordHash(password string, salt []byte) string {
	if !strings.HasPrefix(password, hashPrefix) {
		hash := pbkdf2.Key([]byte(password), salt, 4096, 32, sha1.New)
		// hash := md5.Sum([]byte(password))
		password = fmt.Sprintf("%s:%x", hashPrefix, hash)
	}
	return password
}

//InitIDM initialise the local idm system
func (i *IDM) InitIDM() {
	i.initUsers()
}

func (i *IDM) initUsers() {
	i.reloadUsers()

	go func() {
		background := time.NewTicker(userReloadPeriod)
		for range background.C {
			i.reloadUsers()
		}
	}()
}

//GetSalt getting the salt for a user.
func (i *IDM) GetSalt(username string) ([]byte, bool) {
	username = strings.ToLower(username)
	salt, ok := i.salts[username]
	if ok {
		return salt, true
	}

	return []byte{}, false
}

func (i *IDM) reloadUsers() {
	users, err := GetStorage().GetUsers()
	if err != nil {
		log.Alertf("error loading users: %v", err)
		return
	}
	localUsers := make(map[string]string)
	localSalts := make(map[string][]byte)
	for _, user := range users {
		username := user.Name
		password := user.Password
		salt := user.Salt
		localSalts[username] = salt
		localUsers[username] = BuildPasswordHash(password, salt)
	}
	i.users = localUsers
	i.salts = localSalts
	if len(i.users) == 0 {
		admin := model.User{
			Name:      "admin",
			Firstname: "",
			Lastname:  "Admin",
			Password:  "admin",
			Admin:     true,
			Roles:     []string{"admin"},
		}
		user, err := GetStorage().AddUser(admin)
		if err != nil {
			log.Alertf("error adding user: %v", err)
			return
		}
		i.addUserToMap(user)

		editor := model.User{
			Name:      "editor",
			Firstname: "",
			Lastname:  "Editor",
			Password:  "editor",
			Admin:     false,
			Guest:     false,
			Roles:     []string{"edit"},
		}
		user, err = GetStorage().AddUser(editor)
		if err != nil {
			log.Alertf("error adding user: %v", err)
			return
		}
		i.addUserToMap(user)

		guest := model.User{
			Name:      "guest",
			Firstname: "",
			Lastname:  "Guest",
			Password:  "guest",
			Admin:     false,
			Guest:     true,
			Roles:     []string{"read"},
		}
		user, err = GetStorage().AddUser(guest)
		if err != nil {
			log.Alertf("error adding user: %v", err)
			return
		}
		i.addUserToMap(user)
	}
}

//CheckUser checking username and password... returns true if the user is active and the password for this user is correct
func (i *IDM) CheckUser(username string, password string) bool {
	username = strings.ToLower(username)
	pwd, ok := i.users[username]
	if ok {
		if pwd == password {
			return true
		}
		user, ok := GetStorage().GetUser(username)
		if ok {
			if user.Password == password {
				//change password on another node
				i.addUserToMap(user)
				return true
			}
		}
	}

	if !ok {
		user, ok := GetStorage().GetUser(username)
		if ok {
			if user.Password == password {
				//change password on another node
				i.addUserToMap(user)
				return true
			}
		}
	}

	return false
}

//UserInRoles is a user in the given role
func (i *IDM) UserInRoles(username string, roles []string) bool {
	user, ok := GetStorage().GetUser(username)
	if !ok {
		return false
	}

	for _, role := range roles {
		if slicesutils.Contains(user.Roles, role) {
			return true
		}
	}
	return false
}

// AddUser adding a new user to the system
func (i *IDM) AddUser(user model.User) (model.User, error) {
	if user.Name == "" {
		return model.User{}, errors.New("username should not be empty")
	}
	user.Name = strings.ToLower(user.Name)
	_, ok := GetStorage().GetUser(user.Name)
	if ok {
		return model.User{}, errors.New("username already exists")
	}

	user, err := GetStorage().AddUser(user)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

//DeleteUser deletes a user from the idm
func (i *IDM) DeleteUser(username string) error {
	err := GetStorage().DeleteUser(username)
	if err != nil {
		return err
	}
	delete(i.users, username)
	delete(i.salts, username)
	return nil
}

// ChangePWD changes the apssword of a single user
func (i *IDM) ChangePWD(username string, newpassword string, oldpassword string) error {
	if username == "" {
		return errors.New("username should not be empty")
	}
	username = strings.ToLower(username)
	pwd, ok := i.users[username]
	if !ok {
		return errors.New("username not registered")
	}
	usermodel, ok := GetStorage().GetUser(username)
	if !ok {
		return errors.New("username not registered")
	}

	oldpassword = BuildPasswordHash(oldpassword, usermodel.Salt)
	if pwd != oldpassword {
		return errors.New("actual password incorrect")
	}

	user, err := GetStorage().ChangePWD(username, newpassword)
	if err != nil {
		return err
	}

	i.addUserToMap(user)
	return nil
}

func (i *IDM) addUserToMap(user model.User) {
	i.users[user.Name] = user.Password
	i.salts[user.Name] = user.Salt
}
