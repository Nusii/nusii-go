package nusii

import (
	"context"
	"iter"
	"net/http"
	"net/url"
)

// SectionsService manages the sections of proposals and templates, and the
// reusable sections of the content library.
type SectionsService struct{ n *Nusii }

// SectionListParams selects whose sections to list. Without ProposalID or
// TemplateID, the account's reusable sections (content library) are listed.
type SectionListParams struct {
	// Page is the page number, starting at 1.
	Page int
	// PerPage is the number of sections per page (the API default is 25).
	PerPage int
	// ProposalID lists the sections of this proposal.
	ProposalID int64
	// TemplateID lists the sections of this template.
	TemplateID int64
}

func (p *SectionListParams) query() (url.Values, int, int) {
	query := url.Values{}
	if p == nil {
		return query, 0, 0
	}
	if p.ProposalID != 0 {
		query.Set("proposal_id", id(p.ProposalID))
	}
	if p.TemplateID != 0 {
		query.Set("template_id", id(p.TemplateID))
	}
	return query, p.Page, p.PerPage
}

// SectionParams are the attributes of a section to create, copy or update.
// Attributes left zero are not sent.
type SectionParams struct {
	// Title is the heading shown on the proposal page.
	Title Opt[string] `json:"title,omitzero"`
	// Name is shown in the editor navigation. Defaults to Title when only one is given.
	Name Opt[string] `json:"name,omitzero"`
	// Body is the HTML content.
	Body         Opt[string] `json:"body,omitzero"`
	SectionType  SectionType `json:"section_type,omitempty"`
	Position     Opt[int64]  `json:"position,omitzero"`
	Optional     Opt[bool]   `json:"optional,omitzero"`
	IncludeTotal Opt[bool]   `json:"include_total,omitzero"`
	PageBreak    Opt[bool]   `json:"page_break,omitzero"`
	// Reusable saves the section in the content library, not attached to a document.
	Reusable   Opt[bool]  `json:"reusable,omitzero"`
	ProposalID Opt[int64] `json:"proposal_id,omitzero"`
	TemplateID Opt[int64] `json:"template_id,omitzero"`
}

// List returns one page of sections, line items included. params may be nil.
func (s *SectionsService) List(ctx context.Context, params *SectionListParams, opts ...Option) (*Page[Section], error) {
	query, page, perPage := params.query()
	return listPage[Section](ctx, s.n, "/sections", query, page, perPage, opts)
}

// ListAll iterates over every section matching params, fetching pages as the loop reaches them.
func (s *SectionsService) ListAll(ctx context.Context, params *SectionListParams, opts ...Option) iter.Seq2[Section, error] {
	return listAll(ctx, func() (*Page[Section], error) { return s.List(ctx, params, opts...) })
}

// Get returns a section, line items included.
func (s *SectionsService) Get(ctx context.Context, sectionID int64, opts ...Option) (*Section, error) {
	return resource[Section](ctx, s.n, http.MethodGet, "/sections/"+id(sectionID), nil, opts)
}

// Create creates a section in a proposal (ProposalID), a template
// (TemplateID), or the content library (Reusable).
func (s *SectionsService) Create(ctx context.Context, params SectionParams, opts ...Option) (*Section, error) {
	return resource[Section](ctx, s.n, http.MethodPost, "/sections", map[string]any{"section": params}, opts)
}

// Copy copies a reusable section, line items included, into a proposal or
// template. The copy is not reusable itself unless params.Reusable is true.
func (s *SectionsService) Copy(ctx context.Context, reusableSectionID int64, params SectionParams, opts ...Option) (*Section, error) {
	// The API copies every attribute of the original, reusable included, and a
	// reusable section is detached from its proposal or template.
	if params.Reusable.IsZero() {
		params.Reusable = Bool(false)
	}
	body := map[string]any{"section_id": reusableSectionID, "section": params}
	return resource[Section](ctx, s.n, http.MethodPost, "/sections", body, opts)
}

// Update changes a section.
func (s *SectionsService) Update(ctx context.Context, sectionID int64, params SectionParams, opts ...Option) (*Section, error) {
	return resource[Section](ctx, s.n, http.MethodPut, "/sections/"+id(sectionID), map[string]any{"section": params}, opts)
}

// Delete deletes a section and returns it as it was.
func (s *SectionsService) Delete(ctx context.Context, sectionID int64, opts ...Option) (*Section, error) {
	return resource[Section](ctx, s.n, http.MethodDelete, "/sections/"+id(sectionID), nil, opts)
}
