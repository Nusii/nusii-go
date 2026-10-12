package nusii

// Payloads Nusii POSTs to webhook endpoints. Each event carries the affected
// record under its own key plus the event_name:
//
//	{"event_name": "proposal_accepted", "proposal": {"id": 1, ...}}

import (
	"encoding/json"
	"fmt"
	"time"
)

// WebhookProposal is the proposal of a proposal_* webhook event.
type WebhookProposal struct {
	ID       int64          `json:"id"`
	Title    *string        `json:"title"`
	Status   ProposalStatus `json:"status"`
	PublicID string         `json:"public_id"`
	// PublicURL is the link the client opens to view the proposal.
	PublicURL      *string    `json:"public_url"`
	DocumentNumber *string    `json:"document_number"`
	PreparedByID   *int64     `json:"prepared_by_id"`
	PreparedByName *string    `json:"prepared_by_name"`
	ClientID       *int64     `json:"client_id"`
	SenderID       *int64     `json:"sender_id"`
	SenderName     *string    `json:"sender_name"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	SentAt         *time.Time `json:"sent_at"`
	AcceptedAt     *time.Time `json:"accepted_at"`
	// ExpiresAt is a date (YYYY-MM-DD).
	ExpiresAt              *string    `json:"expires_at"`
	Report                 bool       `json:"report"`
	Currency               string     `json:"currency"`
	PaymentPercentage      *int       `json:"payment_percentage"`
	PaidAmountFormatted    *string    `json:"paid_amount_formatted"`
	PaidAt                 *time.Time `json:"paid_at"`
	TotalFormatted         string     `json:"total_formatted"`
	AcceptedTotalInCents   *int64     `json:"accepted_total_in_cents"`
	AcceptedTotalFormatted *string    `json:"accepted_total_formatted"`
	AccountName            string     `json:"account_name"`
	AccountWeb             *string    `json:"account_web"`
	AccountEmail           string     `json:"account_email"`
	AccountTelephone       *string    `json:"account_telephone"`
	AccountAddress         *string    `json:"account_address"`
	AccountAddressState    *string    `json:"account_address_state"`
	ClientRole             *string    `json:"client_role"`
	ClientName             *string    `json:"client_name"`
	ClientEmail            *string    `json:"client_email"`
	ClientSurname          *string    `json:"client_surname"`
	ClientFullName         *string    `json:"client_full_name"`
	ClientBusiness         *string    `json:"client_business"`
	ClientTelephone        *string    `json:"client_telephone"`
	ClientLocale           *string    `json:"client_locale"`
	ClientAddress          *string    `json:"client_address"`
	ClientCity             *string    `json:"client_city"`
	ClientPostcode         *string    `json:"client_postcode"`
	ClientCountry          *string    `json:"client_country"`
	ClientState            *string    `json:"client_state"`
	ClientWeb              *string    `json:"client_web"`
	ClientNotes            *string    `json:"client_notes"`
}

// WebhookProposalActivity is the activity of a proposal_activity_* webhook event.
type WebhookProposalActivity struct {
	ID           int64                `json:"id"`
	ActivityType ProposalActivityType `json:"activity_type"`
	IPAddress    *string              `json:"ip_address"`
	ViewedAt     time.Time            `json:"viewed_at"`
	// ProposalLink is the link to the proposal in the Nusii app.
	ProposalLink      *string         `json:"proposal_link"`
	ProposalTitle     *string         `json:"proposal_title"`
	ProposalCreatedAt *time.Time      `json:"proposal_created_at"`
	ProposalSentAt    *time.Time      `json:"proposal_sent_at"`
	ProposalStatus    *ProposalStatus `json:"proposal_status"`
	ProposalPublicID  *string         `json:"proposal_public_id"`
	// ProposalExpiresAt is a date (YYYY-MM-DD).
	ProposalExpiresAt *string `json:"proposal_expires_at"`
	ClientName        *string `json:"client_name"`
	ClientEmail       *string `json:"client_email"`
	ClientSurname     *string `json:"client_surname"`
	ClientFullName    *string `json:"client_full_name"`
	ClientBusiness    *string `json:"client_business"`
	ClientTelephone   *string `json:"client_telephone"`
	ClientLocale      *string `json:"client_locale"`
	ClientAddress     *string `json:"client_address"`
	ClientPostcode    *string `json:"client_postcode"`
	ClientCountry     *string `json:"client_country"`
	ClientCity        *string `json:"client_city"`
	ClientState       *string `json:"client_state"`
	ClientWeb         *string `json:"client_web"`
}

// WebhookEvent is an event Nusii delivered to a webhook endpoint. Switch on
// EventName: proposal_* events set Proposal, client_* events set Client and
// proposal_activity_* events set ProposalActivity.
type WebhookEvent struct {
	EventName        WebhookEventName         `json:"event_name"`
	Proposal         *WebhookProposal         `json:"proposal,omitempty"`
	Client           *Client                  `json:"client,omitempty"`
	ProposalActivity *WebhookProposalActivity `json:"proposal_activity,omitempty"`
}

// ParseWebhookEvent parses the body of a webhook delivery.
//
// Nusii does not sign webhook deliveries, so treat the endpoint URL as a
// secret (e.g. include a long random token in its path) and fetch the record
// from the API when you need to be certain of its current state.
func ParseWebhookEvent(payload []byte) (*WebhookEvent, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		return nil, fmt.Errorf("%w: the body is not a JSON object", ErrInvalidWebhookPayload)
	}
	var name string
	if json.Unmarshal(fields["event_name"], &name) != nil || name == "" {
		return nil, fmt.Errorf("%w: missing event_name", ErrInvalidWebhookPayload)
	}

	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidWebhookPayload, err)
	}
	return &event, nil
}
