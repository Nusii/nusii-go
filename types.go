package nusii

// Records of the Nusii API v2. JSON tags match the API attribute names, so
// https://developer.nusii.com applies one to one: client_id is ClientID,
// expires_at is ExpiresAt. Nullable attributes are pointers.

import (
	"encoding/json"
	"time"
)

// Account is the Nusii account the credentials belong to.
type Account struct {
	ID           int64   `json:"id"`
	Email        string  `json:"email"`
	Name         string  `json:"name"`
	Subdomain    *string `json:"subdomain"`
	Web          *string `json:"web"`
	Currency     string  `json:"currency"`
	Address      *string `json:"address"`
	AddressState *string `json:"address_state"`
	Postcode     *string `json:"postcode"`
	City         *string `json:"city"`
	// Country is the country record as the API sends it. Use CountryName to display it.
	Country     json.RawMessage `json:"country"`
	CountryName *string         `json:"country_name"`
	Telephone   *string         `json:"telephone"`
}

// Client is a customer you send proposals to.
type Client struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	Name        *string   `json:"name"`
	Surname     *string   `json:"surname"`
	FullName    string    `json:"full_name"`
	Business    *string   `json:"business"`
	Web         *string   `json:"web"`
	Currency    string    `json:"currency"`
	Locale      *string   `json:"locale"`
	TimeZone    *string   `json:"time_zone"`
	PDFPageSize *string   `json:"pdf_page_size"`
	Notes       *string   `json:"notes"`
	ClientRole  *string   `json:"client_role"`
	Telephone   *string   `json:"telephone"`
	Address     *string   `json:"address"`
	City        *string   `json:"city"`
	Postcode    *string   `json:"postcode"`
	Country     *string   `json:"country"`
	State       *string   `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProposalStatus is the status of a proposal.
type ProposalStatus string

// Proposal statuses.
const (
	ProposalStatusDraft         ProposalStatus = "draft"
	ProposalStatusPending       ProposalStatus = "pending"
	ProposalStatusAccepted      ProposalStatus = "accepted"
	ProposalStatusRejected      ProposalStatus = "rejected"
	ProposalStatusClarification ProposalStatus = "clarification"
)

// ProposalStatuses lists every proposal status.
var ProposalStatuses = []ProposalStatus{
	ProposalStatusDraft,
	ProposalStatusPending,
	ProposalStatusAccepted,
	ProposalStatusRejected,
	ProposalStatusClarification,
}

// Proposal is a proposal document.
type Proposal struct {
	ID             int64          `json:"id"`
	Title          *string        `json:"title"`
	AccountID      int64          `json:"account_id"`
	Status         ProposalStatus `json:"status"`
	PublicID       string         `json:"public_id"`
	DocumentNumber *string        `json:"document_number"`
	PreparedByID   *int64         `json:"prepared_by_id"`
	ClientID       *int64         `json:"client_id"`
	SenderID       *int64         `json:"sender_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	SentAt         *time.Time     `json:"sent_at"`
	SignedAt       *time.Time     `json:"signed_at"`
	AcceptedAt     *time.Time     `json:"accepted_at"`
	ArchivedAt     *time.Time     `json:"archived_at"`
	// ExpiresAt is the date (YYYY-MM-DD) after which the proposal can no longer be accepted.
	ExpiresAt    *string    `json:"expires_at"`
	LastViewedAt *time.Time `json:"last_viewed_at"`
	// DisplayDate is the date (YYYY-MM-DD) shown on the proposal instead of the creation date.
	DisplayDate *string `json:"display_date"`
	// PresentDate is the date (YYYY-MM-DD) shown on the proposal.
	PresentDate              string  `json:"present_date"`
	PresentEmail             string  `json:"present_email"`
	DocumentSectionTitle     *string `json:"document_section_title"`
	ForceSignature           bool    `json:"force_signature"`
	Report                   bool    `json:"report"`
	ExcludeTotal             bool    `json:"exclude_total"`
	ExcludeTotalInPDF        bool    `json:"exclude_total_in_pdf"`
	TableOfContents          bool    `json:"table_of_contents"`
	Currency                 string  `json:"currency"`
	Theme                    string  `json:"theme"`
	CoverImageURL            *string `json:"cover_image_url"`
	EmptyCoverImage          bool    `json:"empty_cover_image"`
	CoverOpacity             int     `json:"cover_opacity"`
	CoverOverlayColor        *string `json:"cover_overlay_color"`
	CoverOverlayDynamicColor *string `json:"cover_overlay_dynamic_color"`
	SenderName               *string `json:"sender_name"`
	Source                   *string `json:"source"`
	AcceptedTotalInCents     *int64  `json:"accepted_total_in_cents"`
	AcceptedTotalFormatted   *string `json:"accepted_total_formatted"`
	PaymentPercentage        *int    `json:"payment_percentage"`
	// PDFURL is the public URL of the proposal PDF.
	PDFURL              string  `json:"pdf_url"`
	AccountName         string  `json:"account_name"`
	AccountWeb          *string `json:"account_web"`
	AccountEmail        string  `json:"account_email"`
	AccountTelephone    *string `json:"account_telephone"`
	AccountAddress      *string `json:"account_address"`
	AccountAddressState *string `json:"account_address_state"`
	ClientRole          *string `json:"client_role"`
	ClientLocale        *string `json:"client_locale"`
	ClientName          *string `json:"client_name"`
	ClientEmail         *string `json:"client_email"`
	ClientSurname       *string `json:"client_surname"`
	ClientFullName      *string `json:"client_full_name"`
	ClientBusiness      *string `json:"client_business"`
	ClientTelephone     *string `json:"client_telephone"`
	ClientAddress       *string `json:"client_address"`
	ClientCity          *string `json:"client_city"`
	ClientPostcode      *string `json:"client_postcode"`
	ClientCountry       *string `json:"client_country"`
	ClientState         *string `json:"client_state"`
	ClientWeb           *string `json:"client_web"`
	ClientNotes         *string `json:"client_notes"`
}

// ProposalSendResult is the outcome of sending a proposal.
type ProposalSendResult struct {
	Status     ProposalStatus `json:"status"`
	SentAt     time.Time      `json:"sent_at"`
	SenderID   int64          `json:"sender_id"`
	SenderName string         `json:"sender_name"`
}

// SectionType is the kind of a section.
type SectionType string

// Section types. Cost sections contain line items.
const (
	SectionTypeText SectionType = "text"
	SectionTypeCost SectionType = "cost"
)

// Section is a block of a proposal or template, or a reusable section of the
// content library.
type Section struct {
	ID         int64   `json:"id"`
	AccountID  *int64  `json:"account_id"`
	ProposalID *int64  `json:"proposal_id"`
	TemplateID *int64  `json:"template_id"`
	Currency   string  `json:"currency"`
	Title      *string `json:"title"`
	Name       *string `json:"name"`
	// Body is the HTML content of the section.
	Body         *string     `json:"body"`
	Position     *int        `json:"position"`
	Reusable     bool        `json:"reusable"`
	SectionType  SectionType `json:"section_type"`
	PageBreak    bool        `json:"page_break"`
	Optional     bool        `json:"optional"`
	Selected     bool        `json:"selected"`
	IncludeTotal bool        `json:"include_total"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	// TotalInCents is the total of the section without optional line items.
	TotalInCents   int64  `json:"total_in_cents"`
	TotalFormatted string `json:"total_formatted"`
	// MaximumTotalInCents is the high end of the total when the section
	// contains price ranges, otherwise nil.
	MaximumTotalInCents   *int64  `json:"maximum_total_in_cents"`
	MaximumTotalFormatted *string `json:"maximum_total_formatted"`
	// LineItems are the line items of a cost section. Always empty for text sections.
	LineItems []LineItem `json:"line_items,omitempty"`
}

// CostType is how a line item is priced.
type CostType string

// Cost types.
const (
	CostTypeFixed     CostType = "fixed"
	CostTypeRecurring CostType = "recurring"
	CostTypePer       CostType = "per"
	CostTypeRange     CostType = "range"
)

// RecurringType is how often a recurring line item is charged.
type RecurringType string

// Recurring types.
const (
	RecurringTypeYearly       RecurringType = "yearly"
	RecurringTypeSemiannually RecurringType = "semiannually"
	RecurringTypeTrimester    RecurringType = "trimester"
	RecurringTypeMonthly      RecurringType = "monthly"
	RecurringTypeQuadweekly   RecurringType = "quadweekly"
	RecurringTypeFortnightly  RecurringType = "fortnightly"
	RecurringTypeWeekly       RecurringType = "weekly"
	RecurringTypeDaily        RecurringType = "daily"
	RecurringTypeHourly       RecurringType = "hourly"
)

// PerType is the unit a "per" line item is priced in.
type PerType string

// Per types.
const (
	PerTypeYear        PerType = "year"
	PerTypeMonth       PerType = "month"
	PerTypeWeek        PerType = "week"
	PerTypeDay         PerType = "day"
	PerTypeHour        PerType = "hour"
	PerTypeItem        PerType = "item"
	PerTypeUnit        PerType = "unit"
	PerTypePerson      PerType = "person"
	PerTypeSession     PerType = "session"
	PerTypeTeam        PerType = "team"
	PerTypeVisit       PerType = "visit"
	PerTypeSquareMeter PerType = "square_meter"
	PerTypeMeter       PerType = "meter"
	PerTypeSong        PerType = "song"
	PerTypeCycle       PerType = "cycle"
	PerTypeKeynote     PerType = "keynote"
)

// ChoiceType controls whether the client can pick a line item.
type ChoiceType string

// Choice types: ChoiceTypeCheckbox lets the client pick any number of items,
// ChoiceTypeRadio exactly one.
const (
	ChoiceTypeNone     ChoiceType = "none"
	ChoiceTypeCheckbox ChoiceType = "checkbox"
	ChoiceTypeRadio    ChoiceType = "radio"
)

// LineItem is a priced item of a cost section. Amounts are in cents.
type LineItem struct {
	ID            int64          `json:"id"`
	SectionID     int64          `json:"section_id"`
	Name          *string        `json:"name"`
	Position      *int           `json:"position"`
	CostType      CostType       `json:"cost_type"`
	RecurringType *RecurringType `json:"recurring_type"`
	PerType       *PerType       `json:"per_type"`
	ChoiceType    ChoiceType     `json:"choice_type"`
	Quantity      *Decimal       `json:"quantity"`
	Currency      string         `json:"currency"`
	// Amount is the price in cents. For a range, the low end.
	Amount          int64  `json:"amount"`
	AmountInCents   int64  `json:"amount_in_cents"`
	AmountFormatted string `json:"amount_formatted"`
	// MaximumAmountInCents is the high end of a price range, nil unless CostType is CostTypeRange.
	MaximumAmountInCents   *int64    `json:"maximum_amount_in_cents"`
	MaximumAmountFormatted *string   `json:"maximum_amount_formatted"`
	TotalInCents           int64     `json:"total_in_cents"`
	TotalFormatted         string    `json:"total_formatted"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// Taxation is a tax, or with a negative percentage a discount, applied to a
// proposal's total.
type Taxation struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Percentage Decimal   `json:"percentage"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TemplateSummary is a template as listed by TemplatesService.List.
type TemplateSummary struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
	PublicTemplate bool      `json:"public_template"`
}

// Template is a template as returned by TemplatesService.Get.
type Template struct {
	TemplateSummary
	AccountID     *int64 `json:"account_id"`
	DummyTemplate bool   `json:"dummy_template"`
}

// ProposalActivityType is the kind of a proposal activity. The constants
// list the known values; new ones may appear.
type ProposalActivityType string

// Proposal activity types.
const (
	ActivityClientViewProposal             ProposalActivityType = "client_view_proposal"
	ActivityClientViewPDF                  ProposalActivityType = "client_view_pdf"
	ActivityUserCreateProposal             ProposalActivityType = "user_create_proposal"
	ActivityUserViewProposal               ProposalActivityType = "user_view_proposal"
	ActivityUserSendProposal               ProposalActivityType = "user_send_proposal"
	ActivityUserManualChangeStatusProposal ProposalActivityType = "user_manual_change_status_proposal"
	ActivityUserUnlockProposal             ProposalActivityType = "user_unlock_proposal"
	ActivityUserSignedProposal             ProposalActivityType = "user_signed_proposal"
	ActivityUserDuplicateProposal          ProposalActivityType = "user_duplicate_proposal"
	ActivityClientRejectedProposal         ProposalActivityType = "client_rejected_proposal"
	ActivityClientClarificationProposal    ProposalActivityType = "client_clarification_proposal"
	ActivityClientAcceptedProposal         ProposalActivityType = "client_accepted_proposal"
	ActivityClientPaidProposal             ProposalActivityType = "client_paid_proposal"
	ActivityClientPaidProposalError        ProposalActivityType = "client_paid_proposal_error"
	ActivityClientPaidInvoice              ProposalActivityType = "client_paid_invoice"
	ActivityClientEmailBounced             ProposalActivityType = "client_email_bounced"
	ActivityClientEmailOpened              ProposalActivityType = "client_email_opened"
	ActivityClientEmailSpamComplaint       ProposalActivityType = "client_email_spam_complaint"
	ActivitySampleProposalView             ProposalActivityType = "sample_proposal_view"
	ActivityAPICreatedProposal             ProposalActivityType = "api_created_proposal"
)

// ProposalActivity is something that happened to a proposal: a view, send,
// acceptance, email open or bounce.
type ProposalActivity struct {
	ID           int64                `json:"id"`
	ActivityType ProposalActivityType `json:"activity_type"`
	IPAddress    *string              `json:"ip_address"`
	// AdditionalFields holds extra details such as the email to, cc and
	// subject, or bounce information.
	AdditionalFields  map[string]any  `json:"additional_fields"`
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

// User is a team member of the account.
type User struct {
	ID    int64   `json:"id"`
	Email string  `json:"email"`
	Name  *string `json:"name"`
}

// WebhookEventName is the name of an event Nusii can POST to a webhook endpoint.
type WebhookEventName string

// Webhook events.
const (
	WebhookEventProposalCreated                      WebhookEventName = "proposal_created"
	WebhookEventProposalUpdated                      WebhookEventName = "proposal_updated"
	WebhookEventProposalDestroyed                    WebhookEventName = "proposal_destroyed"
	WebhookEventProposalAccepted                     WebhookEventName = "proposal_accepted"
	WebhookEventProposalRejected                     WebhookEventName = "proposal_rejected"
	WebhookEventProposalSent                         WebhookEventName = "proposal_sent"
	WebhookEventClientCreated                        WebhookEventName = "client_created"
	WebhookEventClientUpdated                        WebhookEventName = "client_updated"
	WebhookEventClientDestroyed                      WebhookEventName = "client_destroyed"
	WebhookEventProposalActivityClientViewed         WebhookEventName = "proposal_activity_client_viewed"
	WebhookEventProposalActivityClientViewedProposal WebhookEventName = "proposal_activity_client_viewed_proposal"
)

// WebhookEvents lists every webhook event.
var WebhookEvents = []WebhookEventName{
	WebhookEventProposalCreated,
	WebhookEventProposalUpdated,
	WebhookEventProposalDestroyed,
	WebhookEventProposalAccepted,
	WebhookEventProposalRejected,
	WebhookEventProposalSent,
	WebhookEventClientCreated,
	WebhookEventClientUpdated,
	WebhookEventClientDestroyed,
	WebhookEventProposalActivityClientViewed,
	WebhookEventProposalActivityClientViewedProposal,
}

// WebhookEndpoint is a URL subscribed to webhook events.
type WebhookEndpoint struct {
	ID        int64              `json:"id"`
	TargetURL string             `json:"target_url"`
	Events    []WebhookEventName `json:"events"`
}

// Theme is a proposal theme. Use its ID as a proposal's theme.
type Theme struct {
	// ID is the value for a proposal's theme, e.g. "clean".
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Currency is a supported currency. Use its ISOCode as a client's or proposal's currency.
type Currency struct {
	ID      string `json:"id"`
	ISOCode string `json:"iso_code"`
	Name    string `json:"name"`
}

// Locale is a supported language. Use its Code as a client's locale.
type Locale struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// PDFPageSize is a supported PDF page size. Use its ID as a client's pdf_page_size.
type PDFPageSize struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IntegrationInstallation is a connection to a CRM.
type IntegrationInstallation struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
