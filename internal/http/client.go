package http

import (
	"net/http"
	"time"
)

type Client struct {
	client  *http.Client
	retries int
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		retries: 3,
	}
}

func (c *Client) DoRequest(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i <= c.retries; i++ {
		resp, err = c.client.Do(req)
		if err == nil {
			return resp, nil
		}
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return nil, err
}
