// Package ml is a small HTTP client for the Python prediction service.
package ml

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// StatusError is returned when the ML service responds with a non-2xx status.
// The body is preserved so the API can pass a structured error through.
type StatusError struct {
	Status int
	Body   []byte
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("ml service returned HTTP %d", e.Status)
}

// Client calls the ML service. It never retrains or invents a prediction.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient returns a client with the default timeout.
func NewClient(baseURL string) *Client {
	return NewClientWithTimeout(baseURL, 10*time.Second)
}

// NewClientWithTimeout is used by tests and for slow deployments.
func NewClientWithTimeout(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) BaseURL() string { return c.baseURL }

// Do forwards one request. It returns the raw body and status for 2xx
// responses, and a *StatusError for non-2xx responses.
func (c *Client) Do(ctx context.Context, method, path string, body []byte) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return raw, resp.StatusCode, &StatusError{Status: resp.StatusCode, Body: raw}
	}
	return raw, resp.StatusCode, nil
}

func (c *Client) Health(ctx context.Context) ([]byte, error) {
	raw, _, err := c.Do(ctx, http.MethodGet, "/health", nil)
	return raw, err
}

func (c *Client) Models(ctx context.Context) ([]byte, error) {
	raw, _, err := c.Do(ctx, http.MethodGet, "/models", nil)
	return raw, err
}

func (c *Client) Report(ctx context.Context) ([]byte, error) {
	raw, _, err := c.Do(ctx, http.MethodGet, "/report", nil)
	return raw, err
}

func (c *Client) Predict(ctx context.Context, path string, payload []byte) ([]byte, error) {
	raw, _, err := c.Do(ctx, http.MethodPost, path, payload)
	return raw, err
}
