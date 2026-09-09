package json

import (
	stdjson "encoding/json"
	"reflect"
	"testing"
)

func TestEndecEncode(t *testing.T) {
	encoded, err := (Endec{}).Encode(map[string]any{
		"name":  "Zenith",
		"count": 2,
	})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var decoded map[string]any
	if err := stdjson.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("encoded data is invalid JSON: %v", err)
	}
	if decoded["name"] != "Zenith" || decoded["count"] != float64(2) {
		t.Fatalf("encoded data = %#v", decoded)
	}
}

func TestEndecDecode(t *testing.T) {
	want := map[string]any{"name": "Zenith", "enabled": true}
	var got map[string]any

	if err := (Endec{}).Decode([]byte(`{"name":"Zenith","enabled":true}`), &got); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Decode() = %#v, want %#v", got, want)
	}
}

func TestEndecRoundTrip(t *testing.T) {
	input := map[string]any{
		"name":   "Zenith",
		"nested": map[string]any{"version": 1},
	}

	encoded, err := (Endec{}).Encode(input)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var decoded map[string]any
	if err := (Endec{}).Decode(encoded, &decoded); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if decoded["name"] != input["name"] {
		t.Fatalf("round trip name = %v, want %v", decoded["name"], input["name"])
	}
}

func TestEndecErrors(t *testing.T) {
	if _, err := (Endec{}).Encode(map[string]any{"channel": make(chan int)}); err == nil {
		t.Error("Encode() expected an error for an unsupported value")
	}
	if err := (Endec{}).Decode([]byte(`{"invalid"}`), new(map[string]any)); err == nil {
		t.Error("Decode() expected an error for invalid JSON")
	}
	if err := (Endec{}).Decode([]byte(`{}`), nil); err == nil {
		t.Error("Decode() expected an error for a nil destination")
	}
}
