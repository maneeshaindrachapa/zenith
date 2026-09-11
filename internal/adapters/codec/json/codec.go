package json

import (
	"encoding/json"

	zenitherrors "github.com/maneeshaindrachapa/zenith/internal/errors"
)

// Endec encodes and decodes JSON objects.
type Endec struct{}

// Encode converts a map into indented JSON.
func (Endec) Encode(v map[string]any) ([]byte, error) {
	res, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		// Wrap the original error while preserving it for errors.Is/errors.As.
		return nil, zenitherrors.Wrap(zenitherrors.ErrEncode, "encode JSON", err)
	}
	return res, nil
}

// Decode parses JSON into the map pointed to by v.
func (Endec) Decode(b []byte, v *map[string]any) error {
	if v == nil {
		// A pointer is required so Decode can update the caller's map.
		return &zenitherrors.ConfigError{Kind: zenitherrors.ErrDecode, Operation: "decode JSON", Reason: "destination map pointer is nil"}
	}

	if err := json.Unmarshal(b, v); err != nil {
		// Include context while retaining the underlying JSON syntax error.
		return zenitherrors.Wrap(zenitherrors.ErrDecode, "decode JSON", err)
	}

	return nil
}
