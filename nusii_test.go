package nusii

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewCredentials(t *testing.T) {
	t.Setenv("NUSII_API_KEY", "")
	t.Setenv("NUSII_ACCESS_TOKEN", "")
	t.Setenv("NUSII_BASE_URL", "")

	if _, err := New(); !errors.Is(err, ErrMissingCredentials) {
		t.Fatalf("got %v, want ErrMissingCredentials", err)
	}

	cases := []struct {
		name string
		env  map[string]string
		opts []Option
		want string
	}{
		{"api key option", nil, []Option{WithAPIKey("key")}, "Token token=key"},
		{"access token option", nil, []Option{WithAccessToken("token")}, "Bearer token"},
		{"option over environment", map[string]string{"NUSII_API_KEY": "env-key"}, []Option{WithAccessToken("token")}, "Bearer token"},
		{"api key from environment", map[string]string{"NUSII_API_KEY": " env-key \n"}, nil, "Token token=env-key"},
		{"access token from environment", map[string]string{"NUSII_ACCESS_TOKEN": "env-token"}, nil, "Bearer env-token"},
		{"api key wins in environment", map[string]string{"NUSII_API_KEY": "env-key", "NUSII_ACCESS_TOKEN": "env-token"}, nil, "Token token=env-key"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for name, value := range c.env {
				t.Setenv(name, value)
			}
			n, err := New(c.opts...)
			if err != nil {
				t.Fatal(err)
			}
			if got := n.cfg.authorization(); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestNewBaseURL(t *testing.T) {
	t.Setenv("NUSII_BASE_URL", "")
	cases := []struct {
		env, option, want string
	}{
		{"", "", "https://app.nusii.com"},
		{"http://localhost:3000/", "", "http://localhost:3000"},
		{"", "https://example.com/api/v2/", "https://example.com"},
		{"http://env.example", "https://option.example", "https://option.example"},
	}
	for _, c := range cases {
		t.Setenv("NUSII_BASE_URL", c.env)
		opts := []Option{WithAPIKey("key")}
		if c.option != "" {
			opts = append(opts, WithBaseURL(c.option))
		}
		n, err := New(opts...)
		if err != nil {
			t.Fatal(err)
		}
		if n.BaseURL() != c.want {
			t.Errorf("env %q option %q: got %q, want %q", c.env, c.option, n.BaseURL(), c.want)
		}
	}
}

func TestRequestHeaders(t *testing.T) {
	n, api := newMockClient(t, []http.HandlerFunc{ok(map[string]any{}), ok(map[string]any{})},
		WithHeader("X-Request-Source", "tests"),
		WithHeader("Accept", "application/vnd.api+json"),
	)
	ctx := context.Background()

	if err := n.Request(ctx, RequestConfig{Path: "/clients"}, nil); err != nil {
		t.Fatal(err)
	}
	header := api.request(0).Header
	want := map[string]string{
		"Authorization":    "Token token=test-key",
		"Content-Type":     "application/json",
		"Accept":           "application/vnd.api+json",
		"User-Agent":       "nusii-go/" + Version,
		"X-Request-Source": "tests",
	}
	for name, value := range want {
		if got := header.Get(name); got != value {
			t.Errorf("%s: got %q, want %q", name, got, value)
		}
	}

	// Per request options apply to that request only.
	if err := n.Request(ctx, RequestConfig{Path: "/clients"}, nil, WithHeader("X-Request-Source", "once"), WithAccessToken("token")); err != nil {
		t.Fatal(err)
	}
	header = api.request(1).Header
	if header.Get("X-Request-Source") != "once" || header.Get("Authorization") != "Bearer token" {
		t.Errorf("per request options not applied: %v", header)
	}
	if n.cfg.headers.Get("X-Request-Source") != "tests" || n.cfg.apiKey != "test-key" {
		t.Error("per request options changed the client")
	}
}

func TestRequestCannotOverrideAuthorization(t *testing.T) {
	n, api := newMockClient(t, []http.HandlerFunc{ok(map[string]any{})}, WithHeader("Authorization", "Bearer other"))
	if err := n.Request(context.Background(), RequestConfig{Path: "/clients"}, nil); err != nil {
		t.Fatal(err)
	}
	if got := api.request(0).Header.Get("Authorization"); got != "Token token=test-key" {
		t.Errorf("got %q", got)
	}
}

func TestRequest(t *testing.T) {
	n, api := newMockClient(t, []http.HandlerFunc{ok(map[string]any{"data": []any{}})})

	var out map[string]any
	err := n.Request(context.Background(), RequestConfig{
		Method: "post",
		Path:   "/proposals",
		Query:  url.Values{"statuses[]": {"draft", "pending"}},
		Body:   map[string]any{"proposal": map[string]any{"title": "Hi"}},
	}, &out)
	if err != nil {
		t.Fatal(err)
	}

	r := api.request(0)
	if r.Method != http.MethodPost || r.Path != "/api/v2/proposals" {
		t.Errorf("got %s %s", r.Method, r.Path)
	}
	if !reflect.DeepEqual(r.Query["statuses[]"], []string{"draft", "pending"}) {
		t.Errorf("got query %v", r.Query)
	}
	if bodyOf(t, r, "proposal")["title"] != "Hi" {
		t.Errorf("got body %v", r.Body)
	}
	if _, ok := out["data"]; !ok {
		t.Errorf("response not decoded: %v", out)
	}
}

func TestEmptyResponseBody(t *testing.T) {
	n, _ := newMockClient(t, []http.HandlerFunc{func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }})
	var out map[string]any
	if err := n.Request(context.Background(), RequestConfig{Method: http.MethodDelete, Path: "/x"}, &out); err != nil {
		t.Fatal(err)
	}
	if out != nil {
		t.Errorf("got %v", out)
	}
}

func TestNonJSONResponse(t *testing.T) {
	n, _ := newMockClient(t, []http.HandlerFunc{func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html>Log in</html>"))
	}})

	_, err := n.Account.Me(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !strings.Contains(apiErr.Message, "received text/html. Check the base URL") {
		t.Fatalf("got %v", err)
	}
}

func TestRetriesRateLimitedRequests(t *testing.T) {
	limited := reply(http.StatusTooManyRequests, map[string]any{"error": "Too many requests"}, "Retry-After", "3")
	n, api := newMockClient(t, []http.HandlerFunc{limited, ok(one(res("clients", 1, nil)))})

	// Even a POST: a throttled request never reached the application.
	client, err := n.Clients.Create(context.Background(), ClientCreateParams{Name: "Jane", Email: "jane@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if client.ID != 1 || api.requestCount() != 2 {
		t.Errorf("got client %d after %d requests", client.ID, api.requestCount())
	}
	if !reflect.DeepEqual(api.sleeps, []time.Duration{3 * time.Second}) {
		t.Errorf("got sleeps %v", api.sleeps)
	}
}

func TestDoesNotWaitForLongRetryAfter(t *testing.T) {
	limited := reply(http.StatusTooManyRequests, map[string]any{"error": "Too many requests"}, "Retry-After", "120")
	n, api := newMockClient(t, []http.HandlerFunc{limited})

	_, err := n.Account.Me(context.Background())
	var apiErr *APIError
	if !errors.Is(err, ErrRateLimit) || !errors.As(err, &apiErr) {
		t.Fatalf("got %v", err)
	}
	if wait, ok := apiErr.RetryAfter(); !ok || wait != 120*time.Second {
		t.Errorf("got retry after %v %v", wait, ok)
	}
	if api.requestCount() != 1 || len(api.sleeps) != 0 {
		t.Errorf("got %d requests and sleeps %v", api.requestCount(), api.sleeps)
	}
}

func TestRetriesServerErrorsOnlyForReads(t *testing.T) {
	failing := reply(http.StatusBadGateway, nil)

	n, api := newMockClient(t, []http.HandlerFunc{failing, failing, ok(one(res("accounts", 1, nil)))})
	if _, err := n.Account.Me(context.Background()); err != nil {
		t.Fatal(err)
	}
	if api.requestCount() != 3 || len(api.sleeps) != 2 {
		t.Errorf("got %d requests and sleeps %v", api.requestCount(), api.sleeps)
	}
	if api.sleeps[0] > 500*time.Millisecond || api.sleeps[0] < 375*time.Millisecond || api.sleeps[1] > time.Second {
		t.Errorf("unexpected backoff %v", api.sleeps)
	}

	n, api = newMockClient(t, []http.HandlerFunc{failing})
	_, err := n.Proposals.Send(context.Background(), 1, ProposalSendParams{Email: String("jane@example.com")})
	if !errors.Is(err, ErrServer) || api.requestCount() != 1 {
		t.Errorf("got %v after %d requests", err, api.requestCount())
	}
}

func TestStopsRetryingAfterMaxRetries(t *testing.T) {
	failing := reply(http.StatusServiceUnavailable, nil)
	n, api := newMockClient(t, []http.HandlerFunc{failing, failing}, WithMaxRetries(1))

	_, err := n.Account.Me(context.Background())
	if !errors.Is(err, ErrServiceUnavailable) || !errors.Is(err, ErrServer) {
		t.Fatalf("got %v", err)
	}
	if api.requestCount() != 2 {
		t.Errorf("got %d requests", api.requestCount())
	}

	n, api = newMockClient(t, []http.HandlerFunc{failing})
	if _, err := n.Account.Me(context.Background(), WithMaxRetries(0)); !errors.Is(err, ErrServer) || api.requestCount() != 1 {
		t.Errorf("got %v after %d requests", err, api.requestCount())
	}
}

func TestConnectionErrors(t *testing.T) {
	n, err := New(WithAPIKey("key"), WithBaseURL("http://127.0.0.1:1"), WithMaxRetries(2))
	if err != nil {
		t.Fatal(err)
	}
	var sleeps int
	n.cfg.sleep = func(ctx context.Context, d time.Duration) error { sleeps++; return nil }

	_, err = n.Account.Me(context.Background())
	var connErr *ConnectionError
	if !errors.Is(err, ErrConnection) || errors.Is(err, ErrTimeout) || !errors.As(err, &connErr) {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(err.Error(), "could not reach the Nusii API at http://127.0.0.1:1/api/v2/account/me") {
		t.Errorf("got message %q", err)
	}
	if sleeps != 2 {
		t.Errorf("GET retried %d times, want 2", sleeps)
	}

	sleeps = 0
	_, err = n.Clients.Create(context.Background(), ClientCreateParams{Name: "Jane", Email: "jane@example.com"})
	if !errors.Is(err, ErrConnection) || sleeps != 0 {
		t.Errorf("POST: got %v after %d retries", err, sleeps)
	}
}

func TestTimeout(t *testing.T) {
	slow := func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
		}
	}
	n, api := newMockClient(t, []http.HandlerFunc{slow, slow}, WithTimeout(20*time.Millisecond), WithMaxRetries(1))

	_, err := n.Account.Me(context.Background())
	if !errors.Is(err, ErrTimeout) || !errors.Is(err, ErrConnection) {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(err.Error(), "timed out after 20ms") {
		t.Errorf("got message %q", err)
	}
	if api.requestCount() != 2 {
		t.Errorf("got %d requests", api.requestCount())
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	slow := func(w http.ResponseWriter, r *http.Request) {
		cancel()
		<-r.Context().Done()
	}
	n, api := newMockClient(t, []http.HandlerFunc{slow})

	_, err := n.Account.Me(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	if api.requestCount() != 1 {
		t.Errorf("retried a canceled request")
	}
}

func TestWithHTTPClient(t *testing.T) {
	var used bool
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		used = true
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`[{"id":"clean","name":"Modern Theme"}]`)),
		}, nil
	})
	n, err := New(WithAPIKey("key"), WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	themes, err := n.Themes.List(context.Background())
	if err != nil || !used || len(themes) != 1 || themes[0].ID != "clean" {
		t.Errorf("got %+v %v (transport used: %v)", themes, err, used)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDecodeErrorsNameTheRecord(t *testing.T) {
	n, _ := newMockClient(t, []http.HandlerFunc{
		ok(one(res("clients", 1, map[string]any{"created_at": "yesterday"}))),
		ok(map[string]any{"meta": map[string]any{}}),
	})
	ctx := context.Background()

	if _, err := n.Clients.Get(ctx, 1); err == nil || !strings.Contains(err.Error(), "decoding a clients record") {
		t.Errorf("got %v", err)
	}
	if _, err := n.Clients.Get(ctx, 1); err == nil || !strings.Contains(err.Error(), "no data") {
		t.Errorf("got %v", err)
	}
}

func TestSleepStopsWithTheContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sleep(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("got %v", err)
	}
	if err := sleep(context.Background(), time.Millisecond); err != nil {
		t.Errorf("got %v", err)
	}
}
