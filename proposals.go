package nusii

import (
	"context"
	"iter"
	"net/http"
	"net/url"
)

// ProposalsService manages proposals.
type ProposalsService struct{ n *Nusii }

// ProposalListParams filters the list of proposals. Dates are YYYY-MM-DD and
// include the whole day.
type ProposalListParams struct {
	// Page is the page number, starting at 1.
	Page int
	// PerPage is the number of proposals per page (the API default is 25).
	PerPage int
	Status  ProposalStatus
	// Statuses matches any of these statuses.
	Statuses []ProposalStatus
	// Archived returns only archived proposals. Without it, only unarchived ones.
	Archived bool
	// Query searches the title, document number, public id and the client's name, business and email.
	Query string
	// RecipientEmail matches proposals sent to this recipient.
	RecipientEmail string
	// RecipientEmails matches proposals sent to any of these recipients (up to 30).
	RecipientEmails []string
	// NoActivityClientViewProposal returns only proposals the client has never viewed.
	NoActivityClientViewProposal bool
	CreatedAtAfter               string
	CreatedAtBefore              string
	SentAtAfter                  string
	SentAtBefore                 string
	ExpiresAtAfter               string
	ExpiresAtBefore              string
	AcceptedAtAfter              string
	AcceptedAtBefore             string
}

func (p *ProposalListParams) query() (url.Values, int, int) {
	query := url.Values{}
	if p == nil {
		return query, 0, 0
	}
	set := func(key, value string) {
		if value != "" {
			query.Set(key, value)
		}
	}
	set("status", string(p.Status))
	for _, status := range p.Statuses {
		query.Add("statuses[]", string(status))
	}
	if p.Archived {
		query.Set("archived", "true")
	}
	set("query", p.Query)
	set("recipient_email", p.RecipientEmail)
	if len(p.RecipientEmails) > 0 {
		query.Set("recipient_emails", commaList(p.RecipientEmails))
	}
	if p.NoActivityClientViewProposal {
		query.Set("no_activity_client_view_proposal", "true")
	}
	set("created_at_after", p.CreatedAtAfter)
	set("created_at_before", p.CreatedAtBefore)
	set("sent_at_after", p.SentAtAfter)
	set("sent_at_before", p.SentAtBefore)
	set("expires_at_after", p.ExpiresAtAfter)
	set("expires_at_before", p.ExpiresAtBefore)
	set("accepted_at_after", p.AcceptedAtAfter)
	set("accepted_at_before", p.AcceptedAtBefore)
	return query, p.Page, p.PerPage
}

// ProposalCreateParams are the attributes of a new proposal. All are optional.
type ProposalCreateParams struct {
	// TemplateID creates the proposal from this template, copying its sections and settings.
	TemplateID Opt[int64]     `json:"template_id,omitzero"`
	Status     ProposalStatus `json:"status,omitempty"`
	Title      Opt[string]    `json:"title,omitzero"`
	ClientID   Opt[int64]     `json:"client_id,omitzero"`
	// PreparedByID is the user id of the team member the proposal is prepared by.
	PreparedByID Opt[int64] `json:"prepared_by_id,omitzero"`
	// ExpiresAt is the date (YYYY-MM-DD) after which the proposal can no longer be accepted.
	ExpiresAt Opt[string] `json:"expires_at,omitzero"`
	// DisplayDate is the date (YYYY-MM-DD) shown on the proposal instead of the creation date.
	DisplayDate          Opt[string] `json:"display_date,omitzero"`
	DocumentSectionTitle Opt[string] `json:"document_section_title,omitzero"`
	// Theme is "clean" (Modern) or "classic", see ThemesService.List.
	Theme             Opt[string] `json:"theme,omitzero"`
	Report            Opt[bool]   `json:"report,omitzero"`
	ExcludeTotal      Opt[bool]   `json:"exclude_total,omitzero"`
	ExcludeTotalInPDF Opt[bool]   `json:"exclude_total_in_pdf,omitzero"`
	// PaymentPercentage is the percentage of the total requested as payment on acceptance.
	PaymentPercentage Opt[int64]  `json:"payment_percentage,omitzero"`
	ArchivedAt        Opt[string] `json:"archived_at,omitzero"`
}

// ProposalUpdateParams are the attributes to change on a proposal. Attributes
// left zero are not changed. The status cannot be changed this way: it
// follows what happens to the proposal (sent, accepted, rejected).
type ProposalUpdateParams struct {
	Title    Opt[string] `json:"title,omitzero"`
	ClientID Opt[int64]  `json:"client_id,omitzero"`
	// PreparedByID is the user id of the team member the proposal is prepared by.
	PreparedByID Opt[int64] `json:"prepared_by_id,omitzero"`
	// ExpiresAt is the date (YYYY-MM-DD) after which the proposal can no longer be accepted.
	ExpiresAt Opt[string] `json:"expires_at,omitzero"`
	// DisplayDate is the date (YYYY-MM-DD) shown on the proposal instead of the creation date.
	DisplayDate          Opt[string] `json:"display_date,omitzero"`
	DocumentSectionTitle Opt[string] `json:"document_section_title,omitzero"`
	// Theme is "clean" (Modern) or "classic", see ThemesService.List.
	Theme             Opt[string] `json:"theme,omitzero"`
	Report            Opt[bool]   `json:"report,omitzero"`
	ExcludeTotal      Opt[bool]   `json:"exclude_total,omitzero"`
	ExcludeTotalInPDF Opt[bool]   `json:"exclude_total_in_pdf,omitzero"`
	// PaymentPercentage is the percentage of the total requested as payment on acceptance.
	PaymentPercentage Opt[int64] `json:"payment_percentage,omitzero"`
	// ArchivedAt archives the proposal, or unarchives it with Null.
	ArchivedAt Opt[string] `json:"archived_at,omitzero"`
}

// ProposalRecipient is someone a proposal is sent to.
type ProposalRecipient struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	// EligibleToSign controls whether this recipient can accept and sign. Defaults to true.
	EligibleToSign Opt[bool] `json:"eligible_to_sign,omitzero"`
}

// ProposalSendParams describe the email a proposal is sent with. Give
// Recipients or Email.
type ProposalSendParams struct {
	// Recipients are one or more recipients.
	Recipients []ProposalRecipient `json:"recipients,omitempty"`
	// Email is a single recipient, named after the client.
	Email   Opt[string] `json:"email,omitzero"`
	CC      Opt[string] `json:"cc,omitzero"`
	BCC     Opt[string] `json:"bcc,omitzero"`
	Subject Opt[string] `json:"subject,omitzero"`
	// Message is the email body. Supports basic HTML.
	Message Opt[string] `json:"message,omitzero"`
	// SenderID sends as this team member (by user id). Defaults to the account owner.
	SenderID Opt[int64] `json:"sender_id,omitzero"`
	// SenderEmail sends as this team member (by email). Defaults to the account owner.
	SenderEmail Opt[string] `json:"sender_email,omitzero"`
	// SaveEmailTemplate saves the subject and message as the default email template.
	SaveEmailTemplate Opt[bool] `json:"save_email_template,omitzero"`
}

// List returns one page of proposals, newest first. Archived proposals are
// left out unless params.Archived is set. params may be nil.
func (s *ProposalsService) List(ctx context.Context, params *ProposalListParams, opts ...Option) (*Page[Proposal], error) {
	query, page, perPage := params.query()
	return listPage[Proposal](ctx, s.n, "/proposals", query, page, perPage, opts)
}

// ListAll iterates over every proposal matching params, fetching pages as the loop reaches them.
func (s *ProposalsService) ListAll(ctx context.Context, params *ProposalListParams, opts ...Option) iter.Seq2[Proposal, error] {
	return listAll(ctx, func() (*Page[Proposal], error) { return s.List(ctx, params, opts...) })
}

// Get returns a proposal.
func (s *ProposalsService) Get(ctx context.Context, proposalID int64, opts ...Option) (*Proposal, error) {
	return resource[Proposal](ctx, s.n, http.MethodGet, "/proposals/"+id(proposalID), nil, opts)
}

// Create creates a proposal, from scratch or from a template with TemplateID.
func (s *ProposalsService) Create(ctx context.Context, params ProposalCreateParams, opts ...Option) (*Proposal, error) {
	return resource[Proposal](ctx, s.n, http.MethodPost, "/proposals", map[string]any{"proposal": params}, opts)
}

// Update changes a proposal.
func (s *ProposalsService) Update(ctx context.Context, proposalID int64, params ProposalUpdateParams, opts ...Option) (*Proposal, error) {
	return resource[Proposal](ctx, s.n, http.MethodPut, "/proposals/"+id(proposalID), map[string]any{"proposal": params}, opts)
}

// Delete deletes a proposal and returns it as it was.
func (s *ProposalsService) Delete(ctx context.Context, proposalID int64, opts ...Option) (*Proposal, error) {
	return resource[Proposal](ctx, s.n, http.MethodDelete, "/proposals/"+id(proposalID), nil, opts)
}

// Archive archives a proposal.
func (s *ProposalsService) Archive(ctx context.Context, proposalID int64, opts ...Option) (*Proposal, error) {
	return resource[Proposal](ctx, s.n, http.MethodPut, "/proposals/"+id(proposalID)+"/archive", nil, opts)
}

// Send emails a proposal to one or more recipients.
func (s *ProposalsService) Send(ctx context.Context, proposalID int64, params ProposalSendParams, opts ...Option) (*ProposalSendResult, error) {
	result, err := plain[ProposalSendResult](ctx, s.n, http.MethodPut, "/proposals/"+id(proposalID)+"/send_proposal", params, opts)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
