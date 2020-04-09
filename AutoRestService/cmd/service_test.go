package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	api "github.com/willie68/AutoRestIoT/api"
)

const baseURL = "http://127.0.0.1:9080"
const baseSslURL = "https://127.0.0.1:9443"
const restEndpoints = "/api/v1/"
const username = "admin"
const passwd = "admin"
const systemID = "autorest-srv"

var AdminUser = BasicUser{name: "admin", pwd: "admin"}
var EditorUser = BasicUser{name: "editor", pwd: "editor"}
var GuestUser = BasicUser{name: "guest", pwd: "guest"}

type BasicUser struct {
	name string
	pwd  string
}

func getClient() http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return http.Client{Transport: tr}
}

func getTestApikey() string {
	value := fmt.Sprintf("%s_%s", servicename, systemID)
	apikey := fmt.Sprintf("%x", md5.Sum([]byte(value)))
	return strings.ToLower(apikey)
}

func getGetRequest(url string, user BasicUser) (*http.Response, error) {
	client := getClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(user.name, user.pwd)
	req.Header.Add(api.APIKeyHeader, getTestApikey())
	req.Header.Add(api.SystemHeader, systemID)
	return client.Do(req)
}

func getPostRequest(url string, user BasicUser, payload []byte) (*http.Response, error) {
	client := getClient()
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(user.name, user.pwd)
	req.Header.Add(api.APIKeyHeader, getTestApikey())
	req.Header.Add(api.SystemHeader, systemID)
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func getDeleteRequest(url string, user BasicUser) (*http.Response, error) {
	client := getClient()
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(user.name, user.pwd)
	req.Header.Add(api.APIKeyHeader, getTestApikey())
	req.Header.Add(api.SystemHeader, systemID)
	return client.Do(req)
}
