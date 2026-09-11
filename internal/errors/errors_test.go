package errors

import (
	stderrors "errors"
	"io/fs"
	"testing"
)

func TestConfigErrorUnsupportedFormat(t *testing.T) {
	err := NewUnsupportedFormat("xml")
	if got, want := err.Error(), `codec for "xml" is not registered`; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if !stderrors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("errors.Is() = false, want true")
	}
	if got, want := err.FormatName(), "xml"; got != want {
		t.Fatalf("FormatName() = %q, want %q", got, want)
	}
}

func TestConfigErrorUnknownField(t *testing.T) {
	err := NewUnknownField("database.host")
	if got, want := err.Error(), "unknown field database.host"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if !stderrors.Is(err, ErrUnknownField) {
		t.Fatalf("errors.Is() = false, want true")
	}
	if got, want := err.FieldPath(), "database.host"; got != want {
		t.Fatalf("FieldPath() = %q, want %q", got, want)
	}
}

func TestConfigErrorRequiredField(t *testing.T) {
	err := NewRequiredField("database.host", "missing")
	if got, want := err.Error(), "required field database.host: missing"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if !stderrors.Is(err, ErrRequiredField) {
		t.Fatalf("errors.Is() = false, want true")
	}
	if got, want := err.ErrorReason(), "missing"; got != want {
		t.Fatalf("ErrorReason() = %q, want %q", got, want)
	}
}

func TestConfigErrorRequiredFieldWithoutReason(t *testing.T) {
	err := NewRequiredField("database.host", "")
	if got, want := err.Error(), "required field database.host"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestConfigErrorUnknownKind(t *testing.T) {
	err := &ConfigError{}
	if got, want := err.Error(), "config error"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if err.ErrorKind() != nil {
		t.Fatalf("ErrorKind() = %v, want nil", err.ErrorKind())
	}
}

func TestConfigErrorWrapsKindAndCause(t *testing.T) {
	err := Wrap(ErrReadFile, "load config", fs.ErrNotExist)

	if !stderrors.Is(err, ErrReadFile) {
		t.Fatalf("errors.Is() = false for ErrReadFile, want true")
	}
	if !stderrors.Is(err, fs.ErrNotExist) {
		t.Fatalf("errors.Is() = false for wrapped cause, want true")
	}
	if got, want := err.Cause(), fs.ErrNotExist; got != want {
		t.Fatalf("Cause() = %v, want %v", got, want)
	}
	if got, want := err.ErrorKind(), ErrReadFile; got != want {
		t.Fatalf("ErrorKind() = %v, want %v", got, want)
	}
}
