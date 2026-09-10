package yaml

import (
	"reflect"
	"strings"
	"testing"
)

func TestEndecEncode(t *testing.T) {
	encoded, err := (Endec{}).Encode(map[string]any{
		"enabled": true,
		"hosts":   []string{"localhost", "example.com"},
		"name":    "Zenith",
		"port":    8080,
		"database": map[string]any{
			"host":            "db.local",
			"max_connections": 10,
		},
	})
	if err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}

	output := string(encoded)
	for _, want := range []string{
		"enabled: true",
		`hosts: ["localhost", "example.com"]`,
		`name: "Zenith"`,
		"port: 8080",
		"database:",
		`  host: "db.local"`,
		"  max_connections: 10",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("Encode() = %q, want containing %q", output, want)
		}
	}
}

func TestEndecDecode(t *testing.T) {
	input := []byte(`
# app config
name: "Zenith"
port: 8080
enabled: true
hosts: ["localhost", "example.com"]
database:
  host: "db.local"
  max_connections: 10
`)

	var got map[string]any
	if err := (Endec{}).Decode(input, &got); err != nil {
		t.Fatalf("Decode() returned error: %v", err)
	}

	want := map[string]any{
		"name":    "Zenith",
		"port":    int64(8080),
		"enabled": true,
		"hosts":   []any{"localhost", "example.com"},
		"database": map[string]any{
			"host":            "db.local",
			"max_connections": int64(10),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Decode() = %#v, want %#v", got, want)
	}
}

func TestEndecRoundTrip(t *testing.T) {
	codec := Endec{}
	input := map[string]any{
		"name":  "Zenith",
		"ratio": 1.5,
		"flags": []any{true, "stable"},
	}

	encoded, err := codec.Encode(input)
	if err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}

	var decoded map[string]any
	if err := codec.Decode(encoded, &decoded); err != nil {
		t.Fatalf("Decode() returned error: %v", err)
	}

	if decoded["name"] != input["name"] || decoded["ratio"] != input["ratio"] {
		t.Fatalf("round trip decoded = %#v", decoded)
	}
	if !reflect.DeepEqual(decoded["flags"], []any{true, "stable"}) {
		t.Fatalf("round trip flags = %#v", decoded["flags"])
	}
}

func TestEndecErrors(t *testing.T) {
	if _, err := (Endec{}).Encode(map[string]any{"unsupported": []int{1}}); err == nil {
		t.Fatal("Encode() returned nil error for unsupported value")
	}
	if err := (Endec{}).Decode([]byte(" name: value"), new(map[string]any)); err == nil {
		t.Fatal("Decode() returned nil error for invalid indentation")
	}
	if err := (Endec{}).Decode([]byte("name \"Zenith\""), new(map[string]any)); err == nil {
		t.Fatal("Decode() returned nil error for invalid line")
	}
	if err := (Endec{}).Decode([]byte("name: \"unterminated"), new(map[string]any)); err == nil {
		t.Fatal("Decode() returned nil error for invalid string")
	}
	if err := (Endec{}).Decode([]byte("name: Zenith"), nil); err == nil {
		t.Fatal("Decode() returned nil error for nil destination")
	}
}
