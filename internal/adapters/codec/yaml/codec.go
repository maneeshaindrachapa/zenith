package yaml

import (
	"bufio"
	"fmt"
	"sort"
	"strconv"
	"strings"

	zenitherrors "github.com/maneeshaindrachapa/zenith/internal/errors"
)

// Endec encodes and decodes simple YAML mapping data.
type Endec struct{}

// Encode converts a map into simple YAML data.
func (Endec) Encode(v map[string]any) ([]byte, error) {
	var output strings.Builder
	if err := encodeMap(&output, v, 0); err != nil {
		return nil, err
	}
	return []byte(output.String()), nil
}

// Decode parses simple YAML mapping data into the map pointed to by v.
func (Endec) Decode(data []byte, v *map[string]any) error {
	if v == nil {
		return &zenitherrors.ConfigError{Kind: zenitherrors.ErrDecode, Operation: "decode YAML", Reason: "destination map pointer is nil"}
	}
	if *v == nil {
		*v = make(map[string]any)
	}

	type frame struct {
		indent int
		values map[string]any
	}

	stack := []frame{{indent: -1, values: *v}}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		indent := countIndent(raw)
		if indent%2 != 0 {
			return &zenitherrors.ConfigError{Kind: zenitherrors.ErrDecode, Operation: "decode YAML", Reason: fmt.Sprintf("invalid indentation on line %d", lineNumber)}
		}

		for len(stack) > 1 && indent <= stack[len(stack)-1].indent {
			stack = stack[:len(stack)-1]
		}

		key, rawValue, found := strings.Cut(line, ":")
		key = strings.TrimSpace(key)
		rawValue = strings.TrimSpace(rawValue)
		if !found || key == "" {
			return &zenitherrors.ConfigError{Kind: zenitherrors.ErrDecode, Operation: "decode YAML", Reason: fmt.Sprintf("invalid line %d", lineNumber)}
		}

		current := stack[len(stack)-1].values
		if rawValue == "" {
			nested := make(map[string]any)
			current[key] = nested
			stack = append(stack, frame{indent: indent, values: nested})
			continue
		}

		value, err := parseValue(rawValue)
		if err != nil {
			return &zenitherrors.ConfigError{Kind: zenitherrors.ErrDecode, Operation: "decode YAML", Path: key, Reason: fmt.Sprintf("line %d", lineNumber), CauseErr: err}
		}
		current[key] = value
	}

	if err := scanner.Err(); err != nil {
		return zenitherrors.Wrap(zenitherrors.ErrDecode, "decode YAML", err)
	}
	return nil
}

func encodeMap(output *strings.Builder, values map[string]any, indent int) error {
	prefix := strings.Repeat(" ", indent)
	for _, key := range sortedKeys(values) {
		value := values[key]
		if nested, ok := value.(map[string]any); ok {
			fmt.Fprintf(output, "%s%s:\n", prefix, key)
			if err := encodeMap(output, nested, indent+2); err != nil {
				return &zenitherrors.ConfigError{Kind: zenitherrors.ErrEncode, Operation: "encode YAML", Path: key, CauseErr: err}
			}
			continue
		}

		encoded, err := encodeValue(value)
		if err != nil {
			return &zenitherrors.ConfigError{Kind: zenitherrors.ErrEncode, Operation: "encode YAML", Path: key, CauseErr: err}
		}
		fmt.Fprintf(output, "%s%s: %s\n", prefix, key, encoded)
	}
	return nil
}

func countIndent(line string) int {
	count := 0
	for _, char := range line {
		if char != ' ' {
			break
		}
		count++
	}
	return count
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
		return "", &zenitherrors.ConfigError{Kind: zenitherrors.ErrInvalidValue, Reason: fmt.Sprintf("unsupported value type %T", value)}
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
		return nil, &zenitherrors.ConfigError{Kind: zenitherrors.ErrInvalidValue, Reason: "missing value"}
	}
	if strings.HasPrefix(value, `"`) {
		decoded, err := strconv.Unquote(value)
		if err != nil {
			return nil, &zenitherrors.ConfigError{Kind: zenitherrors.ErrInvalidValue, Reason: "invalid quoted string", CauseErr: err}
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
	return value, nil
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
