package http

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Request struct {
	Method  string
	URL     string
	Body    string
	Headers map[string][]string
}

type Response struct {
	StatusCode    int
	Headers       map[string][]string
	ContentLength int64
	Body          string
}

type client struct {
	client *http.Client
}

func NewClient() *client {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &client{client: httpClient}
}

func (c *client) DoRequest(req Request) (*Response, error) {
	body := strings.NewReader(req.Body)
	// Create a new http.Request with the given method, url, body and headers
	request, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return nil, err
	}
	for key, values := range req.Headers {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	// Send the request and get the response
	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	// Read and close the response body
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respHeaders := make(map[string][]string)
	for key, value := range resp.Header {
		respHeaders[key] = value
	}

	response := Response{
		StatusCode:    resp.StatusCode,
		Headers:       respHeaders,
		ContentLength: resp.ContentLength,
		Body:          string(data),
	}

	// Return the information from the function
	return &response, nil
}
