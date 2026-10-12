package nusii_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nusii/nusii-go"
)

func Example() {
	ctx := context.Background()

	// Reads NUSII_API_KEY from the environment.
	n, err := nusii.New()
	if err != nil {
		log.Fatal(err)
	}

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

func ExampleProposalsService_ListAll() {
	n, err := nusii.New()
	if err != nil {
		log.Fatal(err)
	}

	params := &nusii.ProposalListParams{Status: nusii.ProposalStatusPending}
	for proposal, err := range n.Proposals.ListAll(context.Background(), params) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(*proposal.Title)
	}
}

func ExampleAPIError() {
	n, err := nusii.New()
	if err != nil {
		log.Fatal(err)
	}

	_, err = n.Clients.Create(context.Background(), nusii.ClientCreateParams{Name: "", Email: "not-an-email"})

	var apiErr *nusii.APIError
	switch {
	case errors.Is(err, nusii.ErrUnprocessableEntity) && errors.As(err, &apiErr):
		fmt.Println(apiErr.Message) // name can't be blank; email is invalid
		for _, e := range apiErr.Errors {
			fmt.Println(e.Source.Pointer, e.Detail)
		}
	case errors.Is(err, nusii.ErrRateLimit):
		// Retried automatically; this means the limit was hit for too long.
	case err != nil:
		log.Fatal(err)
	}
}

func ExampleParseWebhookEvent() {
	http.HandleFunc("POST /webhooks/nusii/{secret}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("secret") != "a-long-random-secret" {
			http.NotFound(w, r)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		event, err := nusii.ParseWebhookEvent(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch event.EventName {
		case nusii.WebhookEventProposalAccepted:
			fmt.Printf("%s accepted for %s\n", *event.Proposal.Title, *event.Proposal.AcceptedTotalFormatted)
		case nusii.WebhookEventProposalActivityClientViewedProposal:
			fmt.Printf("%s is reading your proposal\n", *event.ProposalActivity.ClientEmail)
		case nusii.WebhookEventClientCreated:
			fmt.Printf("New client: %s\n", event.Client.Email)
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
