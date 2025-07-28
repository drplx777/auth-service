package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

var dbServiceURL = func() string {
	if envURL := os.Getenv("DB_SERVICE_URL"); envURL != "" {
		return envURL
	}
	return "http://db-service:8000"
}()

func Post(path string, payload interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return http.Post(dbServiceURL+path, "application/json", bytes.NewBuffer(jsonData))
}

func Get(path string) (*http.Response, error) {
	return http.Get(dbServiceURL + path)
}
