package nusii

import (
	"context"
	"net/http"
)

// TaxationsService manages the taxes and discounts applied to a proposal's total.
type TaxationsService struct{ n *Nusii }

// TaxationParams are the attributes of a tax to create or update. A new tax
// needs Name and Percentage, or TaxID.
type TaxationParams struct {
	Name Opt[string] `json:"name,omitzero"`
	// Percentage is between -100 and 100. Negative values work as a discount.
	Percentage Opt[float64] `json:"percentage,omitzero"`
	// TaxID copies the name and percentage from one of the account's saved taxes.
	TaxID Opt[int64] `json:"tax_id,omitzero"`
}

// List returns every tax and discount of a proposal.
func (s *TaxationsService) List(ctx context.Context, proposalID int64, opts ...Option) ([]Taxation, error) {
	return collection[Taxation](ctx, s.n, "/proposals/"+id(proposalID)+"/taxations", opts)
}

// Create adds a tax (or, with a negative percentage, a discount) to a proposal.
func (s *TaxationsService) Create(ctx context.Context, proposalID int64, params TaxationParams, opts ...Option) (*Taxation, error) {
	return resource[Taxation](ctx, s.n, http.MethodPost, "/proposals/"+id(proposalID)+"/taxations", params, opts)
}

// Update changes a tax.
func (s *TaxationsService) Update(ctx context.Context, proposalID, taxationID int64, params TaxationParams, opts ...Option) (*Taxation, error) {
	return resource[Taxation](ctx, s.n, http.MethodPut, "/proposals/"+id(proposalID)+"/taxations/"+id(taxationID), params, opts)
}

// Delete removes a tax and returns it as it was.
func (s *TaxationsService) Delete(ctx context.Context, proposalID, taxationID int64, opts ...Option) (*Taxation, error) {
	return resource[Taxation](ctx, s.n, http.MethodDelete, "/proposals/"+id(proposalID)+"/taxations/"+id(taxationID), nil, opts)
}
