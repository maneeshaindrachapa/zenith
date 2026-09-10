package config

import (
	"errors"
	"reflect"
	"testing"

	"github.com/maneeshaindrachapa/zenith/internal/ports"
)

var (
	errRead    = errors.New("read failed")
	errLookup  = errors.New("decoder lookup failed")
	errDecode  = errors.New("decode failed")
	errMapping = errors.New("mapping failed")
)

type fakeFileReader struct {
	data     []byte
	err      error
	readPath string
}

func (f *fakeFileReader) ReadFile(path string) ([]byte, error) {
	f.readPath = path
	if f.err != nil {
		return nil, f.err
	}
	return f.data, nil
}

type fakeDecoderRegistry struct {
	decoder fakeDecoder
	err     error
	format  string
}

func (f *fakeDecoderRegistry) DecoderFor(format string) (ports.Decoder, error) {
	f.format = format
	if f.err != nil {
		return fakeDecoder{}, f.err
	}
	return f.decoder, nil
}

type fakeDecoder struct {
	values map[string]any
	err    error
	data   []byte
}

func (f fakeDecoder) Decode(data []byte, values *map[string]any) error {
	f.data = data
	if f.err != nil {
		return f.err
	}
	*values = f.values
	return nil
}

type fakeMapper struct {
	err    error
	values map[string]any
	target any
}

func (f *fakeMapper) MapToStruct(values map[string]any, target any) error {
	f.values = values
	f.target = target
	return f.err
}

func TestLoaderLoad(t *testing.T) {
	files := &fakeFileReader{data: []byte(`{"name":"Zenith"}`)}
	decoders := &fakeDecoderRegistry{
		decoder: fakeDecoder{values: map[string]any{"name": "Zenith"}},
	}
	mapper := &fakeMapper{}
	loader := NewLoader(files, decoders, mapper)

	var target struct {
		Name string
	}
	if err := loader.Load("config.json", &target); err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if files.readPath != "config.json" {
		t.Fatalf("ReadFile() path = %q, want %q", files.readPath, "config.json")
	}
	if decoders.format != ".json" {
		t.Fatalf("DecoderFor() format = %q, want %q", decoders.format, ".json")
	}
	if !reflect.DeepEqual(mapper.values, map[string]any{"name": "Zenith"}) {
		t.Fatalf("MapToStruct() values = %#v", mapper.values)
	}
	if mapper.target != &target {
		t.Fatalf("MapToStruct() target = %#v, want target pointer", mapper.target)
	}
}

func TestLoaderDecode(t *testing.T) {
	decoders := &fakeDecoderRegistry{
		decoder: fakeDecoder{values: map[string]any{"enabled": true}},
	}
	mapper := &fakeMapper{}
	loader := NewLoader(&fakeFileReader{}, decoders, mapper)

	var target struct {
		Enabled bool
	}
	if err := loader.Decode([]byte(`{"enabled":true}`), "json", &target); err != nil {
		t.Fatalf("Decode() returned error: %v", err)
	}

	if decoders.format != "json" {
		t.Fatalf("DecoderFor() format = %q, want %q", decoders.format, "json")
	}
	if !reflect.DeepEqual(mapper.values, map[string]any{"enabled": true}) {
		t.Fatalf("MapToStruct() values = %#v", mapper.values)
	}
}

func TestLoaderLoadErrors(t *testing.T) {
	tests := []struct {
		name    string
		loader  Loader
		path    string
		wantErr error
	}{
		{
			name:    "missing extension",
			loader:  NewLoader(&fakeFileReader{}, &fakeDecoderRegistry{}, &fakeMapper{}),
			path:    "config",
			wantErr: nil,
		},
		{
			name:    "read failure",
			loader:  NewLoader(&fakeFileReader{err: errRead}, &fakeDecoderRegistry{}, &fakeMapper{}),
			path:    "config.json",
			wantErr: errRead,
		},
		{
			name: "decode failure",
			loader: NewLoader(
				&fakeFileReader{data: []byte("bad")},
				&fakeDecoderRegistry{decoder: fakeDecoder{err: errDecode}},
				&fakeMapper{},
			),
			path:    "config.json",
			wantErr: errDecode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loader.Load(tt.path, &struct{}{})
			if err == nil {
				t.Fatal("Load() returned nil error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Load() error = %v, want wrapping %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoaderDecodeErrors(t *testing.T) {
	tests := []struct {
		name    string
		loader  Loader
		wantErr error
	}{
		{
			name:    "decoder lookup failure",
			loader:  NewLoader(&fakeFileReader{}, &fakeDecoderRegistry{err: errLookup}, &fakeMapper{}),
			wantErr: errLookup,
		},
		{
			name: "decode failure",
			loader: NewLoader(
				&fakeFileReader{},
				&fakeDecoderRegistry{decoder: fakeDecoder{err: errDecode}},
				&fakeMapper{},
			),
			wantErr: errDecode,
		},
		{
			name: "mapper failure",
			loader: NewLoader(
				&fakeFileReader{},
				&fakeDecoderRegistry{decoder: fakeDecoder{values: map[string]any{"name": "Zenith"}}},
				&fakeMapper{err: errMapping},
			),
			wantErr: errMapping,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loader.Decode([]byte("{}"), "json", &struct{}{})
			if err == nil {
				t.Fatal("Decode() returned nil error")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Decode() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
