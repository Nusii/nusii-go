package nusii

import (
	"context"
	"iter"
	"net/url"
)

// UsersService reads the team members of the account.
type UsersService struct{ n *Nusii }

// List returns one page of the account's team members. params may be nil.
func (s *UsersService) List(ctx context.Context, params *ListParams, opts ...Option) (*Page[User], error) {
	page, perPage := params.pagination()
	return listPage[User](ctx, s.n, "/users", url.Values{}, page, perPage, opts)
}

// ListAll iterates over every team member, fetching pages as the loop reaches them.
func (s *UsersService) ListAll(ctx context.Context, params *ListParams, opts ...Option) iter.Seq2[User, error] {
	return listAll(ctx, func() (*Page[User], error) { return s.List(ctx, params, opts...) })
}
