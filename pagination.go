package nusii

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
)

// PaginationMeta tells where a page sits in a paginated list.
type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	// NextPage is nil on the last page.
	NextPage *int `json:"next_page"`
	// PrevPage is nil on the first page.
	PrevPage   *int `json:"prev_page"`
	TotalPages int  `json:"total_pages"`
	TotalCount int  `json:"total_count"`
}

// ListParams selects a page of a list that has no other filters.
type ListParams struct {
	// Page is the page number, starting at 1.
	Page int
	// PerPage is the number of items per page (the API default is 25).
	PerPage int
}

func (p *ListParams) pagination() (page, perPage int) {
	if p == nil {
		return 0, 0
	}
	return p.Page, p.PerPage
}

// Page is one page of a paginated list.
type Page[T any] struct {
	// Data holds the items on this page.
	Data []T
	// Meta holds the current page, total pages and total count.
	Meta PaginationMeta

	fetch func(ctx context.Context, page int) (*Page[T], error)
}

// HasNextPage reports whether there is a page after this one.
func (p *Page[T]) HasNextPage() bool {
	return p.Meta.NextPage != nil
}

// HasPreviousPage reports whether there is a page before this one.
func (p *Page[T]) HasPreviousPage() bool {
	return p.Meta.PrevPage != nil
}

// NextPage fetches the next page. It returns nil and no error on the last page.
func (p *Page[T]) NextPage(ctx context.Context) (*Page[T], error) {
	if p.Meta.NextPage == nil || p.fetch == nil {
		return nil, nil
	}
	return p.fetch(ctx, *p.Meta.NextPage)
}

// PreviousPage fetches the previous page. It returns nil and no error on the first page.
func (p *Page[T]) PreviousPage(ctx context.Context) (*Page[T], error) {
	if p.Meta.PrevPage == nil || p.fetch == nil {
		return nil, nil
	}
	return p.fetch(ctx, *p.Meta.PrevPage)
}

// All iterates over every item from this page to the last one, fetching the
// next pages as the loop reaches them. A failed request ends the loop with
// the error.
func (p *Page[T]) All(ctx context.Context) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for page := p; page != nil; {
			for _, item := range page.Data {
				if !yield(item, nil) {
					return
				}
			}
			next, err := page.NextPage(ctx)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			page = next
		}
	}
}

// listPage requests a page of a paginated JSON:API collection. The returned
// page fetches its neighbors with the same filters and options.
func listPage[T any](ctx context.Context, n *Nusii, path string, filters url.Values, page, perPage int, opts []Option) (*Page[T], error) {
	var fetch func(ctx context.Context, number int) (*Page[T], error)
	fetch = func(ctx context.Context, number int) (*Page[T], error) {
		query := url.Values{}
		for key, values := range filters {
			query[key] = values
		}
		query.Set("page", strconv.Itoa(number))
		if perPage > 0 {
			// The API reads the page size from "per" (Kaminari), not "per_page".
			query.Set("per", strconv.Itoa(perPage))
		}

		body, err := n.do(ctx, RequestConfig{Method: http.MethodGet, Path: path, Query: query}, opts)
		if err != nil {
			return nil, err
		}
		data, document, err := decodeMany[T](body)
		if err != nil {
			return nil, err
		}

		meta := PaginationMeta{CurrentPage: number, TotalPages: 1, TotalCount: len(data)}
		if m := document.Meta; m != nil {
			if m.CurrentPage != nil {
				meta.CurrentPage = *m.CurrentPage
			}
			if m.TotalPages != nil {
				meta.TotalPages = *m.TotalPages
			}
			if m.TotalCount != nil {
				meta.TotalCount = *m.TotalCount
			}
			meta.NextPage = m.NextPage
			meta.PrevPage = m.PrevPage
		}
		return &Page[T]{Data: data, Meta: meta, fetch: fetch}, nil
	}

	return fetch(ctx, max(page, 1))
}

// listAll lazily yields every item of every page, starting at the first page
// the request selects.
func listAll[T any](ctx context.Context, first func() (*Page[T], error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		page, err := first()
		if err != nil {
			var zero T
			yield(zero, err)
			return
		}
		for item, err := range page.All(ctx) {
			if !yield(item, err) {
				return
			}
		}
	}
}
