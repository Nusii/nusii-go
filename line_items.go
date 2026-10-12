package nusii

import (
	"context"
	"net/http"
)

// LineItemsService manages the line items of cost sections. Amounts are in cents.
type LineItemsService struct{ n *Nusii }

// LineItemParams are the attributes of a line item to create or update.
// Attributes left zero are not sent.
type LineItemParams struct {
	Name Opt[string] `json:"name,omitzero"`
	// Amount is the price in cents. For a range, the low end.
	Amount Opt[int64] `json:"amount,omitzero"`
	// MaximumAmount is the high end of a price range in cents, used with
	// CostTypeRange. Requires the price ranges feature on the account.
	MaximumAmount Opt[int64]    `json:"maximum_amount,omitzero"`
	CostType      CostType      `json:"cost_type,omitempty"`
	RecurringType RecurringType `json:"recurring_type,omitempty"`
	PerType       PerType       `json:"per_type,omitempty"`
	ChoiceType    ChoiceType    `json:"choice_type,omitempty"`
	Quantity      Opt[float64]  `json:"quantity,omitzero"`
	Position      Opt[int64]    `json:"position,omitzero"`
}

// List returns every line item of a section, ordered by position.
func (s *LineItemsService) List(ctx context.Context, sectionID int64, opts ...Option) ([]LineItem, error) {
	return collection[LineItem](ctx, s.n, "/sections/"+id(sectionID)+"/line_items", opts)
}

// Get returns a line item.
func (s *LineItemsService) Get(ctx context.Context, lineItemID int64, opts ...Option) (*LineItem, error) {
	return resource[LineItem](ctx, s.n, http.MethodGet, "/line_items/"+id(lineItemID), nil, opts)
}

// Create adds a line item to a cost section.
func (s *LineItemsService) Create(ctx context.Context, sectionID int64, params LineItemParams, opts ...Option) (*LineItem, error) {
	return resource[LineItem](ctx, s.n, http.MethodPost, "/sections/"+id(sectionID)+"/line_items", map[string]any{"line_item": params}, opts)
}

// Update changes a line item.
func (s *LineItemsService) Update(ctx context.Context, lineItemID int64, params LineItemParams, opts ...Option) (*LineItem, error) {
	return resource[LineItem](ctx, s.n, http.MethodPut, "/line_items/"+id(lineItemID), map[string]any{"line_item": params}, opts)
}

// Delete deletes a line item and returns it as it was.
func (s *LineItemsService) Delete(ctx context.Context, lineItemID int64, opts ...Option) (*LineItem, error) {
	return resource[LineItem](ctx, s.n, http.MethodDelete, "/line_items/"+id(lineItemID), nil, opts)
}
