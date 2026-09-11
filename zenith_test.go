package zenith

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type appConfig struct {
	Name     string         `json:"name"`
	Port     int            `json:"port" env:"PORT"`
	Enabled  bool           `json:"enabled" env:"ENABLED"`
	Timeout  time.Duration  `json:"timeout" env:"TIMEOUT"`
	Hosts    []string       `json:"hosts"`
	Database databaseConfig `json:"database"`
}

type databaseConfig struct {
	Host string `json:"host"`
	Max  int    `json:"max_connections"`
}

func TestLoadJSONFileIntoStruct(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := []byte(`{
		"name": "Zenith",
		"port": 8080,
		"enabled": true,
		"timeout": "5s",
		"hosts": ["localhost", "example.com"],
		"database": {
			"host": "db.local",
			"max_connections": 10
		}
	}`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	var got appConfig
	if err := Load(path, &got); err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	want := appConfig{
		Name:    "Zenith",
		Port:    8080,
		Enabled: true,
		Timeout: 5 * time.Second,
		Hosts:   []string{"localhost", "example.com"},
		Database: databaseConfig{
			Host: "db.local",
			Max:  10,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestLoadDotenvFileIntoStruct(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := []byte("NAME=\"Zenith\"\nPORT=\"8080\"\nENABLED=\"true\"\nTIMEOUT=\"250ms\"\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	var got appConfig
	if err := Load(path, &got); err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if got.Name != "Zenith" {
		t.Fatalf("Name = %q, want %q", got.Name, "Zenith")
	}
	if got.Port != 8080 {
		t.Fatalf("Port = %d, want %d", got.Port, 8080)
	}
	if !got.Enabled {
		t.Fatal("Enabled = false, want true")
	}
	if got.Timeout != 250*time.Millisecond {
		t.Fatalf("Timeout = %s, want %s", got.Timeout, 250*time.Millisecond)
	}
}

func TestLoadYAMLFileIntoStruct(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
name: "Zenith"
port: 8080
enabled: true
timeout: "5s"
hosts: ["localhost", "example.com"]
database:
  host: "db.local"
  max_connections: 10
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	var got appConfig
	if err := Load(path, &got); err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	assertAppConfig(t, got)
}

func TestLoadTOMLFileIntoStruct(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := []byte(`
name = "Zenith"
port = 8080
enabled = true
timeout = "5s"
hosts = ["localhost", "example.com"]

[database]
host = "db.local"
max_connections = 10
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	var got appConfig
	if err := Load(path, &got); err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	assertAppConfig(t, got)
}

func TestDecodeErrors(t *testing.T) {
	var got appConfig
	if err := Decode([]byte("{}"), "xml", &got); err == nil {
		t.Fatal("Decode() with unregistered format returned nil error")
	} else {
		if !errors.Is(err, ErrUnsupportedFormat) {
			t.Fatalf("Decode() error = %v, want ErrUnsupportedFormat", err)
		}
		var formatErr *ConfigError
		if !errors.As(err, &formatErr) {
			t.Fatalf("Decode() error = %v, want ConfigError", err)
		}
		if formatErr.Format != "xml" {
			t.Fatalf("ConfigError.Format = %q, want %q", formatErr.Format, "xml")
		}
	}
	if err := Decode([]byte("{}"), "json", got); err == nil {
		t.Fatal("Decode() with non-pointer target returned nil error")
	} else if !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("Decode() error = %v, want ErrInvalidTarget", err)
	}
}

func TestDecodeInvalidDataError(t *testing.T) {
	var got appConfig
	err := Decode([]byte(`{"invalid"}`), "json", &got)
	if err == nil {
		t.Fatal("Decode() returned nil error")
	}
	if !errors.Is(err, ErrDecode) {
		t.Fatalf("Decode() error = %v, want ErrDecode", err)
	}
	var configErr *ConfigError
	if !errors.As(err, &configErr) {
		t.Fatalf("Decode() error = %v, want ConfigError", err)
	}
	if configErr.Cause() == nil {
		t.Fatal("ConfigError.Cause() = nil, want JSON syntax cause")
	}
}

func assertAppConfig(t *testing.T, got appConfig) {
	t.Helper()

	want := appConfig{
		Name:    "Zenith",
		Port:    8080,
		Enabled: true,
		Timeout: 5 * time.Second,
		Hosts:   []string{"localhost", "example.com"},
		Database: databaseConfig{
			Host: "db.local",
			Max:  10,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("config = %#v, want %#v", got, want)
	}
}

func TestDecodeWithStrictMapping(t *testing.T) {
	var got struct {
		Name string `json:"name"`
	}

	err := Decode([]byte(`{"name":"Zenith","unknown":"value"}`), "json", &got, WithStrictMapping())
	if err == nil {
		t.Fatal("Decode() returned nil error")
	}
	if !errors.Is(err, ErrUnknownField) {
		t.Fatalf("Decode() error = %v, want ErrUnknownField", err)
	}
	var fieldErr *ConfigError
	if !errors.As(err, &fieldErr) {
		t.Fatalf("Decode() error = %v, want ConfigError", err)
	}
	if fieldErr.Path != "unknown" {
		t.Fatalf("ConfigError.Path = %q, want %q", fieldErr.Path, "unknown")
	}

	var pathed FieldPathed
	if !errors.As(err, &pathed) {
		t.Fatalf("Decode() error = %v, want FieldPathed", err)
	}
	if pathed.FieldPath() != "unknown" {
		t.Fatalf("FieldPath() = %q, want %q", pathed.FieldPath(), "unknown")
	}
}

func TestDecodeRequiredFieldError(t *testing.T) {
	var got struct {
		Name string `json:"name" required:"true"`
	}

	err := Decode([]byte(`{}`), "json", &got)
	if err == nil {
		t.Fatal("Decode() returned nil error")
	}
	if !errors.Is(err, ErrRequiredField) {
		t.Fatalf("Decode() error = %v, want ErrRequiredField", err)
	}
	var fieldErr *ConfigError
	if !errors.As(err, &fieldErr) {
		t.Fatalf("Decode() error = %v, want ConfigError", err)
	}
	if fieldErr.Path != "name" {
		t.Fatalf("ConfigError.Path = %q, want %q", fieldErr.Path, "name")
	}

	var reasoned Reasoned
	if !errors.As(err, &reasoned) {
		t.Fatalf("Decode() error = %v, want Reasoned", err)
	}
	if reasoned.ErrorReason() != "missing" {
		t.Fatalf("ErrorReason() = %q, want %q", reasoned.ErrorReason(), "missing")
	}
}

func TestDecodeWithPreserveExistingOnEmpty(t *testing.T) {
	got := struct {
		Name string `json:"name"`
	}{Name: "existing"}

	err := Decode([]byte(`{"name":""}`), "json", &got, WithPreserveExistingOnEmpty())
	if err != nil {
		t.Fatalf("Decode() returned error: %v", err)
	}
	if got.Name != "existing" {
		t.Fatalf("Name = %q, want %q", got.Name, "existing")
	}
}

func TestLoadErrors(t *testing.T) {
	if err := Load(filepath.Join(t.TempDir(), "config"), &appConfig{}); err == nil {
		t.Fatal("Load() without extension returned nil error")
	} else if !errors.Is(err, ErrMissingExtension) {
		t.Fatalf("Load() error = %v, want ErrMissingExtension", err)
	}

	if err := Load(filepath.Join(t.TempDir(), "missing.json"), &appConfig{}); err == nil {
		t.Fatal("Load() with missing file returned nil error")
	} else if !errors.Is(err, ErrReadFile) {
		t.Fatalf("Load() error = %v, want ErrReadFile", err)
	}
}
