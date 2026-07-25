package magicseal

import "errors"

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
)
