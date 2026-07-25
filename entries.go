package magicseal

import "fmt"

// checkEntries validates that all required entries exist in the ZIP.
// Layer 3: exact-match, case-sensitive comparison against the entry name set.
// Uses map lookup (O(1)) for each required entry against the set of actual entries.
func checkEntries(zipEntries []zipEntry, required []string) error {
	// Build name set from actual entries (O(n), one pass).
	names := make(map[string]struct{}, len(zipEntries))
	for _, e := range zipEntries {
		names[e.Name] = struct{}{}
	}

	// Every required entry must exist with exact match (case-sensitive).
	for _, want := range required {
		if _, ok := names[want]; !ok {
			return fmt.Errorf("%w: %q", ErrMissingEntry, want)
		}
	}
	return nil
}
