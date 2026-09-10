package zenith

import (
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
	}
	if err := Decode([]byte("{}"), "json", got); err == nil {
		t.Fatal("Decode() with non-pointer target returned nil error")
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
	}
	if err := Load(filepath.Join(t.TempDir(), "missing.json"), &appConfig{}); err == nil {
		t.Fatal("Load() with missing file returned nil error")
	}
}
