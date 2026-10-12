package nusii

// Flattens JSON:API documents into the plain records of this package:
//
//	{"data": {"id": "1", "type": "clients", "attributes": {"name": "Jane"}}}
//
// becomes
//
//	{"id": 1, "name": "Jane"}
//
// Relationships whose resources are all present in "included" (a section's
// line items) are attached as arrays under the relationship name.

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type resourceIdentifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type jsonAPIResource struct {
	ID            string                     `json:"id"`
	Type          string                     `json:"type"`
	Attributes    map[string]json.RawMessage `json:"attributes"`
	Relationships map[string]struct {
		Data json.RawMessage `json:"data"`
	} `json:"relationships"`
}

type jsonAPIDocument struct {
	Data     json.RawMessage   `json:"data"`
	Included []jsonAPIResource `json:"included"`
	Meta     *struct {
		CurrentPage *int `json:"current_page"`
		NextPage    *int `json:"next_page"`
		PrevPage    *int `json:"prev_page"`
		TotalPages  *int `json:"total_pages"`
		TotalCount  *int `json:"total_count"`
	} `json:"meta"`
}

func decodeOne[T any](body []byte) (*T, error) {
	document, included, err := parseDocument(body)
	if err != nil {
		return nil, err
	}
	var resource jsonAPIResource
	if err := json.Unmarshal(document.Data, &resource); err != nil {
		return nil, fmt.Errorf("nusii: decoding the response: %w", err)
	}
	var record T
	if err := decodeResource(resource, included, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func decodeMany[T any](body []byte) ([]T, *jsonAPIDocument, error) {
	document, included, err := parseDocument(body)
	if err != nil {
		return nil, nil, err
	}
	var resources []jsonAPIResource
	if err := json.Unmarshal(document.Data, &resources); err != nil {
		return nil, nil, fmt.Errorf("nusii: decoding the response: %w", err)
	}
	records := make([]T, len(resources))
	for i, resource := range resources {
		if err := decodeResource(resource, included, &records[i]); err != nil {
			return nil, nil, err
		}
	}
	return records, document, nil
}

func parseDocument(body []byte) (*jsonAPIDocument, map[string]jsonAPIResource, error) {
	var document jsonAPIDocument
	if err := json.Unmarshal(body, &document); err != nil {
		return nil, nil, fmt.Errorf("nusii: decoding the response: %w", err)
	}
	if len(document.Data) == 0 {
		return nil, nil, fmt.Errorf("nusii: decoding the response: no data in %q", truncate(body, 200))
	}
	included := make(map[string]jsonAPIResource, len(document.Included))
	for _, resource := range document.Included {
		included[resource.Type+":"+resource.ID] = resource
	}
	return &document, included, nil
}

func decodeResource(resource jsonAPIResource, included map[string]jsonAPIResource, target any) error {
	flat, err := json.Marshal(flatten(resource, included))
	if err != nil {
		return fmt.Errorf("nusii: decoding the response: %w", err)
	}
	if err := json.Unmarshal(flat, target); err != nil {
		return fmt.Errorf("nusii: decoding a %s record: %w", resource.Type, err)
	}
	return nil
}

func flatten(resource jsonAPIResource, included map[string]jsonAPIResource) map[string]json.RawMessage {
	result := make(map[string]json.RawMessage, len(resource.Attributes)+1)
	for name, value := range resource.Attributes {
		result[name] = value
	}
	result["id"] = idJSON(resource.ID)

	for name, relationship := range resource.Relationships {
		var identifiers []resourceIdentifier
		if json.Unmarshal(relationship.Data, &identifiers) != nil {
			continue // a to-one relationship or null
		}
		related := make([]map[string]json.RawMessage, 0, len(identifiers))
		for _, identifier := range identifiers {
			item, ok := included[identifier.Type+":"+identifier.ID]
			if !ok {
				break
			}
			related = append(related, flatten(item, nil))
		}
		if len(related) == len(identifiers) {
			if encoded, err := json.Marshal(related); err == nil {
				result[name] = encoded
			}
		}
	}
	return result
}

// idJSON keeps numeric ids numbers, so they decode into int64 fields.
func idJSON(id string) json.RawMessage {
	if _, err := strconv.ParseUint(id, 10, 64); err == nil {
		return json.RawMessage(id)
	}
	quoted, _ := json.Marshal(id)
	return quoted
}

func truncate(body []byte, length int) string {
	if len(body) <= length {
		return string(body)
	}
	return string(body[:length]) + "..."
}
