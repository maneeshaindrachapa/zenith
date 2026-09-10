package mapper

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

type reflectConfig struct {
	Name       string        `json:"name"`
	Port       int           `env:"PORT"`
	Enabled    bool          `mapstructure:"enabled"`
	Ratio      float64       `zenith:"ratio"`
	Retries    uint          `json:"retries"`
	Timeout    time.Duration `json:"timeout"`
	Hosts      []string      `json:"hosts"`
	Endpoint   endpoint      `json:"endpoint"`
	Optional   *int          `json:"optional"`
	Nested     nestedConfig  `json:"nested"`
	Ignored    string        `json:"-"`
	unexported string
}

type nestedConfig struct {
	MaxConnections int `json:"max_connections"`
}

type endpoint string

func (e *endpoint) UnmarshalText(text []byte) error {
	value := string(text)
	if !strings.HasPrefix(value, "https://") {
		return fmt.Errorf("endpoint must start with https://")
	}
	*e = endpoint(value)
	return nil
}

func TestReflectMapToStruct(t *testing.T) {
	optional := 3
	input := map[string]any{
		"name":            "Zenith",
		"PORT":            "8080",
		"enabled":         "true",
		"ratio":           "1.5",
		"retries":         "2",
		"timeout":         "250ms",
		"hosts":           []any{"localhost", "example.com"},
		"endpoint":        "https://example.com/config",
		"optional":        optional,
		"nested":          map[string]any{"max_connections": float64(10)},
		"Ignored":         "should not map",
		"unexported":      "should not map",
		"missing_ignored": "should not matter",
	}

	var got reflectConfig
	if err := (Reflect{}).MapToStruct(input, &got); err != nil {
		t.Fatalf("MapToStruct() returned error: %v", err)
	}

	want := reflectConfig{
		Name:     "Zenith",
		Port:     8080,
		Enabled:  true,
		Ratio:    1.5,
		Retries:  2,
		Timeout:  250 * time.Millisecond,
		Hosts:    []string{"localhost", "example.com"},
		Endpoint: endpoint("https://example.com/config"),
		Optional: &optional,
		Nested: nestedConfig{
			MaxConnections: 10,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MapToStruct() = %#v, want %#v", got, want)
	}
}

func TestReflectMapToStructNormalizesKeys(t *testing.T) {
	input := map[string]any{
		"MAX-CONNECTIONS": "7",
	}
	var got nestedConfig

	if err := (Reflect{}).MapToStruct(input, &got); err != nil {
		t.Fatalf("MapToStruct() returned error: %v", err)
	}
	if got.MaxConnections != 7 {
		t.Fatalf("MaxConnections = %d, want %d", got.MaxConnections, 7)
	}
}

func TestReflectMapToStructTargetErrors(t *testing.T) {
	tests := []struct {
		name   string
		target any
	}{
		{name: "nil target", target: nil},
		{name: "non pointer", target: reflectConfig{}},
		{name: "nil pointer", target: (*reflectConfig)(nil)},
		{name: "non struct pointer", target: new(string)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := (Reflect{}).MapToStruct(map[string]any{}, tt.target); err == nil {
				t.Fatal("MapToStruct() returned nil error")
			}
		})
	}
}

func TestReflectMapToStructConversionErrors(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]any
		wantError string
	}{
		{
			name:      "invalid bool",
			input:     map[string]any{"enabled": "maybe"},
			wantError: "map field enabled",
		},
		{
			name:      "fractional int",
			input:     map[string]any{"PORT": 1.5},
			wantError: "expected integer",
		},
		{
			name:      "negative uint",
			input:     map[string]any{"retries": -1.0},
			wantError: "expected unsigned integer",
		},
		{
			name:      "invalid float",
			input:     map[string]any{"ratio": "large-ish"},
			wantError: "expected float",
		},
		{
			name:      "invalid duration",
			input:     map[string]any{"timeout": "eventually"},
			wantError: "expected duration",
		},
		{
			name:      "invalid nested object",
			input:     map[string]any{"nested": "not an object"},
			wantError: "expected object",
		},
		{
			name:      "invalid slice",
			input:     map[string]any{"hosts": "localhost"},
			wantError: "expected slice",
		},
		{
			name:      "invalid text unmarshaler",
			input:     map[string]any{"endpoint": "http://example.com/config"},
			wantError: "endpoint must start with https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got reflectConfig
			err := (Reflect{}).MapToStruct(tt.input, &got)
			if err == nil {
				t.Fatal("MapToStruct() returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("MapToStruct() error = %q, want containing %q", err, tt.wantError)
			}
		})
	}
}

type strictConfig struct {
	Name     string             `json:"name" required:"true"`
	Port     int                `json:"port" default:"8080"`
	Region   string             `json:"region" default:"us-east-1"`
	Database strictDatabase     `json:"database"`
	Extra    map[string]string  `json:"extra"`
	Mode     requiredByValidate `json:"mode" validate:"required"`
}

type strictDatabase struct {
	Host string `json:"host" required:"true"`
	Port int    `json:"port" default:"5432"`
}

type requiredByValidate string

type validatingConfig struct {
	Name string `json:"name"`
}

func (c *validatingConfig) Validate() error {
	if c.Name != "Zenith" {
		return fmt.Errorf("name must be Zenith")
	}
	return nil
}

func TestReflectStrictModeRejectsUnknownFields(t *testing.T) {
	var got struct {
		Database strictDatabase `json:"database"`
	}
	err := (Reflect{Options: Options{Strict: true}}).MapToStruct(map[string]any{
		"database": map[string]any{
			"host":    "db.local",
			"unknown": "value",
		},
	}, &got)

	if err == nil {
		t.Fatal("MapToStruct() returned nil error")
	}
	if !strings.Contains(err.Error(), "database.unknown") {
		t.Fatalf("MapToStruct() error = %q, want database.unknown", err)
	}
}

func TestReflectAppliesDefaults(t *testing.T) {
	var got strictConfig
	if err := (Reflect{}).MapToStruct(map[string]any{
		"name":     "Zenith",
		"database": map[string]any{"host": "db.local"},
		"mode":     "prod",
	}, &got); err != nil {
		t.Fatalf("MapToStruct() returned error: %v", err)
	}

	if got.Port != 8080 {
		t.Fatalf("Port = %d, want %d", got.Port, 8080)
	}
	if got.Region != "us-east-1" {
		t.Fatalf("Region = %q, want %q", got.Region, "us-east-1")
	}
	if got.Database.Port != 5432 {
		t.Fatalf("Database.Port = %d, want %d", got.Database.Port, 5432)
	}
}

func TestReflectRequiredFieldErrorsUsePaths(t *testing.T) {
	var got strictConfig
	err := (Reflect{}).MapToStruct(map[string]any{
		"name":     "Zenith",
		"database": map[string]any{},
		"mode":     "prod",
	}, &got)

	if err == nil {
		t.Fatal("MapToStruct() returned nil error")
	}
	if !strings.Contains(err.Error(), "database.host") {
		t.Fatalf("MapToStruct() error = %q, want database.host", err)
	}
}

func TestReflectValidateRequiredTag(t *testing.T) {
	var got strictConfig
	err := (Reflect{}).MapToStruct(map[string]any{
		"name":     "Zenith",
		"database": map[string]any{"host": "db.local"},
	}, &got)

	if err == nil {
		t.Fatal("MapToStruct() returned nil error")
	}
	if !strings.Contains(err.Error(), "mode") {
		t.Fatalf("MapToStruct() error = %q, want mode", err)
	}
}

func TestReflectEmptyStringOverwritesByDefault(t *testing.T) {
	got := struct {
		Name string `json:"name"`
	}{Name: "existing"}

	if err := (Reflect{}).MapToStruct(map[string]any{"name": ""}, &got); err != nil {
		t.Fatalf("MapToStruct() returned error: %v", err)
	}
	if got.Name != "" {
		t.Fatalf("Name = %q, want empty string", got.Name)
	}
}

func TestReflectPreserveExistingOnEmpty(t *testing.T) {
	got := struct {
		Name string `json:"name"`
	}{Name: "existing"}

	err := (Reflect{Options: Options{PreserveExistingOnEmpty: true}}).MapToStruct(map[string]any{"name": ""}, &got)
	if err != nil {
		t.Fatalf("MapToStruct() returned error: %v", err)
	}
	if got.Name != "existing" {
		t.Fatalf("Name = %q, want %q", got.Name, "existing")
	}
}

func TestReflectCallsValidateMethod(t *testing.T) {
	var got validatingConfig
	err := (Reflect{}).MapToStruct(map[string]any{"name": "Wrong"}, &got)
	if err == nil {
		t.Fatal("MapToStruct() returned nil error")
	}
	if !strings.Contains(err.Error(), "validate config") {
		t.Fatalf("MapToStruct() error = %q, want validation context", err)
	}
}
