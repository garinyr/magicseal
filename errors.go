package magicseal

import (
	"errors"
	"fmt"
)

// Sentinel errors — use errors.Is() for stable comparison across versions.
var (
	// ErrUnsupportedFormat is returned when the format name is not registered.
	ErrUnsupportedFormat = errors.New("magicseal: unsupported format")

	// ErrMagicMismatch is returned when magic bytes do not match the claimed format.
	ErrMagicMismatch = errors.New("magicseal: magic bytes mismatch")

	// ErrInvalidZip is returned when the data is not a valid ZIP archive.
	ErrInvalidZip = errors.New("magicseal: invalid zip structure")

	// ErrMissingEntry is returned when a required ZIP entry is absent.
	ErrMissingEntry = errors.New("magicseal: missing required entry")

	// ErrSizeLimitExceeded is returned when total decompressed size exceeds the limit.
	ErrSizeLimitExceeded = errors.New("magicseal: decompressed size limit exceeded")

	// ErrCompressionRatioExceeded is returned when the compression ratio is suspicious.
	ErrCompressionRatioExceeded = errors.New("magicseal: compression ratio limit exceeded")

	// ErrTooManyEntries is returned when the ZIP contains more entries than allowed.
	ErrTooManyEntries = errors.New("magicseal: entry count limit exceeded")

	// ErrFormatAlreadyRegistered is returned when attempting to register a duplicate format.
	ErrFormatAlreadyRegistered = errors.New("magicseal: format already registered")

	// ErrFormatNotRegistered is returned when attempting to unregister an unknown format.
	ErrFormatNotRegistered = errors.New("magicseal: format not registered")

	// ErrPathTraversal is returned when a ZIP entry name contains suspicious
	// patterns such as ../, ..\, absolute paths, or null byte injection.
	ErrPathTraversal = errors.New("magicseal: path traversal in entry name")
)

// ErrorCode is a machine-readable error classification.
type ErrorCode int

const (
	CodeUnsupportedFormat        ErrorCode = iota + 1
	CodeMagicMismatch
	CodeInvalidZip
	CodeMissingEntry
	CodeSizeLimitExceeded
	CodeCompressionRatioExceeded
	CodeTooManyEntries
	CodePathTraversal
	CodeFormatAlreadyRegistered
	CodeFormatNotRegistered
)

// mapSentinelToCode returns the ErrorCode for a given sentinel error, or 0 if unknown.
func mapSentinelToCode(err error) ErrorCode {
	switch {
	case errors.Is(err, ErrUnsupportedFormat):
		return CodeUnsupportedFormat
	case errors.Is(err, ErrMagicMismatch):
		return CodeMagicMismatch
	case errors.Is(err, ErrInvalidZip):
		return CodeInvalidZip
	case errors.Is(err, ErrMissingEntry):
		return CodeMissingEntry
	case errors.Is(err, ErrSizeLimitExceeded):
		return CodeSizeLimitExceeded
	case errors.Is(err, ErrCompressionRatioExceeded):
		return CodeCompressionRatioExceeded
	case errors.Is(err, ErrTooManyEntries):
		return CodeTooManyEntries
	case errors.Is(err, ErrPathTraversal):
		return CodePathTraversal
	case errors.Is(err, ErrFormatAlreadyRegistered):
		return CodeFormatAlreadyRegistered
	case errors.Is(err, ErrFormatNotRegistered):
		return CodeFormatNotRegistered
	default:
		return 0
	}
}

// ValidationError wraps a sentinel error with structured metadata.
// Use errors.Is() for sentinel comparison (backward compatible with MVP 1).
// Use errors.As() to extract Code, Format, and Details.
type ValidationError struct {
	Code    ErrorCode
	Format  FormatName
	Details string
	Err     error
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%v [format=%s] [code=%d]: %s",
			e.Err, e.Format, e.Code, e.Details)
	}
	return fmt.Sprintf("%v [format=%s] [code=%d]", e.Err, e.Format, e.Code)
}

// Unwrap returns the wrapped sentinel error, enabling errors.Is() backward compatibility.
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// newValidationError creates a ValidationError wrapping the given sentinel.
// It auto-maps the code and preserves the full error chain via wrapping.
func newValidationError(sentinel error, format FormatName, details string) error {
	return &ValidationError{
		Code:    mapSentinelToCode(sentinel),
		Format:  format,
		Details: details,
		Err:     sentinel,
	}
}
