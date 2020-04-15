package main

import (
	"io"
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	pingPath     = "/ping"
	uploadPath   = "/upload"
	downloadPath = "/download"
)

type StorageClient struct {
	Name     string
	Endpoint string
	Username string
	Password string
}

func newStorageClient(name, endpoint, username, password string) *StorageClient {
	return &StorageClient{name, endpoint, username, password}
}

func (c *StorageClient) ping() (*http.Response, error) {
	client := resty.New()

	resp, err := client.R().SetBasicAuth(c.Username, c.Password).Get(c.Endpoint + pingPath)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}

func (c *StorageClient) upload(file io.Reader, filename string) (*http.Response, error) {
	client := resty.New()

	resp, err := client.R().
		SetFileReader("file", filename, file).
		SetFormData(map[string]string{
			"filename": filename,
		}).
		SetBasicAuth(c.Username, c.Password).
		Post(c.Endpoint + uploadPath)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}

func (c *StorageClient) download(filename string) (*http.Response, error) {
	client := resty.New()

	resp, err := client.R().
		SetQueryParam(filename, filename).
		SetBasicAuth(c.Username, c.Password).
		Get(c.Endpoint + pingPath)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}
