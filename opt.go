package nusii

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Opt is an optional request parameter. Its zero value leaves the parameter
// out of the request; set it with String, Int, Float or Bool, or send null
// with Null:
//
//	n.Proposals.Update(ctx, id, nusii.ProposalUpdateParams{
//		Title:     nusii.String("Website redesign"),
//		ExpiresAt: nusii.Null[string](), // remove the expiry date
//	})
type Opt[T any] struct {
	value T
	state optState
}

type optState uint8

const (
	optOmitted optState = iota
	optNull
	optSet
)

// String returns a string parameter.
func String(value string) Opt[string] { return Opt[string]{value: value, state: optSet} }

// Int returns an integer parameter, for ids, amounts in cents and positions.
func Int(value int64) Opt[int64] { return Opt[int64]{value: value, state: optSet} }

// Float returns a decimal parameter, for quantities and percentages.
func Float(value float64) Opt[float64] { return Opt[float64]{value: value, state: optSet} }

// Bool returns a boolean parameter.
func Bool(value bool) Opt[bool] { return Opt[bool]{value: value, state: optSet} }

// Null returns a parameter that is sent as null, to clear a value.
func Null[T any]() Opt[T] { return Opt[T]{state: optNull} }

// Get returns the value and whether it is set. It returns false for an
// omitted or null parameter.
func (o Opt[T]) Get() (T, bool) { return o.value, o.state == optSet }

// IsNull reports whether the parameter is sent as null.
func (o Opt[T]) IsNull() bool { return o.state == optNull }

// IsZero reports whether the parameter is left out of the request.
func (o Opt[T]) IsZero() bool { return o.state == optOmitted }

// MarshalJSON encodes the value, or null.
func (o Opt[T]) MarshalJSON() ([]byte, error) {
	if o.state != optSet {
		return []byte("null"), nil
	}
	return json.Marshal(o.value)
}

// Decimal is a decimal number the API sends as a string to keep its
// precision, e.g. a quantity of "2.5" or a tax percentage of "21.0".
type Decimal string

// Float64 returns the number as a float64.
func (d Decimal) Float64() (float64, error) {
	return strconv.ParseFloat(string(d), 64)
}

// String returns the number as the API sent it.
func (d Decimal) String() string { return string(d) }

// UnmarshalJSON accepts the number as a JSON string or a JSON number.
func (d *Decimal) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "null" {
		*d = ""
		return nil
	}
	if strings.HasPrefix(text, `"`) {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*d = Decimal(s)
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return err
	}
	*d = Decimal(number)
	return nil
}
