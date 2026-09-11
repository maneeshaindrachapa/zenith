package config

import (
	stderrors "errors"
	"path/filepath"

	zenitherrors "github.com/maneeshaindrachapa/zenith/internal/errors"
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
		return &zenitherrors.ConfigError{Kind: zenitherrors.ErrMissingExtension, Operation: "load config", Path: path, Reason: "file extension is required"}
	}

	data, err := l.files.ReadFile(path)
	if err != nil {
		if isConfigError(err) {
			return err
		}
		return &zenitherrors.ConfigError{Kind: zenitherrors.ErrReadFile, Operation: "load config", Path: path, CauseErr: err}
	}

	if err := l.Decode(data, format, target); err != nil {
		if isConfigError(err) {
			return err
		}
		return &zenitherrors.ConfigError{Kind: zenitherrors.ErrDecode, Operation: "load config", Path: path, Format: format, CauseErr: err}
	}
	return nil
}

// Decode decodes data with the registered decoder for format and maps it into target.
func (l Loader) Decode(data []byte, format string, target any) error {
	decoder, err := l.decoders.DecoderFor(format)
	if err != nil {
		if isConfigError(err) {
			return err
		}
		return &zenitherrors.ConfigError{Kind: zenitherrors.ErrDecode, Operation: "find decoder", Format: format, CauseErr: err}
	}

	values := make(map[string]any)
	if err := decoder.Decode(data, &values); err != nil {
		if isConfigError(err) {
			return err
		}
		return &zenitherrors.ConfigError{Kind: zenitherrors.ErrDecode, Operation: "decode config", Format: format, CauseErr: err}
	}

	if err := l.mapper.MapToStruct(values, target); err != nil {
		if isConfigError(err) {
			return err
		}
		return &zenitherrors.ConfigError{Kind: zenitherrors.ErrDecode, Operation: "map config", Format: format, CauseErr: err}
	}
	return nil
}

func isConfigError(err error) bool {
	var configErr *zenitherrors.ConfigError
	return stderrors.As(err, &configErr)
}
