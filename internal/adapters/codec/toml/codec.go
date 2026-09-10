package toml

import (
	"bufio"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Endec struct{}

// Encode converts a map into simple TOML data.
func (Endec) Encode(v map[string]any) ([]byte, error) {
	var output strings.Builder

	keys := sortedKeys(v)
	for _, key := range keys {
		value := v[key]
		if _, ok := value.(map[string]any); ok {
			continue
		}

		encoded, err := encodeValue(value)
		if err != nil {
			return nil, fmt.Errorf("encode TOML key %q: %w", key, err)
		}
		fmt.Fprintf(&output, "%s = %s\n", key, encoded)
	}

	for _, key := range keys {
		nested, ok := v[key].(map[string]any)
		if !ok {
			continue
		}

		if output.Len() > 0 {
			output.WriteByte('\n')
		}
		fmt.Fprintf(&output, "[%s]\n", key)
		for _, nestedKey := range sortedKeys(nested) {
			encoded, err := encodeValue(nested[nestedKey])
			if err != nil {
				return nil, fmt.Errorf("encode TOML key %q.%s: %w", key, nestedKey, err)
			}
			fmt.Fprintf(&output, "%s = %s\n", nestedKey, encoded)
		}
	}

	return []byte(output.String()), nil
}

// Decode parses simple TOML data into the map pointed to by v.
func (Endec) Decode(data []byte, v *map[string]any) error {
	if v == nil {
		return fmt.Errorf("decode TOML: destination map pointer is nil")
	}
	if *v == nil {
		*v = make(map[string]any)
	}

	current := *v
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "["), "]"))
			if section == "" {
				return fmt.Errorf("decode TOML: empty section on line %d", lineNumber)
			}

			current = ensurePath(*v, strings.Split(section, "."))
			continue
		}

		key, rawValue, found := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !found || key == "" {
			return fmt.Errorf("decode TOML: invalid line %d", lineNumber)
		}

		value, err := parseValue(strings.TrimSpace(rawValue))
		if err != nil {
			return fmt.Errorf("decode TOML line %d: %w", lineNumber, err)
		}
		current[key] = value
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("decode TOML: %w", err)
	}
	return nil
}

func ensurePath(root map[string]any, parts []string) map[string]any {
	current := root
	for _, part := range parts {
		part = strings.TrimSpace(part)
		next, ok := current[part].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[part] = next
		}
		current = next
	}
	return current
}

func encodeValue(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return strconv.Quote(v), nil
	case bool:
		return strconv.FormatBool(v), nil
	case int:
		return strconv.Itoa(v), nil
	case int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%v", v), nil
	case []any:
		return encodeArray(v)
	case []string:
		values := make([]any, 0, len(v))
		for _, item := range v {
			values = append(values, item)
		}
		return encodeArray(values)
	default:
		return "", fmt.Errorf("unsupported value type %T", value)
	}
}

func encodeArray(values []any) (string, error) {
	encoded := make([]string, 0, len(values))
	for _, value := range values {
		item, err := encodeValue(value)
		if err != nil {
			return "", err
		}
		encoded = append(encoded, item)
	}
	return "[" + strings.Join(encoded, ", ") + "]", nil
}

func parseValue(value string) (any, error) {
	if value == "" {
		return nil, fmt.Errorf("missing value")
	}
	if strings.HasPrefix(value, `"`) {
		decoded, err := strconv.Unquote(value)
		if err != nil {
			return nil, fmt.Errorf("invalid quoted string: %w", err)
		}
		return decoded, nil
	}
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		return parseArray(value)
	}
	if parsed, err := strconv.ParseBool(value); err == nil {
		return parsed, nil
	}
	if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
		return parsed, nil
	}
	if parsed, err := strconv.ParseFloat(value, 64); err == nil {
		return parsed, nil
	}
	return nil, fmt.Errorf("unsupported value %q", value)
}

func parseArray(value string) ([]any, error) {
	body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
	if body == "" {
		return []any{}, nil
	}

	parts := strings.Split(body, ",")
	values := make([]any, 0, len(parts))
	for _, part := range parts {
		parsed, err := parseValue(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		values = append(values, parsed)
	}
	return values, nil
}

func sortedKeys(v map[string]any) []string {
	keys := make([]string, 0, len(v))
	for key := range v {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
