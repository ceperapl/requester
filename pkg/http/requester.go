package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	reqTimeout = 10 * time.Second

	Err
)

type Requester interface {
	DoRequest(req Request) (*Response, error)
}

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

type Client struct {
	Client *http.Client
}

func NewClient() *Client {
	httpClient := &http.Client{
		Timeout: reqTimeout,
	}

	return &Client{Client: httpClient}
}

func (c *Client) DoRequest(ctx context.Context, req Request) (*Response, error) {
	body := strings.NewReader(req.Body)
	// Create a new http.Request with the given method, url, body and headers
	request, err := http.NewRequestWithContext(ctx, req.Method, req.URL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	for key, values := range req.Headers {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	// Send the request and get the response
	resp, err := c.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	// Read and close the response body
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read from http response body: %w", err)
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
