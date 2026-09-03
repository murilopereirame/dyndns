package client

import (
	"fmt"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	MakeUrl(endpoint string) string
}

type Client struct {
	BaseUrl string
}

func (c Client) MakeUrl(endpoint string) string {
	if endpoint == "" {
		return c.BaseUrl
	}
	return fmt.Sprintf("%s/%s", c.BaseUrl, endpoint)
}

func (c Client) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}
