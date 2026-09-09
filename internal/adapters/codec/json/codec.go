package json

import (
	"encoding/json"
	"fmt"
)

// Endec encodes and decodes JSON objects.
type Endec struct{}

// Encode converts a map into indented JSON.
func (Endec) Encode(v map[string]any) ([]byte, error) {
	res, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		// Wrap the original error while preserving it for errors.Is/errors.As.
		return nil, fmt.Errorf("encode JSON: %w", err)
	}
	return res, nil
}

// Decode parses JSON into the map pointed to by v.
func (Endec) Decode(b []byte, v *map[string]any) error {
	if v == nil {
		// A pointer is required so Decode can update the caller's map.
		return fmt.Errorf("decode JSON: destination map pointer is nil")
	}

	if err := json.Unmarshal(b, v); err != nil {
		// Include context while retaining the underlying JSON syntax error.
		return fmt.Errorf("decode JSON: %w", err)
	}

	return nil
}
