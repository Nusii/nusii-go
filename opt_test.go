package nusii

import (
	"encoding/json"
	"testing"
)

func TestOpt(t *testing.T) {
	type params struct {
		Omitted Opt[string]  `json:"omitted,omitzero"`
		Null    Opt[int64]   `json:"null,omitzero"`
		Text    Opt[string]  `json:"text,omitzero"`
		Empty   Opt[string]  `json:"empty,omitzero"`
		Number  Opt[float64] `json:"number,omitzero"`
		False   Opt[bool]    `json:"false,omitzero"`
	}
	encoded, err := json.Marshal(params{
		Null:   Null[int64](),
		Text:   String("hi"),
		Empty:  String(""),
		Number: Float(2.5),
		False:  Bool(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"null":null,"text":"hi","empty":"","number":2.5,"false":false}`
	if string(encoded) != want {
		t.Errorf("got %s, want %s", encoded, want)
	}

	if value, ok := Int(3).Get(); !ok || value != 3 {
		t.Errorf("Int(3).Get() = %v %v", value, ok)
	}
	if _, ok := Null[string]().Get(); ok {
		t.Error("Null is set")
	}
	var omitted Opt[bool]
	if !omitted.IsZero() || omitted.IsNull() || !Null[bool]().IsNull() || Bool(false).IsZero() {
		t.Error("unexpected Opt states")
	}
}

func TestDecimal(t *testing.T) {
	var values struct {
		String Decimal  `json:"string"`
		Number Decimal  `json:"number"`
		Null   *Decimal `json:"null"`
	}
	if err := json.Unmarshal([]byte(`{"string":"21.0","number":2.5,"null":null}`), &values); err != nil {
		t.Fatal(err)
	}
	if values.String != "21.0" || values.Number != "2.5" || values.Null != nil {
		t.Errorf("got %+v", values)
	}
	if f, err := values.String.Float64(); err != nil || f != 21 {
		t.Errorf("Float64() = %v %v", f, err)
	}
	if err := json.Unmarshal([]byte(`{"string":true}`), &values); err == nil {
		t.Error("decoded a boolean as a decimal")
	}
}
