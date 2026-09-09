package filereader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOSReadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	want := []byte(`{"name":"Zenith"}`)
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	got, err := (OS{}).ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("ReadFile() = %q, want %q", got, want)
	}
}

func TestOSReadFileMissingFile(t *testing.T) {
	_, err := (OS{}).ReadFile(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("ReadFile() returned nil error")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ReadFile() error = %v, want os.ErrNotExist", err)
	}
}
