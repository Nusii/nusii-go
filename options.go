package nusii

import (
	"context"
	"net/http"
	"time"
)

// An Option configures the client when passed to New, or a single request
// when passed as the last argument of any method:
//
//	n, err := nusii.New(nusii.WithMaxRetries(5))
//	page, err := n.Proposals.List(ctx, nil, nusii.WithTimeout(5*time.Second))
type Option func(*config)

type config struct {
	apiKey      string
	accessToken string
	baseURL     string
	timeout     time.Duration
	maxRetries  int
	headers     http.Header
	httpClient  *http.Client
	sleep       func(context.Context, time.Duration) error
}

// with returns a copy of the configuration with the request options applied.
func (c config) with(opts []Option) config {
	if len(opts) == 0 {
		return c
	}
	baseURL := c.baseURL
	c.headers = c.headers.Clone()
	for _, opt := range opts {
		opt(&c)
	}
	if c.baseURL == "" {
		c.baseURL = baseURL
	}
	c.baseURL = normalizeBaseURL(c.baseURL)
	return c
}

func (c config) authorization() string {
	if c.apiKey != "" {
		return "Token token=" + c.apiKey
	}
	return "Bearer " + c.accessToken
}

// WithAPIKey authenticates with a personal API token from
// https://app.nusii.com/settings/api. Defaults to the NUSII_API_KEY
// environment variable.
func WithAPIKey(key string) Option {
	return func(c *config) {
		c.apiKey = key
		c.accessToken = ""
	}
}

// WithAccessToken authenticates with an OAuth 2.0 access token instead of an
// API key. Defaults to the NUSII_ACCESS_TOKEN environment variable.
func WithAccessToken(token string) Option {
	return func(c *config) {
		c.accessToken = token
		c.apiKey = ""
	}
}

// WithBaseURL sets the URL of the Nusii app, e.g. http://localhost:3000.
// Defaults to the NUSII_BASE_URL environment variable, then https://app.nusii.com.
func WithBaseURL(url string) Option {
	return func(c *config) {
		c.baseURL = url
	}
}

// WithTimeout sets how long one attempt of a request may take, reading the
// response included. Zero means no timeout. Default: 30 seconds.
func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.timeout = timeout
	}
}

// WithMaxRetries sets how often a failed request is retried. Rate limited
// requests (429) are always safe to retry; connection errors and 5xx
// responses are only retried for GET requests. Default: 2.
func WithMaxRetries(retries int) Option {
	return func(c *config) {
		c.maxRetries = max(retries, 0)
	}
}

// WithHeader adds a header to every request (on New) or to one request.
func WithHeader(name, value string) Option {
	return func(c *config) {
		c.headers.Set(name, value)
	}
}

// WithHTTPClient sets the HTTP client used to send requests, e.g. for a proxy
// or tests. Default: http.DefaultClient.
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		if client != nil {
			c.httpClient = client
		}
	}
}
