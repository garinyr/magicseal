package magicseal

import "io"

// validate orchestrates the layered validation: Layer 1 → Layer 2 → guards → Layer 3.
// Each layer fails fast — if a check fails, subsequent layers are skipped.
// All errors are wrapped in *ValidationError for structured error handling.
func validate(r io.ReaderAt, size int64, f Format, cfg *ValidatorConfig) error {
	cfg = cfg.Defaults()

	// Layer 1: Magic bytes check
	if err := checkMagic(r, size, f); err != nil {
		return newValidationError(err, f.Name, "")
	}

	// Layer 2: ZIP structure parsing (metadata-only)
	entries, err := parseZip(r, size)
	if err != nil {
		return newValidationError(err, f.Name, "")
	}

	// Guard: Safety limits — runs between Layer 2 and Layer 3 (defense-in-depth)
	if err := checkLimits(entries, cfg); err != nil {
		return newValidationError(err, f.Name, "")
	}

	// Guard: Path traversal — check all entry names for suspicious patterns
	if err := checkPathTraversal(entries); err != nil {
		return newValidationError(err, f.Name, "")
	}

	// Layer 3: Required entries check
	if err := checkEntries(entries, f.RequiredEntries); err != nil {
		return newValidationError(err, f.Name, "")
	}

	return nil
}
