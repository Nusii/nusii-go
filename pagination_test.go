package nusii

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"
)

func TestListReturnsAPage(t *testing.T) {
	n, api := newMockClient(t, []http.HandlerFunc{
		ok(many(meta(map[string]any{"current_page": 2, "next_page": 3, "prev_page": 1, "total_pages": 3, "total_count": 5}),
			res("clients", 3, map[string]any{"email": "a@example.com"}),
			res("clients", 4, map[string]any{"email": "b@example.com"}),
		)),
	})

	page, err := n.Clients.List(context.Background(), &ClientListParams{Page: 2, PerPage: 2})
	if err != nil {
		t.Fatal(err)
	}
	if q := api.request(0).Query; q.Get("page") != "2" || q.Get("per") != "2" || q.Has("per_page") {
		t.Errorf("got query %v", q)
	}
	if len(page.Data) != 2 || page.Data[0].ID != 3 || page.Data[1].Email != "b@example.com" {
		t.Errorf("got data %+v", page.Data)
	}
	want := PaginationMeta{CurrentPage: 2, NextPage: ptr(3), PrevPage: ptr(1), TotalPages: 3, TotalCount: 5}
	if page.Meta.CurrentPage != want.CurrentPage || *page.Meta.NextPage != 3 || *page.Meta.PrevPage != 1 ||
		page.Meta.TotalPages != want.TotalPages || page.Meta.TotalCount != want.TotalCount {
		t.Errorf("got meta %+v", page.Meta)
	}
	if !page.HasNextPage() || !page.HasPreviousPage() {
		t.Error("page has no neighbors")
	}
}

func TestListWithoutParams(t *testing.T) {
	n, api := newMockClient(t, []http.HandlerFunc{ok(many(nil, res("users", 1, nil)))})

	page, err := n.Users.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if q := api.request(0).Query; q.Get("page") != "1" || q.Has("per") {
		t.Errorf("got query %v", q)
	}
	// Without meta, the page describes itself.
	if page.Meta.CurrentPage != 1 || page.Meta.TotalPages != 1 || page.Meta.TotalCount != 1 || page.HasNextPage() {
		t.Errorf("got meta %+v", page.Meta)
	}
}

func TestPageNavigation(t *testing.T) {
	n, api := newMockClient(t, []http.HandlerFunc{
		ok(many(meta(map[string]any{"next_page": 2, "total_pages": 2, "total_count": 2}), res("templates", 1, nil))),
		ok(many(meta(map[string]any{"current_page": 2, "prev_page": 1, "total_pages": 2, "total_count": 2}), res("templates", 2, nil))),
		ok(many(meta(map[string]any{"next_page": 2, "total_pages": 2, "total_count": 2}), res("templates", 1, nil))),
	})
	ctx := context.Background()

	first, err := n.Templates.List(ctx, &TemplateListParams{PublicTemplates: true, PerPage: 1})
	if err != nil {
		t.Fatal(err)
	}
	if prev, err := first.PreviousPage(ctx); prev != nil || err != nil {
		t.Errorf("first page has a previous page: %v %v", prev, err)
	}

	second, err := first.NextPage(ctx)
	if err != nil || second.Data[0].ID != 2 {
		t.Fatalf("got %+v %v", second, err)
	}
	// The next page keeps the filters and page size.
	if q := api.request(1).Query; q.Get("page") != "2" || q.Get("per") != "1" || q.Get("public_templates") != "true" {
		t.Errorf("got query %v", q)
	}
	if last, err := second.NextPage(ctx); last != nil || err != nil {
		t.Errorf("last page has a next page: %v %v", last, err)
	}

	back, err := second.PreviousPage(ctx)
	if err != nil || back.Data[0].ID != 1 {
		t.Errorf("got %+v %v", back, err)
	}
}

func TestListAllFetchesPagesLazily(t *testing.T) {
	n, api := newMockClient(t, []http.HandlerFunc{
		ok(many(meta(map[string]any{"next_page": 2, "total_pages": 3, "total_count": 5}), res("proposals", 1, nil), res("proposals", 2, nil))),
		ok(many(meta(map[string]any{"current_page": 2, "next_page": 3, "prev_page": 1, "total_pages": 3, "total_count": 5}), res("proposals", 3, nil), res("proposals", 4, nil))),
		ok(many(meta(map[string]any{"current_page": 3, "prev_page": 2, "total_pages": 3, "total_count": 5}), res("proposals", 5, nil))),
	})

	var ids []int64
	for proposal, err := range n.Proposals.ListAll(context.Background(), &ProposalListParams{Status: ProposalStatusPending}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, proposal.ID)
		if proposal.ID == 3 {
			break
		}
	}
	if !slices.Equal(ids, []int64{1, 2, 3}) {
		t.Errorf("got ids %v", ids)
	}
	if api.requestCount() != 2 {
		t.Errorf("fetched %d pages, want 2", api.requestCount())
	}
	for i := range 2 {
		if api.request(i).Query.Get("status") != "pending" {
			t.Errorf("request %d lost the filter: %v", i, api.request(i).Query)
		}
	}
}

func TestListAllReportsErrors(t *testing.T) {
	n, _ := newMockClient(t, []http.HandlerFunc{
		ok(many(meta(map[string]any{"next_page": 2, "total_pages": 2, "total_count": 2}), res("users", 1, nil))),
		reply(http.StatusNotFound, map[string]any{"error": "Not found"}),
	})

	var ids []int64
	var failure error
	for user, err := range n.Users.ListAll(context.Background(), nil) {
		if err != nil {
			failure = err
			break
		}
		ids = append(ids, user.ID)
	}
	if !slices.Equal(ids, []int64{1}) || !errors.Is(failure, ErrNotFound) {
		t.Errorf("got ids %v and error %v", ids, failure)
	}

	n, _ = newMockClient(t, []http.HandlerFunc{reply(http.StatusUnauthorized, nil)})
	for _, err := range n.WebhookEndpoints.ListAll(context.Background(), nil) {
		if !errors.Is(err, ErrAuthentication) {
			t.Errorf("got %v", err)
		}
	}
}
