package nusii

import (
	"context"
	"iter"
	"net/http"
	"net/url"
)

// ProposalActivitiesService reads what happened to proposals: views, sends,
// acceptances, email opens and bounces.
type ProposalActivitiesService struct{ n *Nusii }

// ProposalActivityListParams filters the list of activities.
type ProposalActivityListParams struct {
	// Page is the page number, starting at 1.
	Page int
	// PerPage is the number of activities per page (the API default is 25).
	PerPage int
	// ProposalID lists the activities of this proposal.
	ProposalID int64
	// ClientID lists the activities of all proposals of this client.
	ClientID int64
}

func (p *ProposalActivityListParams) query() (url.Values, int, int) {
	query := url.Values{}
	if p == nil {
		return query, 0, 0
	}
	if p.ProposalID != 0 {
		query.Set("proposal_id", id(p.ProposalID))
	}
	if p.ClientID != 0 {
		query.Set("client_id", id(p.ClientID))
	}
	return query, p.Page, p.PerPage
}

// List returns one page of activities, newest first. params may be nil.
func (s *ProposalActivitiesService) List(ctx context.Context, params *ProposalActivityListParams, opts ...Option) (*Page[ProposalActivity], error) {
	query, page, perPage := params.query()
	return listPage[ProposalActivity](ctx, s.n, "/proposal_activities", query, page, perPage, opts)
}

// ListAll iterates over every activity matching params, fetching pages as the loop reaches them.
func (s *ProposalActivitiesService) ListAll(ctx context.Context, params *ProposalActivityListParams, opts ...Option) iter.Seq2[ProposalActivity, error] {
	return listAll(ctx, func() (*Page[ProposalActivity], error) { return s.List(ctx, params, opts...) })
}

// Get returns an activity.
func (s *ProposalActivitiesService) Get(ctx context.Context, activityID int64, opts ...Option) (*ProposalActivity, error) {
	return resource[ProposalActivity](ctx, s.n, http.MethodGet, "/proposal_activities/"+id(activityID), nil, opts)
}
