package nusii

import (
	"context"
	"iter"
	"net/http"
	"net/url"
)

// WebhookEndpointsService manages the URLs Nusii POSTs events to.
type WebhookEndpointsService struct{ n *Nusii }

// WebhookEndpointCreateParams subscribe a URL to events.
type WebhookEndpointCreateParams struct {
	// TargetURL is the URL Nusii POSTs the event JSON to.
	TargetURL string             `json:"target_url"`
	Events    []WebhookEventName `json:"events"`
}

// List returns one page of webhook endpoints. params may be nil.
func (s *WebhookEndpointsService) List(ctx context.Context, params *ListParams, opts ...Option) (*Page[WebhookEndpoint], error) {
	page, perPage := params.pagination()
	return listPage[WebhookEndpoint](ctx, s.n, "/webhook_endpoints", url.Values{}, page, perPage, opts)
}

// ListAll iterates over every webhook endpoint, fetching pages as the loop reaches them.
func (s *WebhookEndpointsService) ListAll(ctx context.Context, params *ListParams, opts ...Option) iter.Seq2[WebhookEndpoint, error] {
	return listAll(ctx, func() (*Page[WebhookEndpoint], error) { return s.List(ctx, params, opts...) })
}

// Get returns a webhook endpoint.
func (s *WebhookEndpointsService) Get(ctx context.Context, endpointID int64, opts ...Option) (*WebhookEndpoint, error) {
	return resource[WebhookEndpoint](ctx, s.n, http.MethodGet, "/webhook_endpoints/"+id(endpointID), nil, opts)
}

// Create subscribes a URL to events. Nusii POSTs each event as JSON; parse it
// with ParseWebhookEvent. Answering a delivery with HTTP 410 removes the endpoint.
func (s *WebhookEndpointsService) Create(ctx context.Context, params WebhookEndpointCreateParams, opts ...Option) (*WebhookEndpoint, error) {
	return resource[WebhookEndpoint](ctx, s.n, http.MethodPost, "/webhook_endpoints", map[string]any{"webhook_endpoint": params}, opts)
}

// Delete deletes a webhook endpoint and returns it as it was.
func (s *WebhookEndpointsService) Delete(ctx context.Context, endpointID int64, opts ...Option) (*WebhookEndpoint, error) {
	return resource[WebhookEndpoint](ctx, s.n, http.MethodDelete, "/webhook_endpoints/"+id(endpointID), nil, opts)
}
