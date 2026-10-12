package nusii

// Lists of allowed values. These endpoints answer with plain JSON arrays, not JSON:API.

import (
	"context"
	"net/http"
)

// ThemesService lists the proposal themes.
type ThemesService struct{ n *Nusii }

// List returns the proposal themes. Use a theme's ID as a proposal's theme.
func (s *ThemesService) List(ctx context.Context, opts ...Option) ([]Theme, error) {
	return plain[[]Theme](ctx, s.n, http.MethodGet, "/themes", nil, opts)
}

// CurrenciesService lists the supported currencies.
type CurrenciesService struct{ n *Nusii }

// List returns the supported currencies. Use ISOCode for a client's or proposal's currency.
func (s *CurrenciesService) List(ctx context.Context, opts ...Option) ([]Currency, error) {
	return plain[[]Currency](ctx, s.n, http.MethodGet, "/currencies", nil, opts)
}

// LocalesService lists the supported languages.
type LocalesService struct{ n *Nusii }

// List returns the supported languages. Use Code for a client's locale.
func (s *LocalesService) List(ctx context.Context, opts ...Option) ([]Locale, error) {
	return plain[[]Locale](ctx, s.n, http.MethodGet, "/locales", nil, opts)
}

// PDFPageSizesService lists the supported PDF page sizes.
type PDFPageSizesService struct{ n *Nusii }

// List returns the supported PDF page sizes. Use ID for a client's pdf_page_size.
func (s *PDFPageSizesService) List(ctx context.Context, opts ...Option) ([]PDFPageSize, error) {
	return plain[[]PDFPageSize](ctx, s.n, http.MethodGet, "/pdf_page_sizes", nil, opts)
}
