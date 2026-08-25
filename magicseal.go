package magicseal

import (
	"bytes"
	"io"
	"os"
)

// Validate checks whether data is a valid file of the claimed format.
//
// Validation runs in 3 layers:
//  1. Magic bytes check — fast-fail for non-ZIP files
//  2. ZIP structure parsing — verifies central directory integrity
//  3. Required entries check — ensures mandatory files exist for the format
//
// Guard limits (entry count, decompressed size, compression ratio, path traversal)
// are checked between layers 2 and 3 as defense-in-depth.
//
// Validate is a convenience wrapper around ValidateReader that reads data into memory.
// Uses safe default limits. For custom limits, use ValidateWithConfig.
// For files >100MB, use ValidateReader directly to avoid loading the entire file.
func Validate(data []byte, format FormatName) error {
	return ValidateWithConfig(data, format, nil)
}

// ValidateReader validates a file accessed via io.ReaderAt against the claimed format.
// size must be the total file size in bytes.
// Uses safe default limits.
//
// Use this for large files (>100MB) to avoid loading the entire file into memory.
// The reader must remain valid for the duration of the call.
func ValidateReader(r io.ReaderAt, size int64, format FormatName) error {
	return ValidateReaderWithConfig(r, size, format, nil)
}

// ValidateWithConfig checks data against the format using custom safety limits.
// If cfg is nil, safe defaults are used (see ValidatorConfig.Defaults).
func ValidateWithConfig(data []byte, format FormatName, cfg *ValidatorConfig) error {
	f, err := lookupFormat(format)
	if err != nil {
		return newValidationError(err, format, "")
	}
	return validate(bytes.NewReader(data), int64(len(data)), f, cfg)
}

// ValidateReaderWithConfig validates a reader against the format using custom safety limits.
// If cfg is nil, safe defaults are used.
func ValidateReaderWithConfig(r io.ReaderAt, size int64, format FormatName, cfg *ValidatorConfig) error {
	f, err := lookupFormat(format)
	if err != nil {
		return newValidationError(err, format, "")
	}
	return validate(r, size, f, cfg)
}

// ValidateFromPath reads the file at path and validates it against the claimed format.
// It delegates to Validate after reading the file contents.
func ValidateFromPath(path string, format FormatName) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return Validate(data, format)
}
