package nusii

import (
	"context"
	"net/http"
)

// AccountService reads the account the credentials belong to.
type AccountService struct{ n *Nusii }

// Me returns the account the API key or access token belongs to.
func (s *AccountService) Me(ctx context.Context, opts ...Option) (*Account, error) {
	return resource[Account](ctx, s.n, http.MethodGet, "/account/me", nil, opts)
}
