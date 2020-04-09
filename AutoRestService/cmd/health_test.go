package main

import (
	"net/http"
	"testing"
)

func TestReadinessCheck(t *testing.T) {
	client := getClient()
	resp, err := client.Get(baseURL + "/health/readiness")
	if err != nil {
		t.Errorf("Error getting response. %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Wrong statuscode: %d", resp.StatusCode)
		return
	}
	t.Log("Readinesscheck OK.")
}

func TestHealthCheck(t *testing.T) {
	client := getClient()
	resp, err := client.Get(baseURL + "/health/health")
	if err != nil {
		t.Errorf("Error getting response. %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Wrong statuscode: %d", resp.StatusCode)
		return
	}
	t.Log("Healthcheck OK.")
}

func TestSSLReadinessCheck(t *testing.T) {
	client := getClient()
	resp, err := client.Get(baseSslURL + "/health/readiness")
	if err != nil {
		t.Errorf("Error getting response. %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Wrong statuscode: %d", resp.StatusCode)
		return
	}
	t.Log("SSL Readinesscheck OK.")
}

func TestSSLHealthCheck(t *testing.T) {
	client := getClient()
	resp, err := client.Get(baseSslURL + "/health/health")
	if err != nil {
		t.Errorf("Error getting response. %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Wrong statuscode: %d", resp.StatusCode)
		return
	}
	t.Log("SSL Healthcheck OK.")
}
