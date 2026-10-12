package nusii

import (
	"context"
	"iter"
	"net/http"
	"net/url"
)

// TemplatesService reads the account's templates and Nusii's public template gallery.
type TemplatesService struct{ n *Nusii }

// TemplateListParams selects which templates to list.
type TemplateListParams struct {
	// Page is the page number, starting at 1.
	Page int
	// PerPage is the number of templates per page (the API default is 25).
	PerPage int
	// PublicTemplates lists Nusii's public template gallery instead of the account's templates.
	PublicTemplates bool
}

func (p *TemplateListParams) query() (url.Values, int, int) {
	query := url.Values{}
	if p == nil {
		return query, 0, 0
	}
	if p.PublicTemplates {
		// The API compares the value to the literal string "true".
		query.Set("public_templates", "true")
	}
	return query, p.Page, p.PerPage
}

// List returns one page of templates. params may be nil.
func (s *TemplatesService) List(ctx context.Context, params *TemplateListParams, opts ...Option) (*Page[TemplateSummary], error) {
	query, page, perPage := params.query()
	return listPage[TemplateSummary](ctx, s.n, "/templates", query, page, perPage, opts)
}

// ListAll iterates over every template matching params, fetching pages as the loop reaches them.
func (s *TemplatesService) ListAll(ctx context.Context, params *TemplateListParams, opts ...Option) iter.Seq2[TemplateSummary, error] {
	return listAll(ctx, func() (*Page[TemplateSummary], error) { return s.List(ctx, params, opts...) })
}

// Get returns a template.
func (s *TemplatesService) Get(ctx context.Context, templateID int64, opts ...Option) (*Template, error) {
	return resource[Template](ctx, s.n, http.MethodGet, "/templates/"+id(templateID), nil, opts)
}
