package mapper

import (
	"encoding"
	stderrors "errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	zenitherrors "github.com/maneeshaindrachapa/zenith/internal/errors"
)

var (
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	durationType        = reflect.TypeOf(time.Duration(0))
)

type validator interface {
	Validate() error
}

// Options controls how Reflect maps decoded values.
type Options struct {
	Strict                  bool
	PreserveExistingOnEmpty bool
}

// Reflect maps decoded values into structs with reflection.
type Reflect struct {
	Options Options
}

// MapToStruct copies values from data into the struct pointed to by target.
func (mapper Reflect) MapToStruct(data map[string]any, target any) error {
	if target == nil {
		return zenitherrors.NewInvalidTarget("target is nil")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer || targetValue.IsNil() {
		return zenitherrors.NewInvalidTarget("target must be a non-nil pointer to a struct")
	}

	structValue := targetValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return zenitherrors.NewInvalidTarget("target must point to a struct")
	}

	if err := mapper.fillStruct(data, structValue, ""); err != nil {
		return err
	}

	if targetValidator, ok := target.(validator); ok {
		if err := targetValidator.Validate(); err != nil {
			return &zenitherrors.ConfigError{Kind: zenitherrors.ErrValidation, Operation: "validate config", CauseErr: err}
		}
	}

	return nil
}

type mappedValue struct {
	key   string
	value any
}

func (mapper Reflect) fillStruct(data map[string]any, structValue reflect.Value, path string) error {
	values := make(map[string]mappedValue, len(data))
	for key, value := range data {
		values[normalizeKey(key)] = mappedValue{key: key, value: value}
	}

	used := make(map[string]struct{})
	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		fieldType := structType.Field(i)
		fieldValue := structValue.Field(i)
		if fieldType.PkgPath != "" || !fieldValue.CanSet() {
			continue
		}

		fieldPath := joinPath(path, fieldDisplayName(fieldType))
		item, key, found := lookupFieldValue(values, fieldType)
		if !found {
			defaultValue, hasDefault := fieldType.Tag.Lookup("default")
			if hasDefault {
				if err := mapper.assignValue(fieldValue, defaultValue, fieldPath); err != nil {
					return withPath(err, fieldPath)
				}
				continue
			}
			if fieldRequired(fieldType) {
				return zenitherrors.NewRequiredField(fieldPath, "missing")
			}
			continue
		}
		used[key] = struct{}{}

		if mapper.Options.PreserveExistingOnEmpty && isEmptyString(item.value) {
			if fieldRequired(fieldType) {
				return zenitherrors.NewRequiredField(fieldPath, "empty")
			}
			continue
		}
		if fieldRequired(fieldType) && isEmptyString(item.value) {
			return zenitherrors.NewRequiredField(fieldPath, "empty")
		}

		if err := mapper.assignValue(fieldValue, item.value, fieldPath); err != nil {
			return withPath(err, fieldPath)
		}
	}

	if mapper.Options.Strict {
		unknown := unknownKeys(values, used, path)
		if len(unknown) > 0 {
			return zenitherrors.NewUnknownField(unknown[0])
		}
	}

	return nil
}

func lookupFieldValue(values map[string]mappedValue, field reflect.StructField) (mappedValue, string, bool) {
	for _, name := range fieldNames(field) {
		key := normalizeKey(name)
		value, ok := values[key]
		if ok {
			return value, key, true
		}
	}
	return mappedValue{}, "", false
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

func fieldDisplayName(field reflect.StructField) string {
	for _, tag := range []string{"zenith", "json", "env", "mapstructure"} {
		name := strings.Split(field.Tag.Get(tag), ",")[0]
		if name == "-" {
			return field.Name
		}
		if name != "" {
			return name
		}
	}
	return field.Name
}

func fieldRequired(field reflect.StructField) bool {
	if field.Tag.Get("required") == "true" {
		return true
	}
	for _, tag := range []string{"zenith", "validate"} {
		for _, option := range strings.Split(field.Tag.Get(tag), ",") {
			if strings.TrimSpace(option) == "required" {
				return true
			}
		}
	}
	return false
}

func normalizeKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, "-", "")
	return key
}

func joinPath(parent string, field string) string {
	if parent == "" {
		return field
	}
	return parent + "." + field
}

func unknownKeys(values map[string]mappedValue, used map[string]struct{}, path string) []string {
	unknown := make([]string, 0)
	for key, item := range values {
		if _, ok := used[key]; ok {
			continue
		}
		unknown = append(unknown, joinPath(path, item.key))
	}
	sort.Strings(unknown)
	return unknown
}

func isEmptyString(value any) bool {
	text, ok := value.(string)
	return ok && text == ""
}

func withPath(err error, path string) error {
	var configErr *zenitherrors.ConfigError
	if stderrors.As(err, &configErr) {
		if configErr.Path == "" {
			configErr.Path = path
			configErr.Operation = "map field"
		}
		return configErr
	}
	return &zenitherrors.ConfigError{Kind: zenitherrors.ErrInvalidValue, Operation: "map field", Path: path, CauseErr: err}
}

func (mapper Reflect) assignValue(target reflect.Value, value any, path string) error {
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
		return mapper.assignValue(target.Elem(), value, path)
	}

	if target.CanAddr() && target.Addr().Type().Implements(textUnmarshalerType) {
		text, ok := value.(string)
		if !ok {
			text = fmt.Sprint(value)
		}
		if err := target.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(text)); err != nil {
			return &zenitherrors.ConfigError{Kind: zenitherrors.ErrInvalidValue, Path: path, CauseErr: err}
		}
		return nil
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
			return zenitherrors.NewInvalidValue(path, fmt.Sprintf("expected object, got %T", value))
		}
		return mapper.fillStruct(nested, target, path)
	case reflect.Slice:
		return mapper.assignSlice(target, value, path)
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
		return zenitherrors.NewInvalidValue(path, fmt.Sprintf("unsupported target type %s", target.Type()))
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

func (mapper Reflect) assignSlice(target reflect.Value, value any, path string) error {
	source := reflect.ValueOf(value)
	if !source.IsValid() || source.Kind() != reflect.Slice {
		return zenitherrors.NewInvalidValue(path, fmt.Sprintf("expected slice, got %T", value))
	}

	result := reflect.MakeSlice(target.Type(), source.Len(), source.Len())
	for i := 0; i < source.Len(); i++ {
		if err := mapper.assignValue(result.Index(i), source.Index(i).Interface(), fmt.Sprintf("%s[%d]", path, i)); err != nil {
			return &zenitherrors.ConfigError{Kind: zenitherrors.ErrInvalidValue, Operation: "map slice", Path: fmt.Sprintf("%s[%d]", path, i), CauseErr: err}
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
			return false, zenitherrors.NewInvalidValue("", fmt.Sprintf("expected bool, got %q", v))
		}
		return res, nil
	default:
		return false, zenitherrors.NewInvalidValue("", fmt.Sprintf("expected bool, got %T", value))
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
			return 0, zenitherrors.NewInvalidValue("", fmt.Sprintf("expected integer, got %v", v))
		}
		return int64(v), nil
	case string:
		res, err := strconv.ParseInt(strings.TrimSpace(v), 10, bits)
		if err != nil {
			return 0, zenitherrors.NewInvalidValue("", fmt.Sprintf("expected integer, got %q", v))
		}
		return res, nil
	default:
		return 0, zenitherrors.NewInvalidValue("", fmt.Sprintf("expected integer, got %T", value))
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
			return 0, zenitherrors.NewInvalidValue("", fmt.Sprintf("expected unsigned integer, got %v", v))
		}
		return uint64(v), nil
	case string:
		res, err := strconv.ParseUint(strings.TrimSpace(v), 10, bits)
		if err != nil {
			return 0, zenitherrors.NewInvalidValue("", fmt.Sprintf("expected unsigned integer, got %q", v))
		}
		return res, nil
	default:
		return 0, zenitherrors.NewInvalidValue("", fmt.Sprintf("expected unsigned integer, got %T", value))
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
			return 0, zenitherrors.NewInvalidValue("", fmt.Sprintf("expected float, got %q", v))
		}
		return res, nil
	default:
		return 0, zenitherrors.NewInvalidValue("", fmt.Sprintf("expected float, got %T", value))
	}
}

func toDuration(value any) (time.Duration, error) {
	switch v := value.(type) {
	case time.Duration:
		return v, nil
	case string:
		res, err := time.ParseDuration(strings.TrimSpace(v))
		if err != nil {
			return 0, zenitherrors.NewInvalidValue("", fmt.Sprintf("expected duration, got %q", v))
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
