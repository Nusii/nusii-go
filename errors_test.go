package nusii

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestErrorsMatchTheirStatus(t *testing.T) {
	cases := map[int]error{
		400: ErrBadRequest,
		401: ErrAuthentication,
		402: ErrPaymentRequired,
		403: ErrForbidden,
		404: ErrNotFound,
		405: ErrMethodNotAllowed,
		406: ErrNotAcceptable,
		410: ErrGone,
		422: ErrUnprocessableEntity,
		429: ErrRateLimit,
		500: ErrServer,
		503: ErrServiceUnavailable,
	}
	for status, want := range cases {
		err := newAPIError(status, http.Header{}, nil)
		if !errors.Is(err, want) {
			t.Errorf("%d does not match %v", status, want)
		}
		if status != 404 && errors.Is(err, ErrNotFound) {
			t.Errorf("%d matches ErrNotFound", status)
		}
	}
	if !errors.Is(newAPIError(503, http.Header{}, nil), ErrServer) {
		t.Error("503 does not match ErrServer")
	}
	if errors.Is(newAPIError(418, http.Header{}, nil), ErrServer) {
		t.Error("418 matches ErrServer")
	}
}

func TestErrorBodies(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		body    string
		message string
		code    string
		errors  int
	}{
		{
			name:    "JSON:API validation errors",
			status:  422,
			body:    `{"errors":[{"source":{"pointer":"/data/attributes/name"},"detail":"can't be blank"},{"source":{"pointer":"/data/attributes/email"},"detail":"is invalid"},{"source":{"pointer":"/data/attributes/base"},"detail":"Something else"}]}`,
			message: "nusii: 422 name can't be blank; email is invalid; Something else",
			errors:  3,
		},
		{
			name:    "error code with message",
			status:  422,
			body:    `{"error":"proposal_locked","message":"This proposal is locked"}`,
			message: "nusii: 422 This proposal is locked",
			code:    "proposal_locked",
		},
		{
			name:    "error code alone",
			status:  422,
			body:    `{"error":"status_not_writable"}`,
			message: "nusii: 422 status_not_writable",
			code:    "status_not_writable",
		},
		{
			name:    "error sentence",
			status:  400,
			body:    `{"error":"Invalid request body"}`,
			message: "nusii: 400 Invalid request body",
		},
		{
			name:    "nested error",
			status:  429,
			body:    `{"error":{"status":429,"name":"Too Many Requests","message":"Slow down"}}`,
			message: "nusii: 429 Slow down",
		},
		{
			name:    "message",
			status:  401,
			body:    `{"message":"Authentication required or invalid"}`,
			message: "nusii: 401 Authentication required or invalid",
		},
		{
			name:    "plain text",
			status:  502,
			body:    "  Bad gateway from the proxy \n",
			message: "nusii: 502 Bad gateway from the proxy",
		},
		{
			name:    "empty body",
			status:  404,
			body:    "",
			message: "nusii: 404 Not Found",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := newAPIError(c.status, http.Header{}, []byte(c.body))
			if err.Error() != c.message {
				t.Errorf("got message %q, want %q", err.Error(), c.message)
			}
			if err.Code != c.code {
				t.Errorf("got code %q, want %q", err.Code, c.code)
			}
			if len(err.Errors) != c.errors {
				t.Errorf("got %d errors, want %d", len(err.Errors), c.errors)
			}
			if string(err.Body) != c.body {
				t.Errorf("got body %q", err.Body)
			}
		})
	}
}

func TestValidationErrorDetails(t *testing.T) {
	body := map[string]any{"errors": []any{map[string]any{
		"source": map[string]any{"pointer": "/data/attributes/name"},
		"detail": "can't be blank",
	}}}
	n, _ := newMockClient(t, []http.HandlerFunc{reply(http.StatusUnprocessableEntity, body)})

	_, err := n.Clients.Create(context.Background(), ClientCreateParams{Email: "jane@example.com"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !errors.Is(err, ErrUnprocessableEntity) {
		t.Fatalf("got %v", err)
	}
	want := []JSONAPIError{{Source: &JSONAPIErrorSource{Pointer: "/data/attributes/name"}, Detail: "can't be blank"}}
	if !reflect.DeepEqual(apiErr.Errors, want) {
		t.Errorf("got %+v", apiErr.Errors)
	}
	if apiErr.StatusCode != 422 || apiErr.Header.Get("Content-Type") == "" {
		t.Errorf("got status %d and header %v", apiErr.StatusCode, apiErr.Header)
	}
}

func TestLongTextBodiesAreTruncated(t *testing.T) {
	err := newAPIError(500, http.Header{}, []byte(strings.Repeat("x", 600)))
	if len(err.Message) != 500 {
		t.Errorf("got %d characters", len(err.Message))
	}
}

func TestParseRetryAfter(t *testing.T) {
	cases := []struct {
		value string
		want  time.Duration
		ok    bool
	}{
		{"", 0, false},
		{"abc", 0, false},
		{"5", 5 * time.Second, true},
		{"1.5", 1500 * time.Millisecond, true},
		{"-3", 0, true},
		{time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat), 0, true},
	}
	for _, c := range cases {
		got, ok := parseRetryAfter(c.value)
		if got != c.want || ok != c.ok {
			t.Errorf("%q: got %v %v, want %v %v", c.value, got, ok, c.want, c.ok)
		}
	}

	future := time.Now().Add(30 * time.Second).UTC().Format(http.TimeFormat)
	if got, ok := parseRetryAfter(future); !ok || got < 28*time.Second || got > 31*time.Second {
		t.Errorf("date: got %v %v", got, ok)
	}
}
