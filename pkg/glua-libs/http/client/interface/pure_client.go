package http_interface

import (
	"net/http"
	"net/url"
)

func NewPureClient() *PureClient {
	return &PureClient{Client: &http.Client{}}
}

type PureClient struct {
	*http.Client
}

func (c *PureClient) DoRequest(req *http.Request) (*http.Response, error) {
	return c.Do(req)
}

func (c *PureClient) PostFormRequest(url string, data url.Values) (*http.Response, error) {
	return c.PostForm(url, data)
}
