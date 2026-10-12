# Nusii for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/nusii/nusii-go.svg)](https://pkg.go.dev/github.com/nusii/nusii-go)
[![CI](https://github.com/Nusii/nusii-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Nusii/nusii-go/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

The official Go client for the [Nusii](https://nusii.com) proposal software API.

Create clients, build and send proposals, track when they are viewed and accepted, and react to webhooks, all from Go services, CLIs and serverless functions.

- **Fully typed**: every request and response is a Go struct, with constants for statuses, cost types and webhook events.
- **Zero dependencies**: only the standard library.
- **Batteries included**: automatic pagination with range-over-func iterators, retries with backoff, timeouts, context cancellation and errors you can match with `errors.Is`.
- **Safe for concurrent use**: share one client across goroutines.

Full API reference: [developer.nusii.com](https://developer.nusii.com) · Package documentation: [pkg.go.dev](https://pkg.go.dev/github.com/nusii/nusii-go)

## Table of contents

- [Installation](#installation)
- [Getting an API key](#getting-an-api-key)
- [Quick start](#quick-start)
- [Usage](#usage)
  - [Account](#account)
  - [Clients](#clients)
  - [Proposals](#proposals)
  - [Sections](#sections)
  - [Line items](#line-items)
  - [Taxes and discounts](#taxes-and-discounts)
  - [Templates](#templates)
  - [Proposal activities](#proposal-activities)
  - [Users](#users)
  - [Webhooks](#webhooks)
  - [Themes, currencies, locales and PDF page sizes](#themes-currencies-locales-and-pdf-page-sizes)
  - [CRM integrations](#crm-integrations)
- [Optional parameters](#optional-parameters)
- [Pagination](#pagination)
- [Error handling](#error-handling)
- [Retries, timeouts and cancellation](#retries-timeouts-and-cancellation)
- [Configuration](#configuration)
- [OAuth](#oauth)
- [Calling any endpoint](#calling-any-endpoint)
- [Requirements](#requirements)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Installation

```bash
go get github.com/nusii/nusii-go
```

```go
import "github.com/nusii/nusii-go"
```

The package name is `nusii`.

## Getting an API key

1. Log in to [Nusii](https://app.nusii.com).
2. Go to **Settings → API** ([app.nusii.com/settings/api](https://app.nusii.com/settings/api)).
3. Under **Personal Tokens**, create a token and copy it.

Treat the key like a password: it gives full access to your Nusii account. Keep it out of your code and out of git.

The usual way is an environment variable, which the client reads automatically:

```bash
export NUSII_API_KEY=your-api-key
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nusii/nusii-go"
)

func main() {
	ctx := context.Background()

	// Reads NUSII_API_KEY from the environment.
	n, err := nusii.New()
	if err != nil {
		log.Fatal(err)
	}

	// Or pass the key explicitly:
	// n, err := nusii.New(nusii.WithAPIKey("your-api-key"))

	account, err := n.Account.Me(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connected to %s\n", account.Name)

	client, err := n.Clients.Create(ctx, nusii.ClientCreateParams{
		Name:     "Jane",
		Surname:  nusii.String("Doe"),
		Email:    "jane@example.com",
		Business: nusii.String("Acme Inc"),
	})
	if err != nil {
		log.Fatal(err)
	}

	proposal, err := n.Proposals.Create(ctx, nusii.ProposalCreateParams{
		Title:    nusii.String("Website redesign"),
		ClientID: nusii.Int(client.ID),
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = n.Proposals.Send(ctx, proposal.ID, nusii.ProposalSendParams{
		Recipients: []nusii.ProposalRecipient{{Name: "Jane Doe", Email: "jane@example.com"}},
		Subject:    nusii.String("Your proposal for the website redesign"),
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

Field names are the API attribute names in Go style (`client_id` is `ClientID`, `expires_at` is `ExpiresAt`), and the JSON tags match the API exactly, so everything in the [API documentation](https://developer.nusii.com) applies one to one. Nullable attributes are pointers. Ids are `int64`.

## Usage

Every method takes a `context.Context` first and accepts [options](#retries-timeouts-and-cancellation) as its last arguments. The examples leave out error handling.

### Account

```go
account, err := n.Account.Me(ctx)

account.Name        // "Acme Corp"
account.Currency    // "USD"
*account.Subdomain  // "acme" (nil when not set)
```

### Clients

```go
// List (25 per page by default, newest first)
page, err := n.Clients.List(ctx, &nusii.ClientListParams{Page: 1, PerPage: 50})
page.Data            // []nusii.Client
page.Meta.TotalCount // 132

// Search
n.Clients.List(ctx, &nusii.ClientListParams{Query: "acme"})
n.Clients.List(ctx, &nusii.ClientListParams{Email: "jane@example.com"})
n.Clients.List(ctx, &nusii.ClientListParams{Emails: []string{"jane@example.com", "john@example.com"}})

// Get, create, update, delete
client, err := n.Clients.Get(ctx, 123)

created, err := n.Clients.Create(ctx, nusii.ClientCreateParams{
	Name:      "Jane",             // required
	Email:     "jane@example.com", // required
	Surname:   nusii.String("Doe"),
	Business:  nusii.String("Acme Inc"),
	Telephone: nusii.String("+1 555 1234"),
	Address:   nusii.String("123 Main St"),
	City:      nusii.String("New York"),
	Postcode:  nusii.String("10001"),
	State:     nusii.String("NY"),
	Country:   nusii.String("US"),
	Web:       nusii.String("https://acme.com"),
	Currency:  nusii.String("USD"),
	Locale:    nusii.String("en"),
})

n.Clients.Update(ctx, 123, nusii.ClientUpdateParams{Business: nusii.String("Acme Corporation")})
n.Clients.Delete(ctx, 123)
```

### Proposals

```go
// List with filters
n.Proposals.List(ctx, &nusii.ProposalListParams{Status: nusii.ProposalStatusPending})
n.Proposals.List(ctx, &nusii.ProposalListParams{
	Statuses: []nusii.ProposalStatus{nusii.ProposalStatusAccepted, nusii.ProposalStatusRejected},
})
n.Proposals.List(ctx, &nusii.ProposalListParams{Archived: true})  // only archived ones
n.Proposals.List(ctx, &nusii.ProposalListParams{Query: "website"}) // title, number, client...
n.Proposals.List(ctx, &nusii.ProposalListParams{RecipientEmail: "jane@example.com"})
n.Proposals.List(ctx, &nusii.ProposalListParams{NoActivityClientViewProposal: true}) // never viewed
n.Proposals.List(ctx, &nusii.ProposalListParams{
	SentAtAfter:  "2026-01-01", // dates are YYYY-MM-DD, inclusive
	SentAtBefore: "2026-03-31",
})

proposal, err := n.Proposals.Get(ctx, 456)
proposal.Status // nusii.ProposalStatusDraft, Pending, Accepted, Rejected or Clarification
proposal.PDFURL

// Create from scratch...
n.Proposals.Create(ctx, nusii.ProposalCreateParams{
	Title:     nusii.String("Website redesign"),
	ClientID:  nusii.Int(123),
	ExpiresAt: nusii.String("2026-12-31"),
	Theme:     nusii.String("clean"),
})

// ...or from a template, copying its sections and line items
n.Proposals.Create(ctx, nusii.ProposalCreateParams{
	Title:      nusii.String("Website redesign"),
	ClientID:   nusii.Int(123),
	TemplateID: nusii.Int(42),
})

n.Proposals.Update(ctx, 456, nusii.ProposalUpdateParams{Title: nusii.String("Website redesign v2")})
n.Proposals.Archive(ctx, 456)
n.Proposals.Update(ctx, 456, nusii.ProposalUpdateParams{ArchivedAt: nusii.Null[string]()}) // unarchive
n.Proposals.Delete(ctx, 456)
```

Send a proposal by email to one or more recipients:

```go
result, err := n.Proposals.Send(ctx, 456, nusii.ProposalSendParams{
	Recipients: []nusii.ProposalRecipient{
		{Name: "Jane Doe", Email: "jane@example.com"},
		{Name: "John Doe", Email: "john@example.com", EligibleToSign: nusii.Bool(false)},
	},
	CC:          nusii.String("manager@example.com"),
	BCC:         nusii.String("archive@example.com"),
	Subject:     nusii.String("Your proposal"),
	Message:     nusii.String("<p>Hi Jane, here is the proposal we discussed.</p>"),
	SenderEmail: nusii.String("sales@yourcompany.com"), // send as a team member (default: account owner)
})

result.Status // nusii.ProposalStatusPending
result.SentAt
```

A proposal's status follows what happens to it (sent, accepted, rejected), so it cannot be changed with `Update`.

### Sections

Sections are the blocks of a proposal or template: text sections and cost sections (which contain line items).

```go
// Sections of a proposal or template, with their line items
page, err := n.Sections.List(ctx, &nusii.SectionListParams{ProposalID: 456})
page.Data[0].LineItems // []nusii.LineItem on cost sections

n.Sections.List(ctx, &nusii.SectionListParams{TemplateID: 42})

// Without a proposal or template: the reusable sections of your content library
n.Sections.List(ctx, nil)

section, err := n.Sections.Create(ctx, nusii.SectionParams{
	ProposalID:  nusii.Int(456),
	Title:       nusii.String("Project scope"),
	Body:        nusii.String("<p>What we will deliver.</p>"),
	SectionType: nusii.SectionTypeText, // or nusii.SectionTypeCost
	Position:    nusii.Int(1),
})

// Copy a reusable section (with its line items) into a proposal
n.Sections.Copy(ctx, reusableSectionID, nusii.SectionParams{ProposalID: nusii.Int(456), Position: nusii.Int(2)})

n.Sections.Update(ctx, section.ID, nusii.SectionParams{Title: nusii.String("Scope of work")})
n.Sections.Delete(ctx, section.ID)
```

### Line items

Line items belong to cost sections. Amounts are in cents.

```go
items, err := n.LineItems.List(ctx, sectionID)

// A fixed price
n.LineItems.Create(ctx, sectionID, nusii.LineItemParams{
	Name:     nusii.String("Logo design"),
	CostType: nusii.CostTypeFixed,
	Amount:   nusii.Int(150000), // $1,500.00
})

// Per unit: 10 hours at $75
n.LineItems.Create(ctx, sectionID, nusii.LineItemParams{
	Name:     nusii.String("Development"),
	CostType: nusii.CostTypePer,
	PerType:  nusii.PerTypeHour,
	Quantity: nusii.Float(10),
	Amount:   nusii.Int(7500),
})

// Recurring
n.LineItems.Create(ctx, sectionID, nusii.LineItemParams{
	Name:          nusii.String("Hosting"),
	CostType:      nusii.CostTypeRecurring,
	RecurringType: nusii.RecurringTypeMonthly,
	Amount:        nusii.Int(2500),
})

// A price range: "$5,000 – $8,000" (requires price ranges on your account)
rangeItem, err := n.LineItems.Create(ctx, sectionID, nusii.LineItemParams{
	Name:          nusii.String("Discovery & UX research"),
	CostType:      nusii.CostTypeRange,
	Amount:        nusii.Int(500000),
	MaximumAmount: nusii.Int(800000),
})
*rangeItem.MaximumAmountFormatted // "$8,000.00"

// Let the client pick optional items
n.LineItems.Create(ctx, sectionID, nusii.LineItemParams{
	Name:       nusii.String("Extra revision"),
	ChoiceType: nusii.ChoiceTypeCheckbox,
})

n.LineItems.Get(ctx, itemID)
n.LineItems.Update(ctx, itemID, nusii.LineItemParams{Quantity: nusii.Float(12)})
n.LineItems.Delete(ctx, itemID)
```

Quantities come back as a `nusii.Decimal` (a string such as `"12.0"`, so no precision is lost). Call `Float64()` to compute with it.

### Taxes and discounts

```go
n.Taxations.Create(ctx, proposalID, nusii.TaxationParams{Name: nusii.String("VAT"), Percentage: nusii.Float(21)})
n.Taxations.Create(ctx, proposalID, nusii.TaxationParams{Name: nusii.String("Loyalty discount"), Percentage: nusii.Float(-10)})

taxes, err := n.Taxations.List(ctx, proposalID)
n.Taxations.Update(ctx, proposalID, taxID, nusii.TaxationParams{Percentage: nusii.Float(19)})
n.Taxations.Delete(ctx, proposalID, taxID)
```

### Templates

```go
page, err := n.Templates.List(ctx, nil)
template, err := n.Templates.Get(ctx, 42)

// Nusii's public template gallery
n.Templates.List(ctx, &nusii.TemplateListParams{PublicTemplates: true})
```

### Proposal activities

Every view, send, acceptance, email open and bounce, newest first.

```go
n.ProposalActivities.List(ctx, nil)
n.ProposalActivities.List(ctx, &nusii.ProposalActivityListParams{ProposalID: 456})
n.ProposalActivities.List(ctx, &nusii.ProposalActivityListParams{ClientID: 123})

activity, err := n.ProposalActivities.Get(ctx, 789)
activity.ActivityType // nusii.ActivityClientViewProposal, nusii.ActivityClientAcceptedProposal...
```

### Users

The team members of your account.

```go
page, err := n.Users.List(ctx, nil)
```

### Webhooks

Ask Nusii to POST events to your server:

```go
endpoint, err := n.WebhookEndpoints.Create(ctx, nusii.WebhookEndpointCreateParams{
	TargetURL: "https://example.com/webhooks/nusii/a-long-random-secret",
	Events: []nusii.WebhookEventName{
		nusii.WebhookEventProposalAccepted,
		nusii.WebhookEventProposalRejected,
		nusii.WebhookEventProposalActivityClientViewedProposal,
	},
})

n.WebhookEndpoints.List(ctx, nil)
n.WebhookEndpoints.Get(ctx, endpoint.ID)
n.WebhookEndpoints.Delete(ctx, endpoint.ID)
```

Then receive them. `ParseWebhookEvent` returns the event with the record that matches its `EventName`:

```go
http.HandleFunc("POST /webhooks/nusii/{secret}", func(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("secret") != os.Getenv("NUSII_WEBHOOK_SECRET") {
		http.NotFound(w, r)
		return
	}
	body, _ := io.ReadAll(r.Body)
	event, err := nusii.ParseWebhookEvent(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch event.EventName {
	case nusii.WebhookEventProposalAccepted:
		log.Printf("%s accepted for %s", *event.Proposal.Title, *event.Proposal.AcceptedTotalFormatted)
	case nusii.WebhookEventProposalActivityClientViewedProposal:
		log.Printf("%s is reading your proposal", *event.ProposalActivity.ClientEmail)
	case nusii.WebhookEventClientCreated:
		log.Printf("New client: %s", event.Client.Email)
	}

	w.WriteHeader(http.StatusNoContent)
})
```

`proposal_*` events set `event.Proposal`, `client_*` events set `event.Client` and `proposal_activity_*` events set `event.ProposalActivity`.

Nusii does not sign webhook deliveries, so put a long random secret in the endpoint URL and check it, and fetch the record from the API when you need to be sure of its current state. Responding with HTTP `410 Gone` removes the endpoint.

| Event | When |
| --- | --- |
| `proposal_created` | A proposal was created |
| `proposal_updated` | A proposal was updated |
| `proposal_destroyed` | A proposal was deleted |
| `proposal_sent` | A proposal was sent |
| `proposal_accepted` | The client accepted a proposal |
| `proposal_rejected` | The client rejected a proposal |
| `client_created` | A client was created |
| `client_updated` | A client was updated |
| `client_destroyed` | A client was deleted |
| `proposal_activity_client_viewed` | A client viewed something |
| `proposal_activity_client_viewed_proposal` | A client viewed a proposal |

The full list is `nusii.WebhookEvents`.

### Themes, currencies, locales and PDF page sizes

The allowed values for `theme`, `currency`, `locale` and `pdf_page_size`:

```go
n.Themes.List(ctx)       // [{ID: "clean", Name: "Modern Theme"}, {ID: "classic", ...}]
n.Currencies.List(ctx)   // [{ID: "USD", ISOCode: "USD", Name: "United States Dollar $"}, ...]
n.Locales.List(ctx)      // [{ID: "en", Code: "en", Name: "English"}, ...]
n.PDFPageSizes.List(ctx) // [{ID: "A4", Name: "A4"}, {ID: "US-Letter", ...}]
```

### CRM integrations

```go
n.Integrations.CreateOnePageCRMInstallation(ctx, nusii.OnePageCRMInstallationParams{APIKey: "...", APISecret: "..."})
n.Integrations.CreateLessAnnoyingCRMInstallation(ctx, nusii.LessAnnoyingCRMInstallationParams{Token: "...", UserCode: "..."})
```

## Optional parameters

Optional request parameters have the type `nusii.Opt`. A zero `Opt` is left out of the request, so only what you set is sent. Set a value with `nusii.String`, `nusii.Int`, `nusii.Float` or `nusii.Bool`, and clear one with `nusii.Null`:

```go
n.Proposals.Update(ctx, 456, nusii.ProposalUpdateParams{
	Title:     nusii.String("Website redesign"), // set
	ExpiresAt: nusii.Null[string](),             // clear: the proposal no longer expires
	// Theme is left out and keeps its value
})
```

Required parameters, such as a new client's `Name` and `Email`, are plain fields. Enum parameters such as `CostType` and `SectionType` are plain fields too: their empty value is left out.

## Pagination

`List` methods return one `*nusii.Page` at a time:

```go
page, err := n.Proposals.List(ctx, &nusii.ProposalListParams{PerPage: 50})

page.Data // []nusii.Proposal
page.Meta // {CurrentPage: 1, NextPage: 2, PrevPage: nil, TotalPages: 4, TotalCount: 180}

if page.HasNextPage() {
	next, err := page.NextPage(ctx)
}
```

To walk through everything, range over `ListAll`. It fetches the next page only when the loop reaches it, and stops when you `break`:

```go
for proposal, err := range n.Proposals.ListAll(ctx, &nusii.ProposalListParams{Status: nusii.ProposalStatusPending}) {
	if err != nil {
		return err
	}
	fmt.Println(*proposal.Title)
}
```

`page.All(ctx)` does the same from a page you already have.

## Error handling

Failed requests return an error you can match with `errors.Is`, and inspect with `errors.As`:

```go
_, err := n.Clients.Create(ctx, nusii.ClientCreateParams{Name: "", Email: "not-an-email"})

var apiErr *nusii.APIError
errors.As(err, &apiErr) // set for every HTTP error

switch {
case errors.Is(err, nusii.ErrUnprocessableEntity):
	apiErr.Message // "name can't be blank; email is invalid"
	apiErr.Errors  // [{Source: {Pointer: "/data/attributes/name"}, Detail: "can't be blank"}, ...]
case errors.Is(err, nusii.ErrNotFound):
	// ...
case errors.Is(err, nusii.ErrRateLimit):
	wait, _ := apiErr.RetryAfter()
case errors.Is(err, nusii.ErrConnection):
	// no response: network down, DNS failure, timeout...
}
```

| Error | When |
| --- | --- |
| `ErrBadRequest` | 400: malformed request |
| `ErrAuthentication` | 401: invalid or missing API key |
| `ErrPaymentRequired` | 402: plan limit reached, e.g. active proposals |
| `ErrForbidden` | 403: not allowed, e.g. missing OAuth scope or unconfirmed sender email |
| `ErrNotFound` | 404: record does not exist |
| `ErrMethodNotAllowed` | 405 |
| `ErrNotAcceptable` | 406 |
| `ErrGone` | 410: record already deleted |
| `ErrUnprocessableEntity` | 422: validation failed, or the proposal is locked |
| `ErrRateLimit` | 429: too many requests |
| `ErrServer` | 5xx |
| `ErrServiceUnavailable` | 503 (also matches `ErrServer`) |
| `ErrConnection` | No response: network down, DNS failure... |
| `ErrTimeout` | The request took longer than the timeout (also matches `ErrConnection`) |

HTTP errors are `*nusii.APIError` values with the `StatusCode`, `Header`, raw `Body`, field `Errors`, and a `Code` when the API sends one (for example `proposal_locked`). Connection errors are `*nusii.ConnectionError` values that unwrap to the underlying network error.

## Retries, timeouts and cancellation

The Nusii API allows 100 requests per 30 seconds. The client handles the limit for you: a rate limited request is retried after the `Retry-After` delay the API asks for. Connection errors and 5xx responses are retried with exponential backoff, but only for reads (GET), so a proposal is never sent twice. By default a request is retried twice.

```go
// For every request
n, err := nusii.New(nusii.WithMaxRetries(5), nusii.WithTimeout(10*time.Second))

// For one request
n.Proposals.List(ctx, nil, nusii.WithMaxRetries(0), nusii.WithTimeout(5*time.Second))

// Cancel with the context
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
n.Proposals.List(ctx, nil)
```

The timeout applies to each attempt. The context covers the whole call, retries included.

## Configuration

```go
n, err := nusii.New(
	nusii.WithAPIKey("your-api-key"),              // default: $NUSII_API_KEY
	nusii.WithAccessToken("oauth-token"),          // instead of an API key, default: $NUSII_ACCESS_TOKEN
	nusii.WithBaseURL("https://app.nusii.com"),    // default: $NUSII_BASE_URL or https://app.nusii.com
	nusii.WithTimeout(30*time.Second),             // per attempt
	nusii.WithMaxRetries(2),
	nusii.WithHeader("X-Request-Source", "my-app"), // added to every request
	nusii.WithHTTPClient(&http.Client{Transport: myTransport}), // e.g. for a proxy or tests
)
```

Every option also works on a single request, as the last argument of any method.

## OAuth

Apps that act on behalf of other Nusii accounts use OAuth 2.0 access tokens instead of API keys:

```go
n, err := nusii.New(nusii.WithAccessToken(userAccessToken))

// Or one client for every account, with the token per request:
proposals, err := n.Proposals.List(ctx, nil, nusii.WithAccessToken(userAccessToken))
```

Tokens with only the `read` scope can call GET endpoints; anything else returns an `ErrForbidden` error with `Code` `"insufficient_scope"`.

## Calling any endpoint

`Request` calls any API v2 endpoint with the same authentication, retries and error handling, and decodes the JSON body as is:

```go
var body map[string]any
err := n.Request(ctx, nusii.RequestConfig{
	Method: http.MethodGet,
	Path:   "/proposals",
	Query:  url.Values{"status": {"draft"}, "per": {"10"}},
}, &body)
```

## Requirements

Go 1.24 or newer.

## Development

```bash
git clone https://github.com/Nusii/nusii-go.git
cd nusii-go

go test ./...       # unit tests
go vet ./...
gofmt -l .          # lists unformatted files
```

Integration tests run against a real Nusii account. Copy `.env.example` to `.env`, add your API key, and run:

```bash
go test -tags integration ./...
```

They only read data unless you set `NUSII_INTEGRATION_WRITE=true`, in which case they create clients, proposals, sections, line items, taxes and a webhook endpoint, and delete them afterwards. To test against a local Nusii server, set `NUSII_BASE_URL=http://localhost:3000` in `.env`.

Releases are git tags, see [RELEASING.md](RELEASING.md).

## Contributing

1. Fork it (https://github.com/Nusii/nusii-go/fork)
2. Create your feature branch (`git checkout -b improving-something`)
3. Commit your changes and add tests (`go vet ./...` and `go test ./...` must pass)
4. Push to the branch (`git push origin improving-something`)
5. Create a new Pull Request

## License

Released under the [MIT License](LICENSE).

---

Made by [Nusii](https://nusii.com), proposal software that helps you win more clients.
