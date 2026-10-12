package nusii

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"
)

// recordedRequest is a request the mock server received.
type recordedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   map[string]any
}

// mockAPI is a Nusii API stand in that answers with the queued handlers, in order.
type mockAPI struct {
	t        *testing.T
	mu       sync.Mutex
	queue    []http.HandlerFunc
	requests []recordedRequest
	sleeps   []time.Duration
	server   *httptest.Server
}

// newMockClient returns a client for a mock API answering with responses, in order.
func newMockClient(t *testing.T, responses []http.HandlerFunc, opts ...Option) (*Nusii, *mockAPI) {
	t.Helper()
	m := &mockAPI{t: t, queue: responses}
	m.server = httptest.NewServer(http.HandlerFunc(m.serve))
	t.Cleanup(m.server.Close)

	options := append([]Option{WithAPIKey("test-key"), WithBaseURL(m.server.URL)}, opts...)
	n, err := New(options...)
	if err != nil {
		t.Fatal(err)
	}
	// Record retry delays instead of waiting.
	n.cfg.sleep = func(ctx context.Context, d time.Duration) error {
		m.mu.Lock()
		m.sleeps = append(m.sleeps, d)
		m.mu.Unlock()
		return ctx.Err()
	}
	return n, m
}

func (m *mockAPI) serve(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	var body map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			m.t.Errorf("request body is not a JSON object: %s", raw)
		}
	}

	m.mu.Lock()
	m.requests = append(m.requests, recordedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.Query(),
		Header: r.Header.Clone(),
		Body:   body,
	})
	if len(m.queue) == 0 {
		m.mu.Unlock()
		m.t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		http.Error(w, "unexpected request", http.StatusTeapot)
		return
	}
	next := m.queue[0]
	m.queue = m.queue[1:]
	m.mu.Unlock()

	next(w, r)
}

// request returns the i-th request the mock received.
func (m *mockAPI) request(i int) recordedRequest {
	m.t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	if i >= len(m.requests) {
		m.t.Fatalf("got %d requests, want at least %d", len(m.requests), i+1)
	}
	return m.requests[i]
}

func (m *mockAPI) requestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

// reply answers with a JSON body.
func reply(status int, body any, headers ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		for i := 0; i+1 < len(headers); i += 2 {
			w.Header().Set(headers[i], headers[i+1])
		}
		w.WriteHeader(status)
		if body != nil {
			json.NewEncoder(w).Encode(body)
		}
	}
}

// ok answers with a 200 JSON body.
func ok(body any) http.HandlerFunc {
	return reply(http.StatusOK, body)
}

// res builds a JSON:API resource object.
func res(kind string, id int, attributes map[string]any) map[string]any {
	if attributes == nil {
		attributes = map[string]any{}
	}
	return map[string]any{"id": strconv.Itoa(id), "type": kind, "attributes": attributes}
}

// one builds a JSON:API document with a single resource.
func one(resource map[string]any) map[string]any {
	return map[string]any{"data": resource}
}

// many builds a JSON:API document with a list of resources and pagination meta.
func many(meta map[string]any, resources ...map[string]any) map[string]any {
	if resources == nil {
		resources = []map[string]any{}
	}
	document := map[string]any{"data": resources}
	if meta != nil {
		document["meta"] = meta
	}
	return document
}

// meta builds pagination meta for a single page, with overrides.
func meta(overrides map[string]any) map[string]any {
	m := map[string]any{"current_page": 1, "next_page": nil, "prev_page": nil, "total_pages": 1, "total_count": 1}
	for k, v := range overrides {
		m[k] = v
	}
	return m
}

// bodyOf returns a nested object of a recorded request body.
func bodyOf(t *testing.T, r recordedRequest, key string) map[string]any {
	t.Helper()
	value, ok := r.Body[key].(map[string]any)
	if !ok {
		t.Fatalf("request body has no %q object: %v", key, r.Body)
	}
	return value
}

func ptr[T any](value T) *T { return &value }
