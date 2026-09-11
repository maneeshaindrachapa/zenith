package dotenv

import (
	stderrors "errors"
	"strings"
	"testing"

	zenitherrors "github.com/maneeshaindrachapa/zenith/internal/errors"
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

	for _, input := range []map[string]any{
		{"bad-key": "value"},
		{"COUNT": 42},
	} {
		if _, err := codec.Encode(input); err == nil {
			t.Errorf("Encode(%#v) expected an error", input)
		} else {
			assertConfigError(t, err, zenitherrors.ErrEncode)
		}
	}

	for _, input := range [][]byte{
		[]byte("BROKEN"),
		[]byte("KEY=\"unterminated"),
	} {
		if err := codec.Decode(input, new(map[string]any)); err == nil {
			t.Errorf("Decode(%q) expected an error", input)
		} else {
			assertConfigError(t, err, zenitherrors.ErrDecode)
		}
	}

	if err := codec.Decode([]byte("KEY=value"), nil); err == nil {
		t.Error("Decode() expected an error for a nil destination")
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
