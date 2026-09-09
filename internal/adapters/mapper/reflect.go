package mapper

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	durationType        = reflect.TypeOf(time.Duration(0))
)

// Reflect maps decoded values into structs with reflection.
type Reflect struct{}

// MapToStruct copies values from data into the struct pointed to by target.
func (Reflect) MapToStruct(data map[string]any, target any) error {
	if target == nil {
		return fmt.Errorf("map to struct: target is nil")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer || targetValue.IsNil() {
		return fmt.Errorf("map to struct: target must be a non-nil pointer to a struct")
	}

	structValue := targetValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("map to struct: target must point to a struct")
	}

	return fillStruct(data, structValue)
}

func fillStruct(data map[string]any, structValue reflect.Value) error {
	values := make(map[string]any, len(data))
	for key, value := range data {
		values[normalizeKey(key)] = value
	}

	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		fieldType := structType.Field(i)
		fieldValue := structValue.Field(i)
		if fieldType.PkgPath != "" || !fieldValue.CanSet() {
			continue
		}

		value, found := lookupFieldValue(values, fieldType)
		if !found {
			continue
		}

		if err := assignValue(fieldValue, value); err != nil {
			return fmt.Errorf("map field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

func lookupFieldValue(values map[string]any, field reflect.StructField) (any, bool) {
	for _, name := range fieldNames(field) {
		value, ok := values[normalizeKey(name)]
		if ok {
			return value, true
		}
	}
	return nil, false
}

func fieldNames(field reflect.StructField) []string {
	names := []string{field.Name}
	for _, tag := range []string{"zenith", "json", "env", "mapstructure"} {
		name := strings.Split(field.Tag.Get(tag), ",")[0]
		if name == "-" {
			return nil
		}
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func normalizeKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, "-", "")
	return key
}

func assignValue(target reflect.Value, value any) error {
	if !target.CanSet() {
		return nil
	}
	if value == nil {
		return nil
	}

	if target.Kind() == reflect.Pointer {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		return assignValue(target.Elem(), value)
	}

	if target.CanAddr() && target.Addr().Type().Implements(textUnmarshalerType) {
		text, ok := value.(string)
		if !ok {
			text = fmt.Sprint(value)
		}
		return target.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(text))
	}

	if target.Type() == durationType {
		duration, err := toDuration(value)
		if err != nil {
			return err
		}
		target.SetInt(int64(duration))
		return nil
	}

	source := reflect.ValueOf(value)
	if source.IsValid() && source.Type().AssignableTo(target.Type()) {
		target.Set(source)
		return nil
	}
	if source.IsValid() && source.Type().ConvertibleTo(target.Type()) && valueCanConvertSafely(target, source) {
		target.Set(source.Convert(target.Type()))
		return nil
	}

	switch target.Kind() {
	case reflect.Struct:
		nested, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
		return fillStruct(nested, target)
	case reflect.Slice:
		return assignSlice(target, value)
	case reflect.String:
		target.SetString(fmt.Sprint(value))
	case reflect.Bool:
		boolValue, err := toBool(value)
		if err != nil {
			return err
		}
		target.SetBool(boolValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := toInt(value, target.Type().Bits())
		if err != nil {
			return err
		}
		target.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		uintValue, err := toUint(value, target.Type().Bits())
		if err != nil {
			return err
		}
		target.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := toFloat(value, target.Type().Bits())
		if err != nil {
			return err
		}
		target.SetFloat(floatValue)
	default:
		return fmt.Errorf("unsupported target type %s", target.Type())
	}

	return nil
}

func valueCanConvertSafely(target reflect.Value, source reflect.Value) bool {
	if isScalarKind(target.Kind()) || isScalarKind(source.Kind()) {
		return false
	}
	return true
}

func isScalarKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

func assignSlice(target reflect.Value, value any) error {
	source := reflect.ValueOf(value)
	if !source.IsValid() || source.Kind() != reflect.Slice {
		return fmt.Errorf("expected slice, got %T", value)
	}

	result := reflect.MakeSlice(target.Type(), source.Len(), source.Len())
	for i := 0; i < source.Len(); i++ {
		if err := assignValue(result.Index(i), source.Index(i).Interface()); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	target.Set(result)
	return nil
}

func toBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		res, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return false, fmt.Errorf("expected bool, got %q", v)
		}
		return res, nil
	default:
		return false, fmt.Errorf("expected bool, got %T", value)
	}
}

func toInt(value any, bits int) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		if v != float64(int64(v)) {
			return 0, fmt.Errorf("expected integer, got %v", v)
		}
		return int64(v), nil
	case string:
		res, err := strconv.ParseInt(strings.TrimSpace(v), 10, bits)
		if err != nil {
			return 0, fmt.Errorf("expected integer, got %q", v)
		}
		return res, nil
	default:
		return 0, fmt.Errorf("expected integer, got %T", value)
	}
}

func toUint(value any, bits int) (uint64, error) {
	switch v := value.(type) {
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case float64:
		if v < 0 || v != float64(uint64(v)) {
			return 0, fmt.Errorf("expected unsigned integer, got %v", v)
		}
		return uint64(v), nil
	case string:
		res, err := strconv.ParseUint(strings.TrimSpace(v), 10, bits)
		if err != nil {
			return 0, fmt.Errorf("expected unsigned integer, got %q", v)
		}
		return res, nil
	default:
		return 0, fmt.Errorf("expected unsigned integer, got %T", value)
	}
}

func toFloat(value any, bits int) (float64, error) {
	switch v := value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		res, err := strconv.ParseFloat(strings.TrimSpace(v), bits)
		if err != nil {
			return 0, fmt.Errorf("expected float, got %q", v)
		}
		return res, nil
	default:
		return 0, fmt.Errorf("expected float, got %T", value)
	}
}

func toDuration(value any) (time.Duration, error) {
	switch v := value.(type) {
	case time.Duration:
		return v, nil
	case string:
		res, err := time.ParseDuration(strings.TrimSpace(v))
		if err != nil {
			return 0, fmt.Errorf("expected duration, got %q", v)
		}
		return res, nil
	default:
		intValue, err := toInt(value, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(intValue), nil
	}
}
