package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/willie68/AutoRestIoT/model"
)

func TestGetUsers(t *testing.T) {
	url := baseSslURL + restEndpoints + "users"
	resp, err := getGetRequest(url, AdminUser)
	if err != nil {
		t.Errorf("Error getting response. %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Wrong statuscode: %d", resp.StatusCode)
		return
	}

	bodyText, err := ioutil.ReadAll(resp.Body)
	var users []model.User
	err = json.Unmarshal(bodyText, &users)
	if err != nil {
		t.Errorf("Error ummarshalling response. %v", err)
		return
	}

	t.Logf("response: url: %s, response: %v", url, users)
	if len(users) != 3 {
		t.Errorf("predefined user count not correct: %d", len(users))
		return
	}

	t.Log("Get users OK.")
}

func TestPostUser(t *testing.T) {
	url := baseSslURL + restEndpoints + "users"

	newUser := model.User{
		Name:     "newuser",
		Password: "newuser",
		Admin:    false,
		Guest:    false,
		Roles:    []string{"admin", "edit", "read"},
	}
	basicUser := BasicUser{name: newUser.Name, pwd: newUser.Password}

	payload, err := json.Marshal(newUser)
	if err != nil {
		t.Errorf("Error marshall new user. %v", err)
		return
	}

	resp, err := getPostRequest(url, AdminUser, payload)
	if err != nil {
		t.Errorf("Error getting response. %v", err)
		return
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	s := string(bodyText)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Wrong statuscode: %d, %s", resp.StatusCode, s)

	var user model.User
	err = json.Unmarshal(bodyText, &user)
	if err != nil {
		t.Errorf("Error ummarshalling response. %v", err)
		return
	}

	t.Logf("response: url: %s, response: %v", url, user)
	assert.Equal(t, newUser.Name, user.Name, "user name not identically")
	assert.Equal(t, "#####", user.Password, "password not identically")
	assert.Equal(t, newUser.Admin, user.Admin, "admin not identically")
	assert.Equal(t, newUser.Guest, user.Guest, "guest not identically")
	assert.Contains(t, user.Roles, "admin", "roles not inserted")
	assert.Contains(t, user.Roles, "edit", "roles not inserted")
	assert.Contains(t, user.Roles, "read", "roles not inserted")

	url = baseSslURL + restEndpoints + "users/" + basicUser.name
	resp, err = getGetRequest(url, basicUser)
	if err != nil {
		t.Errorf("Error getting response. %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Wrong statuscode: %d", resp.StatusCode)
		return
	}

	resp, err = getDeleteRequest(url, AdminUser)
	if err != nil {
		t.Errorf("Error getting response. %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Wrong statuscode: %d", resp.StatusCode)
		return
	}

	t.Log("Get users OK.")
}
