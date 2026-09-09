package config

import (
	"fmt"
	"path/filepath"

	"github.com/maneeshaindrachapa/zenith/internal/ports"
)

// Loader coordinates reading, decoding, and mapping configuration.
type Loader struct {
	files    ports.FileReader
	decoders ports.DecoderRegistry
	mapper   ports.Mapper
}

// NewLoader creates a configuration loader from its required ports.
func NewLoader(files ports.FileReader, decoders ports.DecoderRegistry, mapper ports.Mapper) Loader {
	return Loader{
		files:    files,
		decoders: decoders,
		mapper:   mapper,
	}
}

// Load reads path, decodes it using the file extension, and maps values into target.
func (l Loader) Load(path string, target any) error {
	format := filepath.Ext(path)
	if format == "" {
		return fmt.Errorf("load config %q: file extension is required", path)
	}

	data, err := l.files.ReadFile(path)
	if err != nil {
		return fmt.Errorf("load config %q: %w", path, err)
	}

	if err := l.Decode(data, format, target); err != nil {
		return fmt.Errorf("load config %q: %w", path, err)
	}
	return nil
}

// Decode decodes data with the registered decoder for format and maps it into target.
func (l Loader) Decode(data []byte, format string, target any) error {
	decoder, err := l.decoders.DecoderFor(format)
	if err != nil {
		return err
	}

	values := make(map[string]any)
	if err := decoder.Decode(data, &values); err != nil {
		return err
	}

	return l.mapper.MapToStruct(values, target)
}
