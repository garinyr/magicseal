package magicseal

import "os"

// Validate checks whether data is a valid file of the claimed format.
//
// Validation runs in 3 layers:
//  1. Magic bytes check — fast-fail for non-ZIP files
//  2. ZIP structure parsing — verifies central directory integrity
//  3. Required entries check — ensures mandatory files exist for the format
//
// Guard limits (entry count, decompressed size, compression ratio)
// are checked between layers 2 and 3 as defense-in-depth.
func Validate(data []byte, format FormatName) error {
	f, err := lookupFormat(format)
	if err != nil {
		return err
	}
	return validate(data, f)
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
