package zenith

import (
	"github.com/maneeshaindrachapa/zenith/encoding"
	"github.com/maneeshaindrachapa/zenith/internal/adapters/filereader"
	"github.com/maneeshaindrachapa/zenith/internal/adapters/mapper"
	"github.com/maneeshaindrachapa/zenith/internal/app/config"
	zenitherrors "github.com/maneeshaindrachapa/zenith/internal/errors"
)

var (
	// ErrUnknownField reports decoded data that does not map to a struct field.
	ErrUnknownField = zenitherrors.ErrUnknownField

	// ErrRequiredField reports a required struct field with no usable value.
	ErrRequiredField = zenitherrors.ErrRequiredField

	// ErrUnsupportedFormat reports that no codec is registered for a format.
	ErrUnsupportedFormat = zenitherrors.ErrUnsupportedFormat

	// ErrMissingExtension reports a config file path with no extension.
	ErrMissingExtension = zenitherrors.ErrMissingExtension

	// ErrReadFile reports a file read failure.
	ErrReadFile = zenitherrors.ErrReadFile

	// ErrEncode reports an encoding failure.
	ErrEncode = zenitherrors.ErrEncode

	// ErrDecode reports a decoding failure.
	ErrDecode = zenitherrors.ErrDecode

	// ErrInvalidTarget reports an invalid mapping target.
	ErrInvalidTarget = zenitherrors.ErrInvalidTarget

	// ErrInvalidValue reports an invalid or unsupported config value.
	ErrInvalidValue = zenitherrors.ErrInvalidValue

	// ErrValidation reports a user validation failure.
	ErrValidation = zenitherrors.ErrValidation

	// ErrRegistry reports invalid codec registry usage.
	ErrRegistry = zenitherrors.ErrRegistry
)

// ConfigError is the single structured error type used by Zenith.
type ConfigError = zenitherrors.ConfigError

// Kinded exposes the sentinel error kind wrapped by ConfigError.
type Kinded = zenitherrors.Kinded

// FieldPathed exposes the config field path related to an error.
type FieldPathed = zenitherrors.FieldPathed

// Formatted exposes the config format related to an error.
type Formatted = zenitherrors.Formatted

// Reasoned exposes a short reason related to an error.
type Reasoned = zenitherrors.Reasoned

// Caused exposes the lower-level error wrapped by ConfigError.
type Caused = zenitherrors.Caused

// Option customizes how configuration values are loaded and mapped.
type Option func(*options)

type options struct {
	mapper mapper.Options
}

// Load reads path, decodes it using the file extension, and maps values into target.
func Load(path string, target any, opts ...Option) error {
	return defaultLoader(opts...).Load(path, target)
}

// Decode decodes data with the registered decoder for format and maps it into target.
func Decode(data []byte, format string, target any, opts ...Option) error {
	return defaultLoader(opts...).Decode(data, format, target)
}

// WithStrictMapping returns an error when decoded data contains unknown fields.
func WithStrictMapping() Option {
	return func(opts *options) {
		opts.mapper.Strict = true
	}
}

// WithPreserveExistingOnEmpty keeps existing struct values when config contains "".
func WithPreserveExistingOnEmpty() Option {
	return func(opts *options) {
		opts.mapper.PreserveExistingOnEmpty = true
	}
}

func defaultLoader(opts ...Option) config.Loader {
	settings := options{}
	for _, opt := range opts {
		if opt != nil {
			opt(&settings)
		}
	}

	return config.NewLoader(filereader.OS{}, encoding.DefaultRegistry(), mapper.Reflect{Options: settings.mapper})
}
