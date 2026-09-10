package zenith

import (
	"github.com/maneeshaindrachapa/zenith/encoding"
	"github.com/maneeshaindrachapa/zenith/internal/adapters/filereader"
	"github.com/maneeshaindrachapa/zenith/internal/adapters/mapper"
	"github.com/maneeshaindrachapa/zenith/internal/app/config"
)

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
