package toml

import (
	"github.com/pelletier/go-toml/v2"

	zenitherrors "github.com/maneeshaindrachapa/zenith/internal/errors"
)

// Endec encodes and decodes TOML mapping data.
type Endec struct{}

// Encode converts a map into TOML data.
func (Endec) Encode(v map[string]any) ([]byte, error) {
	data, err := toml.Marshal(v)
	if err != nil {
		return nil, zenitherrors.Wrap(zenitherrors.ErrEncode, "encode TOML", err)
	}
	return data, nil
}

// Decode parses TOML mapping data into the map pointed to by v.
func (Endec) Decode(data []byte, v *map[string]any) error {
	if v == nil {
		return &zenitherrors.ConfigError{Kind: zenitherrors.ErrDecode, Operation: "decode TOML", Reason: "destination map pointer is nil"}
	}
	if *v == nil {
		*v = make(map[string]any)
	}

	if err := toml.Unmarshal(data, v); err != nil {
		return zenitherrors.Wrap(zenitherrors.ErrDecode, "decode TOML", err)
	}
	return nil
}
