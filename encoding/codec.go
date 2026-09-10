package encoding

import (
	"fmt"
	"sort"
	"strings"

	"github.com/maneeshaindrachapa/zenith/internal/adapters/codec/dotenv"
	zenith_json "github.com/maneeshaindrachapa/zenith/internal/adapters/codec/json"
	"github.com/maneeshaindrachapa/zenith/internal/adapters/codec/toml"
	"github.com/maneeshaindrachapa/zenith/internal/adapters/codec/yaml"
	"github.com/maneeshaindrachapa/zenith/internal/ports"
)

type Encoder = ports.Encoder

type Decoder = ports.Decoder

type Codec = ports.Codec

type Registry struct{}

// codecs stores the built-in and user-registered codecs by normalized format.
var codecs = map[string]Codec{
	"dotenv": dotenv.Endec{},
	"env":    dotenv.Endec{},
	"json":   zenith_json.Endec{},
	"toml":   toml.Endec{},
	"yaml":   yaml.Endec{},
	"yml":    yaml.Endec{},
}

// DefaultRegistry returns the registry used by the package-level helpers.
func DefaultRegistry() Registry {
	return Registry{}
}

// Register makes codec available for future lookups by format.
func Register(format string, codec Codec) error {
	return DefaultRegistry().Register(format, codec)
}

// Register makes codec available for future lookups by format.
func (Registry) Register(format string, codec Codec) error {
	name := normalizeFormat(format)
	if name == "" {
		return fmt.Errorf("register codec: format is required")
	}
	if codec == nil {
		return fmt.Errorf("register codec %q: codec is nil", name)
	}

	codecs[name] = codec
	return nil
}

// CodecFor returns the codec registered for format.
func CodecFor(format string) (Codec, error) {
	return DefaultRegistry().CodecFor(format)
}

// CodecFor returns the codec registered for format.
func (Registry) CodecFor(format string) (Codec, error) {
	name := normalizeFormat(format)
	codec, ok := codecs[name]
	if !ok {
		return nil, fmt.Errorf("codec for %q is not registered", format)
	}
	return codec, nil
}

// EncoderFor returns the encoder registered for format.
func EncoderFor(format string) (Encoder, error) {
	return DefaultRegistry().EncoderFor(format)
}

// EncoderFor returns the encoder registered for format.
func (registry Registry) EncoderFor(format string) (Encoder, error) {
	return registry.CodecFor(format)
}

// DecoderFor returns the decoder registered for format.
func DecoderFor(format string) (Decoder, error) {
	return DefaultRegistry().DecoderFor(format)
}

// DecoderFor returns the decoder registered for format.
func (registry Registry) DecoderFor(format string) (Decoder, error) {
	return registry.CodecFor(format)
}

// RegisteredFormats returns the registered format names in sorted order.
func RegisteredFormats() []string {
	return DefaultRegistry().RegisteredFormats()
}

// RegisteredFormats returns the registered format names in sorted order.
func (Registry) RegisteredFormats() []string {
	formats := make([]string, 0, len(codecs))
	for format := range codecs {
		formats = append(formats, format)
	}
	sort.Strings(formats)
	return formats
}

// normalizeFormat makes user-provided format names safe for registry lookup.
func normalizeFormat(format string) string {
	format = strings.TrimSpace(strings.ToLower(format))
	format = strings.TrimPrefix(format, ".")
	return format
}
