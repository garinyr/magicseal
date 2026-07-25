package magicseal

// validate orchestrates the layered validation: Layer 1 → Layer 2 → guard → Layer 3.
// Each layer fails fast — if a check fails, subsequent layers are skipped.
func validate(data []byte, f Format) error {
	// Layer 1: Magic bytes check
	if err := checkMagic(data, f); err != nil {
		return err
	}

	// Layer 2: ZIP structure parsing (metadata-only)
	entries, err := parseZip(data)
	if err != nil {
		return err
	}

	// Guard: Safety limits — runs between Layer 2 and Layer 3 (defense-in-depth)
	if err := checkLimits(entries); err != nil {
		return err
	}

	// Layer 3: Required entries check
	if err := checkEntries(entries, f.RequiredEntries); err != nil {
		return err
	}

	return nil
}
