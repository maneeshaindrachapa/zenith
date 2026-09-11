package yaml

import (
	"fmt"

	"gopkg.in/yaml.v3"

	zenitherrors "github.com/maneeshaindrachapa/zenith/internal/errors"
)

// Endec encodes and decodes YAML mapping data.
type Endec struct{}

// Encode converts a map into YAML data.
func (Endec) Encode(v map[string]any) ([]byte, error) {
	data, err := marshal(v)
	if err != nil {
		return nil, zenitherrors.Wrap(zenitherrors.ErrEncode, "encode YAML", err)
	}
	return data, nil
}

func marshal(v map[string]any) (data []byte, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("unsupported YAML value: %v", recovered)
		}
	}()
	return yaml.Marshal(v)
}

// Decode parses YAML mapping data into the map pointed to by v.
func (Endec) Decode(data []byte, v *map[string]any) error {
	if v == nil {
		return &zenitherrors.ConfigError{Kind: zenitherrors.ErrDecode, Operation: "decode YAML", Reason: "destination map pointer is nil"}
	}
	if *v == nil {
		*v = make(map[string]any)
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return zenitherrors.Wrap(zenitherrors.ErrDecode, "decode YAML", err)
	}
	return nil
}
