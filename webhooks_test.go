package nusii

import (
	"errors"
	"testing"
)

func TestParseWebhookEvent(t *testing.T) {
	proposal, err := ParseWebhookEvent([]byte(`{
		"event_name": "proposal_accepted",
		"proposal": {"id": 1, "title": "Redesign", "status": "accepted", "created_at": "2026-10-02T10:36:28.172Z",
			"expires_at": "2026-12-31", "accepted_total_formatted": "$2,400.00", "paid_at": null}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if proposal.EventName != WebhookEventProposalAccepted || proposal.Proposal == nil || proposal.Client != nil {
		t.Fatalf("got %+v", proposal)
	}
	if p := proposal.Proposal; p.ID != 1 || p.Status != ProposalStatusAccepted || *p.ExpiresAt != "2026-12-31" || *p.AcceptedTotalFormatted != "$2,400.00" {
		t.Errorf("got proposal %+v", p)
	}

	client, err := ParseWebhookEvent([]byte(`{"event_name": "client_created", "client": {"id": 2, "email": "jane@example.com"}}`))
	if err != nil || client.Client == nil || client.Client.Email != "jane@example.com" {
		t.Errorf("got %+v %v", client, err)
	}

	activity, err := ParseWebhookEvent([]byte(`{"event_name": "proposal_activity_client_viewed_proposal",
		"proposal_activity": {"id": 3, "activity_type": "client_view_proposal", "viewed_at": "2026-10-02T10:36:28Z", "client_email": "jane@example.com"}}`))
	if err != nil || activity.ProposalActivity == nil || activity.ProposalActivity.ActivityType != ActivityClientViewProposal {
		t.Errorf("got %+v %v", activity, err)
	}
}

func TestParseWebhookEventRejectsOtherPayloads(t *testing.T) {
	for _, payload := range []string{``, `not json`, `[]`, `{}`, `{"event_name": ""}`, `{"event_name": 5}`, `{"event_name": "proposal_created", "proposal": {"id": "abc"}}`} {
		if _, err := ParseWebhookEvent([]byte(payload)); !errors.Is(err, ErrInvalidWebhookPayload) {
			t.Errorf("%q: got %v, want ErrInvalidWebhookPayload", payload, err)
		}
	}
}
