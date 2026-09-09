package zenith

import (
	"github.com/maneeshaindrachapa/zenith/encoding"
	"github.com/maneeshaindrachapa/zenith/internal/adapters/filereader"
	"github.com/maneeshaindrachapa/zenith/internal/adapters/mapper"
	"github.com/maneeshaindrachapa/zenith/internal/app/config"
)

// Load reads path, decodes it using the file extension, and maps values into target.
func Load(path string, target any) error {
	return defaultLoader().Load(path, target)
}

// Decode decodes data with the registered decoder for format and maps it into target.
func Decode(data []byte, format string, target any) error {
	return defaultLoader().Decode(data, format, target)
}

func defaultLoader() config.Loader {
	return config.NewLoader(filereader.OS{}, encoding.DefaultRegistry(), mapper.Reflect{})
}
