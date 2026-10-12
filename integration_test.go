//go:build integration

// Runs against a real Nusii account: go test -tags integration ./...
//
// Reads credentials from .env (see .env.example). Skipped without them.
// Read-only by default. Set NUSII_INTEGRATION_WRITE=true to also create
// clients, proposals, sections, line items, taxes and a webhook endpoint,
// which are deleted afterwards. Set NUSII_INTEGRATION_SEND_TO=you@example.com
// to also send the test proposal to that address (this sends a real email).

package nusii_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nusii/nusii-go"
)

func TestMain(m *testing.M) {
	loadEnvFile(".env")
	os.Exit(m.Run())
}

// loadEnvFile sets the variables of a .env file that are not set yet.
func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		if _, set := os.LookupEnv(name); !set {
			os.Setenv(name, strings.Trim(strings.TrimSpace(value), `"'`))
		}
	}
}

func liveClient(t *testing.T) *nusii.Nusii {
	t.Helper()
	if os.Getenv("NUSII_API_KEY") == "" && os.Getenv("NUSII_ACCESS_TOKEN") == "" {
		t.Skip("set NUSII_API_KEY in .env to run the integration tests")
	}
	n, err := nusii.New()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func writableClient(t *testing.T) *nusii.Nusii {
	t.Helper()
	n := liveClient(t)
	if os.Getenv("NUSII_INTEGRATION_WRITE") != "true" {
		t.Skip("set NUSII_INTEGRATION_WRITE=true to run the integration tests that write")
	}
	return n
}

// failure is how must reports an error to stopOnFailure.
type failure struct{ error }

// must returns value, or stops the test with err. Each test using it starts
// with defer stopOnFailure(t).
func must[T any](value T, err error) T {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		panic(failure{fmt.Errorf("%s:%d: %w", filepath.Base(file), line, err)})
	}
	return value
}

func stopOnFailure(t *testing.T) {
	if r := recover(); r != nil {
		if f, ok := r.(failure); ok {
			t.Fatal(f.error)
		}
		panic(r)
	}
}

func TestLiveReads(t *testing.T) {
	n := liveClient(t)
	ctx := context.Background()

	t.Run("account", func(t *testing.T) {
		defer stopOnFailure(t)
		account := must(n.Account.Me(ctx))
		if account.ID == 0 || account.Name == "" {
			t.Errorf("unexpected account %+v", account)
		}
	})

	t.Run("paginated lists", func(t *testing.T) {
		defer stopOnFailure(t)
		checkPage(t, "clients", must(n.Clients.List(ctx, &nusii.ClientListParams{PerPage: 2})))
		checkPage(t, "proposals", must(n.Proposals.List(ctx, &nusii.ProposalListParams{PerPage: 2})))
		checkPage(t, "templates", must(n.Templates.List(ctx, &nusii.TemplateListParams{PerPage: 2})))
		checkPage(t, "users", must(n.Users.List(ctx, &nusii.ListParams{PerPage: 2})))
		checkPage(t, "proposal activities", must(n.ProposalActivities.List(ctx, &nusii.ProposalActivityListParams{PerPage: 2})))
		checkPage(t, "webhook endpoints", must(n.WebhookEndpoints.List(ctx, &nusii.ListParams{PerPage: 2})))
		checkPage(t, "sections", must(n.Sections.List(ctx, &nusii.SectionListParams{PerPage: 2})))
	})

	t.Run("every proposal decodes", func(t *testing.T) {
		defer stopOnFailure(t)
		count := 0
		for _, err := range n.Proposals.ListAll(ctx, &nusii.ProposalListParams{PerPage: 100}) {
			if err != nil {
				t.Fatal(err)
			}
			count++
		}
		for _, err := range n.Proposals.ListAll(ctx, &nusii.ProposalListParams{PerPage: 100, Archived: true}) {
			if err != nil {
				t.Fatal(err)
			}
		}
		t.Logf("%d proposals", count)
	})

	t.Run("every activity decodes", func(t *testing.T) {
		defer stopOnFailure(t)
		for _, err := range n.ProposalActivities.ListAll(ctx, &nusii.ProposalActivityListParams{PerPage: 100}) {
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("public templates", func(t *testing.T) {
		defer stopOnFailure(t)
		page := must(n.Templates.List(ctx, &nusii.TemplateListParams{PublicTemplates: true, PerPage: 1}))
		for _, template := range page.Data {
			got := must(n.Templates.Get(ctx, template.ID))
			if got.Name != template.Name {
				t.Errorf("template %d: got name %q, want %q", template.ID, got.Name, template.Name)
			}
		}
	})

	t.Run("reference data", func(t *testing.T) {
		defer stopOnFailure(t)
		if themes := must(n.Themes.List(ctx)); len(themes) == 0 || themes[0].ID == "" {
			t.Errorf("unexpected themes %+v", themes)
		}
		if currencies := must(n.Currencies.List(ctx)); len(currencies) == 0 || currencies[0].ISOCode == "" {
			t.Errorf("unexpected currencies %+v", currencies)
		}
		if locales := must(n.Locales.List(ctx)); len(locales) == 0 || locales[0].Code == "" {
			t.Errorf("unexpected locales %+v", locales)
		}
		if sizes := must(n.PDFPageSizes.List(ctx)); len(sizes) == 0 || sizes[0].ID == "" {
			t.Errorf("unexpected PDF page sizes %+v", sizes)
		}
	})

	t.Run("not found", func(t *testing.T) {
		defer stopOnFailure(t)
		_, err := n.Clients.Get(ctx, 0)
		if !errors.Is(err, nusii.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("invalid key", func(t *testing.T) {
		defer stopOnFailure(t)
		invalid := must(nusii.New(nusii.WithAPIKey("invalid-key"), nusii.WithBaseURL(n.BaseURL())))
		_, err := invalid.Account.Me(ctx)
		if !errors.Is(err, nusii.ErrAuthentication) {
			t.Errorf("got %v, want ErrAuthentication", err)
		}
	})
}

func checkPage[T any](t *testing.T, name string, page *nusii.Page[T]) {
	t.Helper()
	if page.Meta.TotalCount < len(page.Data) {
		t.Errorf("%s: total count %d is below the %d items on the page", name, page.Meta.TotalCount, len(page.Data))
	}
	if len(page.Data) > 2 {
		t.Errorf("%s: got %d items, want at most 2", name, len(page.Data))
	}
}

func TestLiveWebhookEndpoints(t *testing.T) {
	n := writableClient(t)
	ctx := context.Background()
	defer stopOnFailure(t)
	stamp := time.Now().UnixMilli()

	endpoint := must(n.WebhookEndpoints.Create(ctx, nusii.WebhookEndpointCreateParams{
		TargetURL: fmt.Sprintf("https://example.com/nusii-sdk-test/%d", stamp),
		Events:    []nusii.WebhookEventName{nusii.WebhookEventProposalAccepted},
	}))
	t.Cleanup(func() { n.WebhookEndpoints.Delete(context.Background(), endpoint.ID) })

	if !slices.Equal(endpoint.Events, []nusii.WebhookEventName{nusii.WebhookEventProposalAccepted}) {
		t.Errorf("got events %v", endpoint.Events)
	}
	got := must(n.WebhookEndpoints.Get(ctx, endpoint.ID))
	if got.TargetURL != endpoint.TargetURL {
		t.Errorf("got target URL %q, want %q", got.TargetURL, endpoint.TargetURL)
	}

	var ids []int64
	for item, err := range n.WebhookEndpoints.ListAll(ctx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, item.ID)
	}
	if !slices.Contains(ids, endpoint.ID) {
		t.Errorf("endpoint %d not in %v", endpoint.ID, ids)
	}

	deleted := must(n.WebhookEndpoints.Delete(ctx, endpoint.ID))
	if deleted.ID != endpoint.ID {
		t.Errorf("deleted %d, want %d", deleted.ID, endpoint.ID)
	}
	if _, err := n.WebhookEndpoints.Get(ctx, endpoint.ID); !errors.Is(err, nusii.ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestLiveProposalLifecycle(t *testing.T) {
	n := writableClient(t)
	ctx := context.Background()
	defer stopOnFailure(t)
	stamp := time.Now().UnixMilli()
	title := fmt.Sprintf("SDK test %d", stamp)

	// Clients
	email := fmt.Sprintf("sdk-test+%d@example.com", stamp)
	client := must(n.Clients.Create(ctx, nusii.ClientCreateParams{Name: "SDK", Surname: nusii.String("Test"), Email: email}))
	t.Cleanup(func() { n.Clients.Delete(context.Background(), client.ID) })
	if client.FullName != "SDK Test" {
		t.Errorf("got full name %q", client.FullName)
	}

	_, err := n.Clients.Create(ctx, nusii.ClientCreateParams{Name: "", Email: "invalid"})
	var apiErr *nusii.APIError
	if !errors.Is(err, nusii.ErrUnprocessableEntity) || !errors.As(err, &apiErr) || len(apiErr.Errors) == 0 {
		t.Errorf("got %v, want a validation error", err)
	}

	must(n.Clients.Update(ctx, client.ID, nusii.ClientUpdateParams{Business: nusii.String("SDK Test Inc")}))
	found := must(n.Clients.List(ctx, &nusii.ClientListParams{Email: email}))
	if len(found.Data) != 1 || found.Data[0].ID != client.ID || *found.Data[0].Business != "SDK Test Inc" {
		t.Errorf("unexpected clients %+v", found.Data)
	}
	cleared := must(n.Clients.Update(ctx, client.ID, nusii.ClientUpdateParams{Business: nusii.Null[string]()}))
	if cleared.Business != nil && *cleared.Business != "" {
		t.Errorf("business not cleared: %q", *cleared.Business)
	}

	// Proposals
	proposal := must(n.Proposals.Create(ctx, nusii.ProposalCreateParams{
		Title:     nusii.String(title),
		ClientID:  nusii.Int(client.ID),
		ExpiresAt: nusii.String("2099-12-31"),
	}))
	t.Cleanup(func() { n.Proposals.Delete(context.Background(), proposal.ID) })
	if proposal.Status != nusii.ProposalStatusDraft || *proposal.ClientEmail != email || *proposal.ExpiresAt != "2099-12-31" {
		t.Errorf("unexpected proposal %+v", proposal)
	}

	renamed := must(n.Proposals.Update(ctx, proposal.ID, nusii.ProposalUpdateParams{
		Title:     nusii.String(title + " v2"),
		ExpiresAt: nusii.Null[string](),
	}))
	if *renamed.Title != title+" v2" || renamed.ExpiresAt != nil {
		t.Errorf("got title %q and expiry %v", *renamed.Title, renamed.ExpiresAt)
	}

	var statusChange struct {
		Proposal struct {
			Status string `json:"status"`
		} `json:"proposal"`
	}
	statusChange.Proposal.Status = "accepted"
	err = n.Request(ctx, nusii.RequestConfig{Method: "PUT", Path: fmt.Sprintf("/proposals/%d", proposal.ID), Body: statusChange}, nil)
	if !errors.As(err, &apiErr) || apiErr.Code != "status_not_writable" {
		t.Errorf("got %v, want status_not_writable", err)
	}

	search := must(n.Proposals.List(ctx, &nusii.ProposalListParams{Query: title}))
	if len(search.Data) != 1 || search.Data[0].ID != proposal.ID {
		t.Errorf("search found %d proposals", len(search.Data))
	}
	drafts := must(n.Proposals.List(ctx, &nusii.ProposalListParams{
		Query:    title,
		Statuses: []nusii.ProposalStatus{nusii.ProposalStatusDraft, nusii.ProposalStatusPending},
	}))
	if len(drafts.Data) != 1 {
		t.Errorf("statuses filter found %d proposals", len(drafts.Data))
	}

	// A cost section with line items
	section := must(n.Sections.Create(ctx, nusii.SectionParams{
		ProposalID:  nusii.Int(proposal.ID),
		Title:       nusii.String("Costs"),
		SectionType: nusii.SectionTypeCost,
	}))
	t.Cleanup(func() { n.Sections.Delete(context.Background(), section.ID) })
	if *section.Name != "Costs" {
		t.Errorf("got section name %q", *section.Name)
	}

	design := must(n.LineItems.Create(ctx, section.ID, nusii.LineItemParams{
		Name:     nusii.String("Design"),
		CostType: nusii.CostTypeFixed,
		Amount:   nusii.Int(150_000),
	}))
	hours := must(n.LineItems.Create(ctx, section.ID, nusii.LineItemParams{
		Name:     nusii.String("Development"),
		CostType: nusii.CostTypePer,
		PerType:  nusii.PerTypeHour,
		Quantity: nusii.Float(10),
		Amount:   nusii.Int(7_500),
	}))
	if hours.TotalInCents != 75_000 {
		t.Errorf("got total %d, want 75000", hours.TotalInCents)
	}

	updated := must(n.LineItems.Update(ctx, hours.ID, nusii.LineItemParams{Quantity: nusii.Float(12.5)}))
	if quantity, _ := updated.Quantity.Float64(); quantity != 12.5 || updated.TotalInCents != 93_750 {
		t.Errorf("got quantity %v and total %d", updated.Quantity, updated.TotalInCents)
	}
	if got := must(n.LineItems.Get(ctx, hours.ID)); *got.Name != "Development" {
		t.Errorf("got line item %q", *got.Name)
	}

	items := must(n.LineItems.List(ctx, section.ID))
	if len(items) != 2 || items[0].ID != design.ID || items[1].ID != hours.ID {
		t.Errorf("unexpected line items %+v", items)
	}

	withItems := must(n.Sections.Get(ctx, section.ID))
	if len(withItems.LineItems) != 2 || withItems.TotalInCents != 243_750 {
		t.Errorf("got %d line items and total %d", len(withItems.LineItems), withItems.TotalInCents)
	}
	listed := must(n.Sections.List(ctx, &nusii.SectionListParams{ProposalID: proposal.ID}))
	if len(listed.Data) != 1 || len(listed.Data[0].LineItems) != 2 {
		t.Errorf("unexpected sections %+v", listed.Data)
	}

	// Price ranges are behind a feature flag: either works or explains why not.
	rangeItem, err := n.LineItems.Create(ctx, section.ID, nusii.LineItemParams{
		Name: nusii.String("Range"), CostType: nusii.CostTypeRange, Amount: nusii.Int(1), MaximumAmount: nusii.Int(2),
	})
	if err != nil {
		if !errors.As(err, &apiErr) || apiErr.Code != "cost_type_not_available" {
			t.Errorf("got %v, want cost_type_not_available", err)
		}
	} else if rangeItem.MaximumAmountInCents == nil || *rangeItem.MaximumAmountInCents != 2 {
		t.Errorf("unexpected range %+v", rangeItem)
	}

	// Taxes
	vat := must(n.Taxations.Create(ctx, proposal.ID, nusii.TaxationParams{Name: nusii.String("VAT"), Percentage: nusii.Float(21)}))
	if percentage, _ := vat.Percentage.Float64(); percentage != 21 {
		t.Errorf("got percentage %v", vat.Percentage)
	}
	changed := must(n.Taxations.Update(ctx, proposal.ID, vat.ID, nusii.TaxationParams{Percentage: nusii.Float(9.5)}))
	if percentage, _ := changed.Percentage.Float64(); percentage != 9.5 {
		t.Errorf("got percentage %v", changed.Percentage)
	}
	if taxes := must(n.Taxations.List(ctx, proposal.ID)); len(taxes) != 1 || taxes[0].ID != vat.ID {
		t.Errorf("unexpected taxes %+v", taxes)
	}
	must(n.Taxations.Delete(ctx, proposal.ID, vat.ID))
	if taxes := must(n.Taxations.List(ctx, proposal.ID)); len(taxes) != 0 {
		t.Errorf("taxes left: %+v", taxes)
	}

	// Reusable sections: create in the content library, copy into the proposal
	reusable := must(n.Sections.Create(ctx, nusii.SectionParams{
		Title:       nusii.String(fmt.Sprintf("Reusable %d", stamp)),
		Body:        nusii.String("<p>About us</p>"),
		SectionType: nusii.SectionTypeText,
		Reusable:    nusii.Bool(true),
	}))
	t.Cleanup(func() { n.Sections.Delete(context.Background(), reusable.ID) })
	if reusable.ProposalID != nil {
		t.Errorf("reusable section attached to proposal %d", *reusable.ProposalID)
	}

	var library []int64
	for item, err := range n.Sections.ListAll(ctx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		library = append(library, item.ID)
	}
	if !slices.Contains(library, reusable.ID) {
		t.Errorf("reusable section %d not in the content library", reusable.ID)
	}

	copied := must(n.Sections.Copy(ctx, reusable.ID, nusii.SectionParams{ProposalID: nusii.Int(proposal.ID)}))
	t.Cleanup(func() { n.Sections.Delete(context.Background(), copied.ID) })
	if copied.ID == reusable.ID || copied.ProposalID == nil || *copied.ProposalID != proposal.ID || *copied.Title != *reusable.Title {
		t.Errorf("unexpected copy %+v", copied)
	}

	// Deleting twice
	must(n.Sections.Delete(ctx, copied.ID))
	if _, err := n.Sections.Delete(ctx, copied.ID); !errors.Is(err, nusii.ErrGone) {
		t.Errorf("got %v, want ErrGone", err)
	}

	// Activities of the proposal
	must(n.ProposalActivities.List(ctx, &nusii.ProposalActivityListParams{ProposalID: proposal.ID}))

	// Optional: really send it
	if sendTo := os.Getenv("NUSII_INTEGRATION_SEND_TO"); sendTo != "" {
		result := must(n.Proposals.Send(ctx, proposal.ID, nusii.ProposalSendParams{
			Recipients: []nusii.ProposalRecipient{{Name: "SDK Test", Email: sendTo}},
			Subject:    nusii.String("Nusii Go SDK test"),
			Message:    nusii.String("<p>Sent by the nusii-go integration tests.</p>"),
		}))
		if result.Status != nusii.ProposalStatusPending || result.SentAt.IsZero() {
			t.Errorf("unexpected send result %+v", result)
		}
	}

	// Archive and unarchive
	archived := must(n.Proposals.Archive(ctx, proposal.ID))
	if archived.ArchivedAt == nil {
		t.Error("proposal not archived")
	}
	inArchive := must(n.Proposals.List(ctx, &nusii.ProposalListParams{Archived: true, Query: title}))
	if len(inArchive.Data) != 1 || inArchive.Data[0].ID != proposal.ID {
		t.Errorf("archive search found %d proposals", len(inArchive.Data))
	}
	if active := must(n.Proposals.List(ctx, &nusii.ProposalListParams{Query: title})); len(active.Data) != 0 {
		t.Errorf("archived proposal still listed as active")
	}
	unarchived := must(n.Proposals.Update(ctx, proposal.ID, nusii.ProposalUpdateParams{ArchivedAt: nusii.Null[string]()}))
	if unarchived.ArchivedAt != nil {
		t.Errorf("proposal still archived at %v", unarchived.ArchivedAt)
	}
}

func TestLiveProposalFromTemplate(t *testing.T) {
	n := writableClient(t)
	ctx := context.Background()
	defer stopOnFailure(t)

	templates := must(n.Templates.List(ctx, &nusii.TemplateListParams{PerPage: 1}))
	if len(templates.Data) == 0 {
		t.Skip("the account has no templates")
	}
	template := templates.Data[0]

	proposal := must(n.Proposals.Create(ctx, nusii.ProposalCreateParams{
		Title:      nusii.String(fmt.Sprintf("SDK template test %d", time.Now().UnixMilli())),
		TemplateID: nusii.Int(template.ID),
	}))
	t.Cleanup(func() { n.Proposals.Delete(context.Background(), proposal.ID) })

	templateSections := must(n.Sections.List(ctx, &nusii.SectionListParams{TemplateID: template.ID}))
	proposalSections := must(n.Sections.List(ctx, &nusii.SectionListParams{ProposalID: proposal.ID}))
	if proposalSections.Meta.TotalCount != templateSections.Meta.TotalCount {
		t.Errorf("proposal has %d sections, template %d", proposalSections.Meta.TotalCount, templateSections.Meta.TotalCount)
	}

	must(n.Proposals.Delete(ctx, proposal.ID))
	if _, err := n.Proposals.Get(ctx, proposal.ID); !errors.Is(err, nusii.ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}
