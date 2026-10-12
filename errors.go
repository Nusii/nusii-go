package nusii

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Errors to match with errors.Is. Every *APIError matches the error of its
// HTTP status, and every *ConnectionError matches ErrConnection:
//
//	if errors.Is(err, nusii.ErrNotFound) { ... }
var (
	// ErrBadRequest (400): the request was malformed, e.g. a filter given as an object instead of a string.
	ErrBadRequest = errors.New("nusii: bad request")
	// ErrAuthentication (401): the API key or access token is missing or invalid.
	ErrAuthentication = errors.New("nusii: authentication failed")
	// ErrPaymentRequired (402): the account has reached a plan limit, e.g. its active proposal limit.
	ErrPaymentRequired = errors.New("nusii: payment required")
	// ErrForbidden (403): not allowed, e.g. missing OAuth scope, suspended account or unconfirmed sender email.
	ErrForbidden = errors.New("nusii: forbidden")
	// ErrNotFound (404): the record does not exist or belongs to another account.
	ErrNotFound = errors.New("nusii: not found")
	// ErrMethodNotAllowed (405): the HTTP method is not supported for this endpoint.
	ErrMethodNotAllowed = errors.New("nusii: method not allowed")
	// ErrNotAcceptable (406): the Accept header is not supported.
	ErrNotAcceptable = errors.New("nusii: not acceptable")
	// ErrGone (410): the record has already been deleted.
	ErrGone = errors.New("nusii: gone")
	// ErrUnprocessableEntity (422): validation failed, or the proposal is locked.
	// APIError.Errors lists the invalid fields.
	ErrUnprocessableEntity = errors.New("nusii: unprocessable entity")
	// ErrRateLimit (429): too many requests. Nusii allows 100 requests per 30 seconds per token.
	ErrRateLimit = errors.New("nusii: rate limit exceeded")
	// ErrServer (5xx): something went wrong on Nusii's side.
	ErrServer = errors.New("nusii: server error")
	// ErrServiceUnavailable (503): the API is temporarily unavailable. Also matches ErrServer.
	ErrServiceUnavailable = errors.New("nusii: service unavailable")

	// ErrConnection: the request never got a response (DNS failure, refused
	// connection, reset socket...). Matched by every *ConnectionError.
	ErrConnection = errors.New("nusii: connection failed")
	// ErrTimeout: the request took longer than the configured timeout. Also matches ErrConnection.
	ErrTimeout = errors.New("nusii: request timed out")

	// ErrMissingCredentials is returned by New when no API key or access token is given or set in the environment.
	ErrMissingCredentials = errors.New("nusii: missing credentials. Pass nusii.WithAPIKey or nusii.WithAccessToken to nusii.New, or set the NUSII_API_KEY environment variable")
	// ErrInvalidWebhookPayload is returned by ParseWebhookEvent for a body that is not a Nusii webhook event.
	ErrInvalidWebhookPayload = errors.New("nusii: invalid webhook payload")
)

var statusErrors = map[int]error{
	http.StatusBadRequest:          ErrBadRequest,
	http.StatusUnauthorized:        ErrAuthentication,
	http.StatusPaymentRequired:     ErrPaymentRequired,
	http.StatusForbidden:           ErrForbidden,
	http.StatusNotFound:            ErrNotFound,
	http.StatusMethodNotAllowed:    ErrMethodNotAllowed,
	http.StatusNotAcceptable:       ErrNotAcceptable,
	http.StatusGone:                ErrGone,
	http.StatusUnprocessableEntity: ErrUnprocessableEntity,
	http.StatusTooManyRequests:     ErrRateLimit,
	http.StatusServiceUnavailable:  ErrServiceUnavailable,
}

// JSONAPIError is a single entry of a JSON:API errors array, as returned for
// validation failures.
type JSONAPIError struct {
	Source *JSONAPIErrorSource `json:"source,omitempty"`
	Detail string              `json:"detail,omitempty"`
	Title  string              `json:"title,omitempty"`
	Status string              `json:"status,omitempty"`
}

// JSONAPIErrorSource points at the attribute an error is about, e.g.
// "/data/attributes/name".
type JSONAPIErrorSource struct {
	Pointer string `json:"pointer,omitempty"`
}

// APIError is returned when the Nusii API answers with an HTTP error status.
// Match the status with errors.Is (see ErrNotFound and the other errors), or
// read the details with errors.As:
//
//	var apiErr *nusii.APIError
//	if errors.As(err, &apiErr) {
//		fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Errors)
//	}
type APIError struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int
	// Message describes the error, e.g. "name can't be blank; email is invalid".
	Message string
	// Code is a machine readable error code when the API sends one, e.g. "proposal_locked".
	Code string
	// Errors are the field level validation errors (JSON:API format), when present.
	Errors []JSONAPIError
	// Header holds the response headers.
	Header http.Header
	// Body is the raw response body.
	Body []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("nusii: %d %s", e.StatusCode, e.Message)
}

// Is reports whether the error matches target, one of the status errors such
// as ErrNotFound.
func (e *APIError) Is(target error) bool {
	if target == ErrServer {
		return e.StatusCode >= 500
	}
	return target == statusErrors[e.StatusCode]
}

// RetryAfter returns how long to wait before retrying, from the Retry-After
// header of a 429 response. ok is false when the header is missing.
func (e *APIError) RetryAfter() (wait time.Duration, ok bool) {
	return parseRetryAfter(e.Header.Get("Retry-After"))
}

// ConnectionError is returned when a request got no response: DNS failure,
// refused connection, reset socket, or a timeout. It matches ErrConnection,
// and ErrTimeout when Timeout is true. Unwrap returns the underlying error.
type ConnectionError struct {
	// URL is the URL of the request.
	URL string
	// Timeout reports whether the request took longer than the configured timeout.
	Timeout bool
	// Err is the underlying error.
	Err error

	timeout time.Duration
}

func (e *ConnectionError) Error() string {
	if e.Timeout {
		return fmt.Sprintf("nusii: request to %s timed out after %s", e.URL, e.timeout)
	}
	return fmt.Sprintf("nusii: could not reach the Nusii API at %s: %v", e.URL, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// Is reports whether target is ErrConnection, or ErrTimeout for a timeout.
func (e *ConnectionError) Is(target error) bool {
	return target == ErrConnection || (target == ErrTimeout && e.Timeout)
}

func newAPIError(status int, header http.Header, body []byte) *APIError {
	message, code, errs := describeErrorBody(body)
	if message == "" {
		message = http.StatusText(status)
	}
	if message == "" {
		message = "Request failed"
	}
	return &APIError{
		StatusCode: status,
		Message:    message,
		Code:       code,
		Errors:     errs,
		Header:     header,
		Body:       body,
	}
}

func parseRetryAfter(value string) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		return time.Duration(max(seconds, 0) * float64(time.Second)), true
	}
	if date, err := http.ParseTime(value); err == nil {
		return max(time.Until(date).Round(time.Second), 0), true
	}
	return 0, false
}

var errorCodePattern = regexp.MustCompile(`^[a-z_]+$`)

// describeErrorBody reads the message, error code and validation errors of an
// error response. The API reports errors in several shapes depending on where
// they originate:
//
//	{"errors": [{"source": {"pointer": "/data/attributes/name"}, "detail": "can't be blank"}]}
//	{"error": "proposal_locked", "message": "This proposal is locked..."}
//	{"error": "Invalid request body"}
//	{"error": {"status": 429, "name": "Too Many Requests", "message": "..."}}
//	{"message": "Authentication required or invalid"}
func describeErrorBody(body []byte) (message, code string, errs []JSONAPIError) {
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		text := strings.TrimSpace(string(body))
		if runes := []rune(text); len(runes) > 500 {
			text = string(runes[:500])
		}
		return text, "", nil
	}
	if text, ok := parsed.(string); ok {
		return strings.TrimSpace(text), "", nil
	}
	object, ok := parsed.(map[string]any)
	if !ok {
		return "", "", nil
	}

	if list, ok := object["errors"].([]any); ok && len(list) > 0 {
		var document struct {
			Errors []JSONAPIError `json:"errors"`
		}
		if json.Unmarshal(body, &document) == nil && len(document.Errors) > 0 {
			details := make([]string, len(document.Errors))
			for i, e := range document.Errors {
				details[i] = formatJSONAPIError(e)
			}
			return strings.Join(details, "; "), "", document.Errors
		}
	}

	msg, _ := object["message"].(string)
	switch e := object["error"].(type) {
	case string:
		if msg != "" {
			return msg, e, nil
		}
		if errorCodePattern.MatchString(e) {
			return e, e, nil
		}
		return e, "", nil
	case map[string]any:
		if nested, ok := e["message"].(string); ok {
			return nested, "", nil
		}
	}
	return msg, "", nil
}

func formatJSONAPIError(e JSONAPIError) string {
	detail := e.Detail
	if detail == "" {
		detail = e.Title
	}
	if detail == "" {
		detail = "is invalid"
	}
	if e.Source == nil {
		return detail
	}
	pointer := e.Source.Pointer
	attribute := pointer[strings.LastIndex(pointer, "/")+1:]
	if attribute == "" || attribute == "data" || attribute == "base" {
		return detail
	}
	return attribute + " " + detail
}
