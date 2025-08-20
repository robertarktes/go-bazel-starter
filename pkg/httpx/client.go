package httpx

import (
	"context"
	"net/http"
	"time"

	"github.com/robertarktes/go-bazel-starter/pkg/retry"
)

type Client struct {
	httpClient   *http.Client
	timeout      time.Duration
	maxRetries   int
	backoffFunc  func(attempt int) time.Duration
	requestHook  func(*http.Request)
	responseHook func(*http.Response, time.Duration)
}

type Option func(*Client)

func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

func WithRetries(max int, backoff func(attempt int) time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = max
		c.backoffFunc = backoff
	}
}

func WithRequestHook(hook func(*http.Request)) Option {
	return func(c *Client) { c.requestHook = hook }
}

func WithResponseHook(hook func(*http.Response, time.Duration)) Option {
	return func(c *Client) { c.responseHook = hook }
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{},
		timeout:    10 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	var resp *http.Response
	var latency time.Duration
	err := retry.Retry(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}
		if c.requestHook != nil {
			c.requestHook(req)
		}
		start := time.Now()
		r, err := c.httpClient.Do(req)
		latency = time.Since(start)
		if err != nil {
			return err
		}
		if c.responseHook != nil {
			c.responseHook(r, latency)
		}
		resp = r
		return nil
	}, retry.WithMaxAttempts(c.maxRetries+1), retry.WithBackoff(c.backoffFunc))
	return resp, err
}
