package toml

import (
	stderrors "errors"
	"reflect"
	"strings"
	"testing"

	zenitherrors "github.com/maneeshaindrachapa/zenith/internal/errors"
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
		"enabled = true",
		"hosts = [",
		"name = 'Zenith'",
		"port = 8080",
		"[database]",
		"host = '",
		"max_connections = 10",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("Encode() = %q, want containing %q", output, want)
		}
	}
}

func TestEndecDecode(t *testing.T) {
	input := []byte(`
# app config
name = "Zenith"
port = 8080
enabled = true
hosts = ["localhost", "example.com"]

[database]
host = "db.local"
max_connections = 10
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
	if _, err := (Endec{}).Encode(map[string]any{"unsupported": func() {}}); err == nil {
		t.Fatal("Encode() returned nil error for unsupported value")
	} else {
		assertConfigError(t, err, zenitherrors.ErrEncode)
	}

	for _, input := range [][]byte{
		[]byte("name = \"unterminated"),
		[]byte("[ ]"),
		[]byte("broken"),
	} {
		if err := (Endec{}).Decode(input, new(map[string]any)); err == nil {
			t.Fatalf("Decode(%q) returned nil error", input)
		} else {
			assertConfigError(t, err, zenitherrors.ErrDecode)
		}
	}

	if err := (Endec{}).Decode([]byte("name = \"Zenith\""), nil); err == nil {
		t.Fatal("Decode() returned nil error for nil destination")
	} else {
		assertConfigError(t, err, zenitherrors.ErrDecode)
	}
}

func assertConfigError(t *testing.T, err error, kind error) {
	t.Helper()
	if !stderrors.Is(err, kind) {
		t.Fatalf("error = %v, want %v", err, kind)
	}
	var configErr *zenitherrors.ConfigError
	if !stderrors.As(err, &configErr) {
		t.Fatalf("error = %v, want ConfigError", err)
	}
}

func TestEndecDecodeStandardTOMLSyntax(t *testing.T) {
	input := []byte(`
title = 'Zenith' # inline comment

[database.connection]
host = 'db.local'
ports = [8080, 8081]
`)

	var got map[string]any
	if err := (Endec{}).Decode(input, &got); err != nil {
		t.Fatalf("Decode() returned error: %v", err)
	}
	if got["title"] != "Zenith" {
		t.Fatalf("title = %#v, want %q", got["title"], "Zenith")
	}
	database, ok := got["database"].(map[string]any)
	if !ok {
		t.Fatalf("database = %#v, want nested table", got["database"])
	}
	connection, ok := database["connection"].(map[string]any)
	if !ok {
		t.Fatalf("connection = %#v, want two ports", database["connection"])
	}
	ports, ok := connection["ports"].([]any)
	if !ok || len(ports) != 2 {
		t.Fatalf("ports = %#v, want two ports", connection["ports"])
	}
}
