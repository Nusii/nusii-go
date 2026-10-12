package nusii

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

// TestEndpoints checks the method, path and body every method sends.
func TestEndpoints(t *testing.T) {
	ctx := context.Background()
	record := one(res("records", 7, nil))

	cases := []struct {
		name   string
		call   func(n *Nusii) error
		method string
		path   string
		body   string // JSON, empty for no body
		answer any
	}{
		{"account me", func(n *Nusii) error { _, err := n.Account.Me(ctx); return err }, "GET", "/account/me", "", record},

		{"clients get", func(n *Nusii) error { _, err := n.Clients.Get(ctx, 7); return err }, "GET", "/clients/7", "", record},
		{"clients create", func(n *Nusii) error {
			_, err := n.Clients.Create(ctx, ClientCreateParams{Name: "Jane", Email: "jane@example.com", Surname: String("Doe"), PDFPageSize: String("A4")})
			return err
		}, "POST", "/clients", `{"client":{"name":"Jane","email":"jane@example.com","surname":"Doe","pdf_page_size":"A4"}}`, record},
		{"clients update", func(n *Nusii) error {
			_, err := n.Clients.Update(ctx, 7, ClientUpdateParams{Business: String("Acme"), Notes: Null[string]()})
			return err
		}, "PUT", "/clients/7", `{"client":{"business":"Acme","notes":null}}`, record},
		{"clients delete", func(n *Nusii) error { _, err := n.Clients.Delete(ctx, 7); return err }, "DELETE", "/clients/7", "", record},

		{"proposals get", func(n *Nusii) error { _, err := n.Proposals.Get(ctx, 7); return err }, "GET", "/proposals/7", "", record},
		{"proposals create", func(n *Nusii) error {
			_, err := n.Proposals.Create(ctx, ProposalCreateParams{Title: String("Redesign"), ClientID: Int(3), TemplateID: Int(42), Report: Bool(false)})
			return err
		}, "POST", "/proposals", `{"proposal":{"template_id":42,"title":"Redesign","client_id":3,"report":false}}`, record},
		{"proposals create empty", func(n *Nusii) error {
			_, err := n.Proposals.Create(ctx, ProposalCreateParams{})
			return err
		}, "POST", "/proposals", `{"proposal":{}}`, record},
		{"proposals update", func(n *Nusii) error {
			_, err := n.Proposals.Update(ctx, 7, ProposalUpdateParams{ExpiresAt: Null[string](), ArchivedAt: Null[string]()})
			return err
		}, "PUT", "/proposals/7", `{"proposal":{"expires_at":null,"archived_at":null}}`, record},
		{"proposals delete", func(n *Nusii) error { _, err := n.Proposals.Delete(ctx, 7); return err }, "DELETE", "/proposals/7", "", record},
		{"proposals archive", func(n *Nusii) error { _, err := n.Proposals.Archive(ctx, 7); return err }, "PUT", "/proposals/7/archive", "", record},
		{"proposals send", func(n *Nusii) error {
			_, err := n.Proposals.Send(ctx, 7, ProposalSendParams{
				Recipients: []ProposalRecipient{{Name: "Jane", Email: "jane@example.com"}, {Name: "John", Email: "john@example.com", EligibleToSign: Bool(false)}},
				CC:         String("cc@example.com"),
				Subject:    String("Your proposal"),
			})
			return err
		}, "PUT", "/proposals/7/send_proposal",
			`{"recipients":[{"name":"Jane","email":"jane@example.com"},{"name":"John","email":"john@example.com","eligible_to_sign":false}],"cc":"cc@example.com","subject":"Your proposal"}`,
			map[string]any{"status": "pending", "sent_at": "2026-10-12T10:00:00.000Z", "sender_id": 1, "sender_name": "Jo"}},

		{"sections get", func(n *Nusii) error { _, err := n.Sections.Get(ctx, 7); return err }, "GET", "/sections/7", "", record},
		{"sections create", func(n *Nusii) error {
			_, err := n.Sections.Create(ctx, SectionParams{ProposalID: Int(3), Title: String("Scope"), SectionType: SectionTypeCost, Position: Int(1)})
			return err
		}, "POST", "/sections", `{"section":{"title":"Scope","section_type":"cost","position":1,"proposal_id":3}}`, record},
		{"sections copy", func(n *Nusii) error {
			_, err := n.Sections.Copy(ctx, 9, SectionParams{ProposalID: Int(3)})
			return err
		}, "POST", "/sections", `{"section_id":9,"section":{"reusable":false,"proposal_id":3}}`, record},
		{"sections copy reusable", func(n *Nusii) error {
			_, err := n.Sections.Copy(ctx, 9, SectionParams{Reusable: Bool(true)})
			return err
		}, "POST", "/sections", `{"section_id":9,"section":{"reusable":true}}`, record},
		{"sections update", func(n *Nusii) error {
			_, err := n.Sections.Update(ctx, 7, SectionParams{Title: String("Scope of work")})
			return err
		}, "PUT", "/sections/7", `{"section":{"title":"Scope of work"}}`, record},
		{"sections delete", func(n *Nusii) error { _, err := n.Sections.Delete(ctx, 7); return err }, "DELETE", "/sections/7", "", record},

		{"line items list", func(n *Nusii) error { _, err := n.LineItems.List(ctx, 3); return err }, "GET", "/sections/3/line_items", "", many(nil)},
		{"line items get", func(n *Nusii) error { _, err := n.LineItems.Get(ctx, 7); return err }, "GET", "/line_items/7", "", record},
		{"line items create", func(n *Nusii) error {
			_, err := n.LineItems.Create(ctx, 3, LineItemParams{Name: String("Dev"), CostType: CostTypePer, PerType: PerTypeHour, Quantity: Float(2.5), Amount: Int(7500)})
			return err
		}, "POST", "/sections/3/line_items", `{"line_item":{"name":"Dev","amount":7500,"cost_type":"per","per_type":"hour","quantity":2.5}}`, record},
		{"line items update", func(n *Nusii) error {
			_, err := n.LineItems.Update(ctx, 7, LineItemParams{MaximumAmount: Null[int64]()})
			return err
		}, "PUT", "/line_items/7", `{"line_item":{"maximum_amount":null}}`, record},
		{"line items delete", func(n *Nusii) error { _, err := n.LineItems.Delete(ctx, 7); return err }, "DELETE", "/line_items/7", "", record},

		{"taxations list", func(n *Nusii) error { _, err := n.Taxations.List(ctx, 3); return err }, "GET", "/proposals/3/taxations", "", many(nil)},
		{"taxations create", func(n *Nusii) error {
			_, err := n.Taxations.Create(ctx, 3, TaxationParams{Name: String("VAT"), Percentage: Float(21)})
			return err
		}, "POST", "/proposals/3/taxations", `{"name":"VAT","percentage":21}`, record},
		{"taxations update", func(n *Nusii) error {
			_, err := n.Taxations.Update(ctx, 3, 7, TaxationParams{Percentage: Float(-9.5)})
			return err
		}, "PUT", "/proposals/3/taxations/7", `{"percentage":-9.5}`, record},
		{"taxations delete", func(n *Nusii) error { _, err := n.Taxations.Delete(ctx, 3, 7); return err }, "DELETE", "/proposals/3/taxations/7", "", record},

		{"templates get", func(n *Nusii) error { _, err := n.Templates.Get(ctx, 7); return err }, "GET", "/templates/7", "", record},
		{"proposal activities get", func(n *Nusii) error { _, err := n.ProposalActivities.Get(ctx, 7); return err }, "GET", "/proposal_activities/7", "", record},

		{"webhook endpoints get", func(n *Nusii) error { _, err := n.WebhookEndpoints.Get(ctx, 7); return err }, "GET", "/webhook_endpoints/7", "", record},
		{"webhook endpoints create", func(n *Nusii) error {
			_, err := n.WebhookEndpoints.Create(ctx, WebhookEndpointCreateParams{TargetURL: "https://example.com/hook", Events: []WebhookEventName{WebhookEventProposalAccepted}})
			return err
		}, "POST", "/webhook_endpoints", `{"webhook_endpoint":{"target_url":"https://example.com/hook","events":["proposal_accepted"]}}`, record},
		{"webhook endpoints delete", func(n *Nusii) error { _, err := n.WebhookEndpoints.Delete(ctx, 7); return err }, "DELETE", "/webhook_endpoints/7", "", record},

		{"themes", func(n *Nusii) error { _, err := n.Themes.List(ctx); return err }, "GET", "/themes", "", []any{}},
		{"currencies", func(n *Nusii) error { _, err := n.Currencies.List(ctx); return err }, "GET", "/currencies", "", []any{}},
		{"locales", func(n *Nusii) error { _, err := n.Locales.List(ctx); return err }, "GET", "/locales", "", []any{}},
		{"pdf page sizes", func(n *Nusii) error { _, err := n.PDFPageSizes.List(ctx); return err }, "GET", "/pdf_page_sizes", "", []any{}},

		{"one page crm", func(n *Nusii) error {
			_, err := n.Integrations.CreateOnePageCRMInstallation(ctx, OnePageCRMInstallationParams{APIKey: "k", APISecret: "s"})
			return err
		}, "POST", "/integrations/one_page_crm_installations", `{"one_page_crm_installation":{"api_key":"k","api_secret":"s"}}`, record},
		{"less annoying crm", func(n *Nusii) error {
			_, err := n.Integrations.CreateLessAnnoyingCRMInstallation(ctx, LessAnnoyingCRMInstallationParams{Name: String("LACRM"), Token: "t", UserCode: "u", UserID: String("42")})
			return err
		}, "POST", "/integrations/less_annoying_crm_installations", `{"less_annoying_crm_installation":{"name":"LACRM","token":"t","user_code":"u","user_id":"42"}}`, record},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n, api := newMockClient(t, []http.HandlerFunc{ok(c.answer)})
			if err := c.call(n); err != nil {
				t.Fatal(err)
			}
			r := api.request(0)
			if r.Method != c.method || r.Path != "/api/v2"+c.path {
				t.Errorf("got %s %s, want %s /api/v2%s", r.Method, r.Path, c.method, c.path)
			}
			if c.body == "" {
				if r.Body != nil {
					t.Errorf("got body %v, want none", r.Body)
				}
				return
			}
			var want map[string]any
			if err := json.Unmarshal([]byte(c.body), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(r.Body, want) {
				got, _ := json.Marshal(r.Body)
				t.Errorf("got body %s, want %s", got, c.body)
			}
		})
	}
}

func TestListFilters(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name string
		call func(n *Nusii) error
		path string
		want map[string][]string
	}{
		{"clients", func(n *Nusii) error {
			_, err := n.Clients.List(ctx, &ClientListParams{Query: "acme", Email: "jane@example.com", Emails: []string{"a@example.com", "b@example.com"}})
			return err
		}, "/clients", map[string][]string{"page": {"1"}, "query": {"acme"}, "email": {"jane@example.com"}, "emails": {"a@example.com,b@example.com"}}},
		{"proposals", func(n *Nusii) error {
			_, err := n.Proposals.List(ctx, &ProposalListParams{
				PerPage:                      50,
				Status:                       ProposalStatusDraft,
				Statuses:                     []ProposalStatus{ProposalStatusAccepted, ProposalStatusRejected},
				Archived:                     true,
				Query:                        "web",
				RecipientEmail:               "jane@example.com",
				RecipientEmails:              []string{"a@example.com", "b@example.com"},
				NoActivityClientViewProposal: true,
				CreatedAtAfter:               "2026-01-01",
				CreatedAtBefore:              "2026-01-31",
				SentAtAfter:                  "2026-02-01",
				SentAtBefore:                 "2026-02-28",
				ExpiresAtAfter:               "2026-03-01",
				ExpiresAtBefore:              "2026-03-31",
				AcceptedAtAfter:              "2026-04-01",
				AcceptedAtBefore:             "2026-04-30",
			})
			return err
		}, "/proposals", map[string][]string{
			"page": {"1"}, "per": {"50"}, "status": {"draft"}, "statuses[]": {"accepted", "rejected"}, "archived": {"true"},
			"query": {"web"}, "recipient_email": {"jane@example.com"}, "recipient_emails": {"a@example.com,b@example.com"},
			"no_activity_client_view_proposal": {"true"},
			"created_at_after":                 {"2026-01-01"}, "created_at_before": {"2026-01-31"},
			"sent_at_after": {"2026-02-01"}, "sent_at_before": {"2026-02-28"},
			"expires_at_after": {"2026-03-01"}, "expires_at_before": {"2026-03-31"},
			"accepted_at_after": {"2026-04-01"}, "accepted_at_before": {"2026-04-30"},
		}},
		{"proposals without archived", func(n *Nusii) error {
			_, err := n.Proposals.List(ctx, &ProposalListParams{Archived: false})
			return err
		}, "/proposals", map[string][]string{"page": {"1"}}},
		{"sections", func(n *Nusii) error {
			_, err := n.Sections.List(ctx, &SectionListParams{ProposalID: 3, TemplateID: 4})
			return err
		}, "/sections", map[string][]string{"page": {"1"}, "proposal_id": {"3"}, "template_id": {"4"}}},
		{"templates", func(n *Nusii) error {
			_, err := n.Templates.List(ctx, &TemplateListParams{PublicTemplates: true})
			return err
		}, "/templates", map[string][]string{"page": {"1"}, "public_templates": {"true"}}},
		{"proposal activities", func(n *Nusii) error {
			_, err := n.ProposalActivities.List(ctx, &ProposalActivityListParams{Page: 3, ProposalID: 5, ClientID: 6})
			return err
		}, "/proposal_activities", map[string][]string{"page": {"3"}, "proposal_id": {"5"}, "client_id": {"6"}}},
		{"users", func(n *Nusii) error {
			_, err := n.Users.List(ctx, &ListParams{Page: 2, PerPage: 10})
			return err
		}, "/users", map[string][]string{"page": {"2"}, "per": {"10"}}},
		{"webhook endpoints", func(n *Nusii) error {
			_, err := n.WebhookEndpoints.List(ctx, nil)
			return err
		}, "/webhook_endpoints", map[string][]string{"page": {"1"}}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n, api := newMockClient(t, []http.HandlerFunc{ok(many(nil))})
			if err := c.call(n); err != nil {
				t.Fatal(err)
			}
			r := api.request(0)
			if r.Path != "/api/v2"+c.path {
				t.Errorf("got path %s", r.Path)
			}
			if !reflect.DeepEqual(map[string][]string(r.Query), c.want) {
				t.Errorf("got query %v, want %v", r.Query, c.want)
			}
		})
	}
}

func TestDecodesRecords(t *testing.T) {
	section := map[string]any{
		"data": map[string]any{
			"id":   "12",
			"type": "sections",
			"attributes": map[string]any{
				"title":          "Costs",
				"section_type":   "cost",
				"proposal_id":    3,
				"template_id":    nil,
				"created_at":     "2026-10-02T10:36:28.172Z",
				"total_in_cents": 240000,
			},
			"relationships": map[string]any{
				"line_items": map[string]any{"data": []any{
					map[string]any{"id": "5", "type": "line_items"},
					map[string]any{"id": "6", "type": "line_items"},
				}},
			},
		},
		"included": []any{
			map[string]any{"id": "6", "type": "line_items", "attributes": map[string]any{"name": "Hours", "cost_type": "per", "per_type": "hour", "quantity": "10.0", "amount": 7500}},
			map[string]any{"id": "5", "type": "line_items", "attributes": map[string]any{"name": "Design", "cost_type": "fixed", "per_type": nil, "quantity": nil, "amount": 150000}},
		},
	}
	n, _ := newMockClient(t, []http.HandlerFunc{ok(section)})

	got, err := n.Sections.Get(context.Background(), 12)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 12 || *got.Title != "Costs" || got.SectionType != SectionTypeCost || *got.ProposalID != 3 || got.TemplateID != nil {
		t.Errorf("got section %+v", got)
	}
	if got.CreatedAt.Year() != 2026 || got.CreatedAt.Nanosecond() != 172_000_000 {
		t.Errorf("got created at %v", got.CreatedAt)
	}
	if len(got.LineItems) != 2 {
		t.Fatalf("got %d line items", len(got.LineItems))
	}
	design, hours := got.LineItems[0], got.LineItems[1]
	if design.ID != 5 || design.PerType != nil || design.Quantity != nil || design.Amount != 150000 {
		t.Errorf("got design %+v", design)
	}
	if hours.ID != 6 || *hours.PerType != PerTypeHour || *hours.Quantity != "10.0" {
		t.Errorf("got hours %+v", hours)
	}
}

func TestRelationshipsMissingFromIncludedAreLeftOut(t *testing.T) {
	document := one(map[string]any{
		"id": "1", "type": "sections", "attributes": map[string]any{"section_type": "text"},
		"relationships": map[string]any{"line_items": map[string]any{"data": []any{map[string]any{"id": "5", "type": "line_items"}}}},
	})
	n, _ := newMockClient(t, []http.HandlerFunc{ok(document)})

	got, err := n.Sections.Get(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.LineItems != nil {
		t.Errorf("got line items %+v", got.LineItems)
	}
}

func TestDecodesPlainResponses(t *testing.T) {
	n, _ := newMockClient(t, []http.HandlerFunc{
		ok([]any{map[string]any{"id": "USD", "iso_code": "USD", "name": "United States Dollar $"}}),
		ok(map[string]any{"status": "pending", "sent_at": "2026-10-12T10:00:00.000Z", "sender_id": 4, "sender_name": "Jo"}),
	})
	ctx := context.Background()

	currencies, err := n.Currencies.List(ctx)
	if err != nil || !reflect.DeepEqual(currencies, []Currency{{ID: "USD", ISOCode: "USD", Name: "United States Dollar $"}}) {
		t.Errorf("got %+v %v", currencies, err)
	}

	result, err := n.Proposals.Send(ctx, 1, ProposalSendParams{Email: String("jane@example.com")})
	if err != nil || result.Status != ProposalStatusPending || result.SenderID != 4 || result.SentAt.Hour() != 10 {
		t.Errorf("got %+v %v", result, err)
	}
}

func TestDecodesAccountCountry(t *testing.T) {
	n, _ := newMockClient(t, []http.HandlerFunc{ok(one(res("accounts", 1, map[string]any{
		"name": "Acme", "subdomain": nil, "country": map[string]any{"alpha2": "NL"}, "country_name": "Netherlands",
	})))})

	account, err := n.Account.Me(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if account.Subdomain != nil || *account.CountryName != "Netherlands" || string(account.Country) != `{"alpha2":"NL"}` {
		t.Errorf("got %+v", account)
	}
}
