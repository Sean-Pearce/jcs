package client

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

// StorageClient is a client of storage service.
type StorageClient struct {
	Name     string
	Endpoint string
	Username string
	Password string
}

// NewStorageClient constructs a new storage client.
func NewStorageClient(name, endpoint, username, password string) *StorageClient {
	return &StorageClient{name, endpoint, username, password}
}

// Ping pings storage server with basic auth.
func (c *StorageClient) Ping() (*http.Response, error) {
	client := resty.New()

	resp, err := client.R().SetBasicAuth(c.Username, c.Password).Get(c.Endpoint + pingPath)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}

// Upload uploads a file to storage server using given io.Reader.
func (c *StorageClient) Upload(file io.Reader, filename string) (*http.Response, error) {
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

// Download downloads given filename from storage server.
func (c *StorageClient) Download(filename string) (*http.Response, error) {
	client := resty.New()

	resp, err := client.R().
		SetQueryParam("filename", filename).
		SetBasicAuth(c.Username, c.Password).
		SetDoNotParseResponse(true).
		Get(c.Endpoint + downloadPath)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}
