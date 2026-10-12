// Package nusii is the official Go client for the Nusii proposal software API.
//
// Create clients, build and send proposals, track when they are viewed and
// accepted, and react to webhooks:
//
//	n, err := nusii.New() // reads NUSII_API_KEY from the environment
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	client, err := n.Clients.Create(ctx, nusii.ClientCreateParams{
//		Name:     "Jane",
//		Surname:  nusii.String("Doe"),
//		Email:    "jane@example.com",
//		Business: nusii.String("Acme Inc"),
//	})
//
//	proposal, err := n.Proposals.Create(ctx, nusii.ProposalCreateParams{
//		Title:    nusii.String("Website redesign"),
//		ClientID: nusii.Int(client.ID),
//	})
//
//	_, err = n.Proposals.Send(ctx, proposal.ID, nusii.ProposalSendParams{
//		Recipients: []nusii.ProposalRecipient{{Name: "Jane Doe", Email: "jane@example.com"}},
//		Subject:    nusii.String("Your proposal for the website redesign"),
//	})
//
// Every method takes a context and accepts request options (see Option) as
// its last arguments. Optional request parameters are of type Opt: set them
// with String, Int, Float and Bool, or clear a value with Null.
//
// Failed requests return an *APIError for HTTP error statuses, matched with
// errors.Is against ErrNotFound, ErrUnprocessableEntity and the other status
// errors, or a *ConnectionError when no response arrived.
//
// Full API reference: https://developer.nusii.com
package nusii
