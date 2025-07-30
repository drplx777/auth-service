package api_clients

import (
	"auth-service/internal/dtos"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

type DbServiceClient struct {
	baseUrl string
}

func NewDbServiceClient() *DbServiceClient {
	baseUrl := os.Getenv("DB_SERVICE_URL")
	if baseUrl == "" {
		baseUrl = "http://db-service:8000"
	}

	return &DbServiceClient{
		baseUrl: baseUrl,
	}
}

func (c *DbServiceClient) Post(path string, payload any) (*http.Response, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return http.Post(c.baseUrl+path, "application/json", bytes.NewBuffer(jsonData))
}

func (c *DbServiceClient) RegisterUser(dto *dtos.RegisterRequest) (*dtos.RegisterInnerApiResponse, error) {
	resp, err := c.Post("/user/register", dto)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}
	defer func() { _ = resp.Body.Close() }()

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read register user response body")
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.Wrap(errors.New(string(bodyData)), "failed to register user")
	}

	var output dtos.RegisterInnerApiResponse

	if err = json.Unmarshal(bodyData, &output); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal register user response body")
	}

	return &output, nil
}
