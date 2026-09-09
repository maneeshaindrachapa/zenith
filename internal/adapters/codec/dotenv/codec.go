package dotenv

import (
	"bufio"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Endec encodes and decodes dotenv key-value data.
type Endec struct{}

// Encode converts a map into dotenv-formatted data.
func (Endec) Encode(v map[string]any) ([]byte, error) {
	// Sort keys so the output is deterministic and easy to compare.
	keys := make([]string, 0, len(v))
	for key := range v {
		if !validKey(key) {
			return nil, fmt.Errorf("encode dotenv: invalid key %q", key)
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)
	var output strings.Builder
	for _, key := range keys {
		value, ok := v[key].(string)
		if !ok {
			return nil, fmt.Errorf("encode dotenv: value for %q must be a string", key)
		}

		// Quote values so spaces, # characters, and newlines are preserved.
		fmt.Fprintf(&output, "%s=%s\n", key, strconv.Quote(value))
	}

	return []byte(output.String()), nil
}

// Decode parses dotenv data into the map pointed to by v.
func (Endec) Decode(data []byte, v *map[string]any) error {
	if v == nil {
		// A pointer is required so Decode can initialize the caller's map.
		return fmt.Errorf("decode dotenv: destination map pointer is nil")
	}
	if *v == nil {
		*v = make(map[string]any)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")
		key, value, found := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !found || !validKey(key) {
			return fmt.Errorf("decode dotenv: invalid line %d", lineNumber)
		}

		value = strings.TrimSpace(value)
		if len(value) > 1 && value[0] == '"' {
			decoded, err := strconv.Unquote(value)
			if err != nil {
				return fmt.Errorf("decode dotenv: invalid quoted value on line %d: %w", lineNumber, err)
			}
			value = decoded
		}
		(*v)[key] = value
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("decode dotenv: %w", err)
	}
	return nil
}

func validKey(key string) bool {
	if key == "" || !((key[0] >= 'A' && key[0] <= 'Z') || (key[0] >= 'a' && key[0] <= 'z') || key[0] == '_') {
		return false
	}
	for _, char := range key[1:] {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	return true
}
