package errors

import (
	"errors"
	"fmt"
)

// ErrUnsupportedFormat reports that no codec is registered for a format.
var ErrUnsupportedFormat = errors.New("unsupported format")

// ErrUnknownField reports decoded data that does not map to a struct field.
var ErrUnknownField = errors.New("unknown field")

// ErrRequiredField reports a required struct field with no usable value.
var ErrRequiredField = errors.New("required field")

// ErrMissingExtension reports a config file path with no extension.
var ErrMissingExtension = errors.New("missing file extension")

// ErrReadFile reports a file read failure.
var ErrReadFile = errors.New("read file")

// ErrEncode reports an encoding failure.
var ErrEncode = errors.New("encode")

// ErrDecode reports a decoding failure.
var ErrDecode = errors.New("decode")

// ErrInvalidTarget reports an invalid mapping target.
var ErrInvalidTarget = errors.New("invalid target")

// ErrInvalidValue reports an invalid or unsupported config value.
var ErrInvalidValue = errors.New("invalid value")

// ErrValidation reports a user validation failure.
var ErrValidation = errors.New("validation")

// ErrRegistry reports invalid codec registry usage.
var ErrRegistry = errors.New("registry")

// Kinded exposes the sentinel error kind wrapped by ConfigError.
type Kinded interface {
	ErrorKind() error
}

// FieldPathed exposes the config field path related to an error.
type FieldPathed interface {
	FieldPath() string
}

// Formatted exposes the config format related to an error.
type Formatted interface {
	FormatName() string
}

// Reasoned exposes a short reason related to an error.
type Reasoned interface {
	ErrorReason() string
}

// Caused exposes the lower-level error wrapped by ConfigError.
type Caused interface {
	Cause() error
}

// ConfigError is the single structured error type used by Zenith.
type ConfigError struct {
	Kind      error
	Path      string
	Format    string
	Reason    string
	Operation string
	CauseErr  error
}

// NewUnsupportedFormat creates an unsupported-format ConfigError.
func NewUnsupportedFormat(format string) *ConfigError {
	return &ConfigError{Kind: ErrUnsupportedFormat, Format: format}
}

// NewUnknownField creates an unknown-field ConfigError.
func NewUnknownField(path string) *ConfigError {
	return &ConfigError{Kind: ErrUnknownField, Path: path}
}

// NewRequiredField creates a required-field ConfigError.
func NewRequiredField(path string, reason string) *ConfigError {
	return &ConfigError{Kind: ErrRequiredField, Path: path, Reason: reason}
}

// NewInvalidTarget creates an invalid-target ConfigError.
func NewInvalidTarget(reason string) *ConfigError {
	return &ConfigError{Kind: ErrInvalidTarget, Reason: reason}
}

// NewInvalidValue creates an invalid-value ConfigError.
func NewInvalidValue(path string, reason string) *ConfigError {
	return &ConfigError{Kind: ErrInvalidValue, Path: path, Reason: reason}
}

// Wrap creates a ConfigError that preserves a lower-level cause.
func Wrap(kind error, operation string, cause error) *ConfigError {
	return &ConfigError{Kind: kind, Operation: operation, CauseErr: cause}
}

func (e *ConfigError) Error() string {
	message := e.message()
	if e.Operation != "" {
		if e.Path != "" {
			message = e.Operation + " " + e.Path + ": " + message
		} else {
			message = e.Operation + ": " + message
		}
	}
	if e.CauseErr != nil {
		message = message + ": " + e.CauseErr.Error()
	}
	return message
}

func (e *ConfigError) message() string {
	switch e.Kind {
	case ErrUnsupportedFormat:
		return fmt.Sprintf("codec for %q is not registered", e.Format)
	case ErrUnknownField:
		return fmt.Sprintf("unknown field %s", e.Path)
	case ErrRequiredField:
		if e.Reason == "" {
			return fmt.Sprintf("required field %s", e.Path)
		}
		return fmt.Sprintf("required field %s: %s", e.Path, e.Reason)
	case ErrMissingExtension, ErrReadFile, ErrEncode, ErrDecode, ErrInvalidTarget, ErrInvalidValue, ErrValidation, ErrRegistry:
		if e.Reason != "" {
			if e.Path != "" {
				return e.Kind.Error() + " " + e.Path + ": " + e.Reason
			}
			return e.Kind.Error() + ": " + e.Reason
		}
		if e.Path != "" {
			return e.Kind.Error() + " " + e.Path
		}
		return e.Kind.Error()
	default:
		return "config error"
	}
}

func (e *ConfigError) Unwrap() []error {
	errs := make([]error, 0, 2)
	if e.Kind != nil {
		errs = append(errs, e.Kind)
	}
	if e.CauseErr != nil {
		errs = append(errs, e.CauseErr)
	}
	return errs
}

func (e *ConfigError) ErrorKind() error {
	return e.Kind
}

func (e *ConfigError) FieldPath() string {
	return e.Path
}

func (e *ConfigError) FormatName() string {
	return e.Format
}

func (e *ConfigError) ErrorReason() string {
	return e.Reason
}

func (e *ConfigError) Cause() error {
	return e.CauseErr
}
