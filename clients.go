package nusii

import (
	"context"
	"iter"
	"net/http"
	"net/url"
)

// ClientsService manages the clients (customers) you send proposals to.
type ClientsService struct{ n *Nusii }

// ClientListParams filters the list of clients.
type ClientListParams struct {
	// Page is the page number, starting at 1.
	Page int
	// PerPage is the number of clients per page (the API default is 25).
	PerPage int
	// Query searches name, surname, business and email.
	Query string
	// Email matches one email exactly.
	Email string
	// Emails matches any of these emails (up to 30).
	Emails []string
}

func (p *ClientListParams) query() (url.Values, int, int) {
	query := url.Values{}
	if p == nil {
		return query, 0, 0
	}
	if p.Query != "" {
		query.Set("query", p.Query)
	}
	if p.Email != "" {
		query.Set("email", p.Email)
	}
	if len(p.Emails) > 0 {
		query.Set("emails", commaList(p.Emails))
	}
	return query, p.Page, p.PerPage
}

// ClientCreateParams are the attributes of a new client. Name and Email are required.
type ClientCreateParams struct {
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	Surname   Opt[string] `json:"surname,omitzero"`
	Business  Opt[string] `json:"business,omitzero"`
	Telephone Opt[string] `json:"telephone,omitzero"`
	Web       Opt[string] `json:"web,omitzero"`
	// Currency is an ISO 4217 code, see CurrenciesService.List.
	Currency Opt[string] `json:"currency,omitzero"`
	// Locale is a language code, see LocalesService.List.
	Locale Opt[string] `json:"locale,omitzero"`
	// PDFPageSize is a page size id, see PDFPageSizesService.List.
	PDFPageSize Opt[string] `json:"pdf_page_size,omitzero"`
	TimeZone    Opt[string] `json:"time_zone,omitzero"`
	Notes       Opt[string] `json:"notes,omitzero"`
	ClientRole  Opt[string] `json:"client_role,omitzero"`
	VATNumber   Opt[string] `json:"vat_number,omitzero"`
	Address     Opt[string] `json:"address,omitzero"`
	City        Opt[string] `json:"city,omitzero"`
	Postcode    Opt[string] `json:"postcode,omitzero"`
	State       Opt[string] `json:"state,omitzero"`
	Country     Opt[string] `json:"country,omitzero"`
}

// ClientUpdateParams are the attributes to change on a client. Attributes
// left zero are not changed.
type ClientUpdateParams struct {
	Name      Opt[string] `json:"name,omitzero"`
	Email     Opt[string] `json:"email,omitzero"`
	Surname   Opt[string] `json:"surname,omitzero"`
	Business  Opt[string] `json:"business,omitzero"`
	Telephone Opt[string] `json:"telephone,omitzero"`
	Web       Opt[string] `json:"web,omitzero"`
	// Currency is an ISO 4217 code, see CurrenciesService.List.
	Currency Opt[string] `json:"currency,omitzero"`
	// Locale is a language code, see LocalesService.List.
	Locale Opt[string] `json:"locale,omitzero"`
	// PDFPageSize is a page size id, see PDFPageSizesService.List.
	PDFPageSize Opt[string] `json:"pdf_page_size,omitzero"`
	TimeZone    Opt[string] `json:"time_zone,omitzero"`
	Notes       Opt[string] `json:"notes,omitzero"`
	ClientRole  Opt[string] `json:"client_role,omitzero"`
	VATNumber   Opt[string] `json:"vat_number,omitzero"`
	Address     Opt[string] `json:"address,omitzero"`
	City        Opt[string] `json:"city,omitzero"`
	Postcode    Opt[string] `json:"postcode,omitzero"`
	State       Opt[string] `json:"state,omitzero"`
	Country     Opt[string] `json:"country,omitzero"`
}

// List returns one page of clients, newest first. params may be nil.
func (s *ClientsService) List(ctx context.Context, params *ClientListParams, opts ...Option) (*Page[Client], error) {
	query, page, perPage := params.query()
	return listPage[Client](ctx, s.n, "/clients", query, page, perPage, opts)
}

// ListAll iterates over every client matching params, fetching pages as the loop reaches them.
func (s *ClientsService) ListAll(ctx context.Context, params *ClientListParams, opts ...Option) iter.Seq2[Client, error] {
	return listAll(ctx, func() (*Page[Client], error) { return s.List(ctx, params, opts...) })
}

// Get returns a client.
func (s *ClientsService) Get(ctx context.Context, clientID int64, opts ...Option) (*Client, error) {
	return resource[Client](ctx, s.n, http.MethodGet, "/clients/"+id(clientID), nil, opts)
}

// Create creates a client.
func (s *ClientsService) Create(ctx context.Context, params ClientCreateParams, opts ...Option) (*Client, error) {
	return resource[Client](ctx, s.n, http.MethodPost, "/clients", map[string]any{"client": params}, opts)
}

// Update changes a client.
func (s *ClientsService) Update(ctx context.Context, clientID int64, params ClientUpdateParams, opts ...Option) (*Client, error) {
	return resource[Client](ctx, s.n, http.MethodPut, "/clients/"+id(clientID), map[string]any{"client": params}, opts)
}

// Delete deletes a client and returns it as it was.
func (s *ClientsService) Delete(ctx context.Context, clientID int64, opts ...Option) (*Client, error) {
	return resource[Client](ctx, s.n, http.MethodDelete, "/clients/"+id(clientID), nil, opts)
}
