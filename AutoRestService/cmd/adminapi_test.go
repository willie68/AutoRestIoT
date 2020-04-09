package main

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetBackend(t *testing.T) {
	url := baseSslURL + restEndpoints + "admin/mybe/"
	resp, err := getGetRequest(url, AdminUser)
	if err != nil {
		t.Errorf("Error getting response. %v", err)
		return
	}
	defer resp.Body.Close()

	bodyText, err := ioutil.ReadAll(resp.Body)
	s := string(bodyText)
	t.Logf("response: url: %s, response: %s", url, s)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Wrong statuscode: %d", resp.StatusCode)
		return
	}

	t.Log("Get Backend OK.")
}
