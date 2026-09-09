package dotenv

import (
	"strings"
	"testing"
)

func TestEndecEncode(t *testing.T) {
	codec := Endec{}
	input := map[string]any{
		"APP_NAME": "Zenith",
		"GREETING": "hello world",
	}

	encoded, err := codec.Encode(input)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	want := "APP_NAME=\"Zenith\"\nGREETING=\"hello world\"\n"
	if string(encoded) != want {
		t.Fatalf("Encode() = %q, want %q", encoded, want)
	}
}

func TestEndecDecode(t *testing.T) {
	codec := Endec{}
	input := []byte("# comment\nAPP_NAME=\"Zenith\"\nexport GREETING=\"hello world\"\n")
	want := map[string]any{
		"APP_NAME": "Zenith",
		"GREETING": "hello world",
	}

	var got map[string]any
	if err := codec.Decode(input, &got); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("Decode() returned %d values, want %d", len(got), len(want))
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Errorf("Decode()[%q] = %v, want %v", key, got[key], wantValue)
		}
	}
}

func TestEndecErrors(t *testing.T) {
	codec := Endec{}

	if _, err := codec.Encode(map[string]any{"bad-key": "value"}); err == nil {
		t.Error("Encode() expected an error for an invalid key")
	}
	if _, err := codec.Encode(map[string]any{"COUNT": 42}); err == nil {
		t.Error("Encode() expected an error for a non-string value")
	}
	if err := codec.Decode([]byte("BROKEN"), new(map[string]any)); err == nil {
		t.Error("Decode() expected an error for an invalid line")
	}
	if err := codec.Decode([]byte("KEY=\"unterminated"), new(map[string]any)); err == nil {
		t.Error("Decode() expected an error for an invalid quoted value")
	}
	if err := codec.Decode([]byte("KEY=value"), nil); err == nil {
		t.Error("Decode() expected an error for a nil destination")
	}
}

func TestEndecRoundTrip(t *testing.T) {
	codec := Endec{}
	input := map[string]any{"MESSAGE": "line one\nline two"}

	encoded, err := codec.Encode(input)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if !strings.Contains(string(encoded), `MESSAGE="line one\nline two"`) {
		t.Fatalf("Encode() did not quote newline value: %q", encoded)
	}

	var decoded map[string]any
	if err := codec.Decode(encoded, &decoded); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if decoded["MESSAGE"] != input["MESSAGE"] {
		t.Fatalf("round trip value = %q, want %q", decoded["MESSAGE"], input["MESSAGE"])
	}
}
