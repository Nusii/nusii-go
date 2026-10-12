package nusii

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// resource requests a single JSON:API resource and flattens it.
func resource[T any](ctx context.Context, n *Nusii, method, path string, body any, opts []Option) (*T, error) {
	raw, err := n.do(ctx, RequestConfig{Method: method, Path: path, Body: body}, opts)
	if err != nil {
		return nil, err
	}
	return decodeOne[T](raw)
}

// collection requests an unpaginated JSON:API collection and flattens it.
func collection[T any](ctx context.Context, n *Nusii, path string, opts []Option) ([]T, error) {
	raw, err := n.do(ctx, RequestConfig{Method: http.MethodGet, Path: path}, opts)
	if err != nil {
		return nil, err
	}
	records, _, err := decodeMany[T](raw)
	return records, err
}

// plain requests an endpoint that answers with plain JSON, not JSON:API.
func plain[T any](ctx context.Context, n *Nusii, method, path string, body any, opts []Option) (T, error) {
	var result T
	raw, err := n.do(ctx, RequestConfig{Method: method, Path: path, Body: body}, opts)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, fmt.Errorf("nusii: decoding the response of %s %s: %w", method, path, err)
	}
	return result, nil
}

// id formats a record id for a URL path.
func id(value int64) string {
	return strconv.FormatInt(value, 10)
}

// commaList joins values in the comma separated form the API expects.
func commaList[T ~string](values []T) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = string(value)
	}
	return strings.Join(parts, ",")
}
