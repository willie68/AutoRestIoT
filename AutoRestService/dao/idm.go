package dao

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

/*
  This is the identity management system for the AutoRest service. Here you will find all methods regarding the identity of a user,
  authentication and authorisation.
*/

const hashPrefix = "hash:"

//BuildPasswordHash building a hash value of the password
func BuildPasswordHash(password string, salt []byte) string {
	if !strings.HasPrefix(password, hashPrefix) {
		hash := pbkdf2.Key([]byte(password), salt, 4096, 32, sha1.New)
		// hash := md5.Sum([]byte(password))
		password = fmt.Sprintf("%s:%x", hashPrefix, hash)
	}
	return password
}
