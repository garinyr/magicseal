package magicseal

import (
	"fmt"
	"math"
)

const (
	// DefaultMaxEntryCount is the maximum number of ZIP entries allowed.
	DefaultMaxEntryCount = 10_000

	// DefaultMaxDecompressedSize is the maximum total decompressed size in bytes (256 MB).
	DefaultMaxDecompressedSize = 256 << 20

	// DefaultMaxCompressionRatio is the maximum allowed ratio of uncompressed to compressed size.
	DefaultMaxCompressionRatio = 100.0
)

// checkLimits validates the ZIP entries against safety limits:
// entry count, total decompressed size, and per-entry compression ratio.
//
// Guard runs immediately after Layer 2 (ZIP parse), before Layer 3 (required entries).
// This is defense-in-depth: files that clearly violate limits are rejected early.
func checkLimits(zipEntries []zipEntry) error {
	// 1. Entry count guard
	if len(zipEntries) > DefaultMaxEntryCount {
		return fmt.Errorf("%w: %d entries exceeds limit %d",
			ErrTooManyEntries, len(zipEntries), DefaultMaxEntryCount)
	}

	// 2. Total decompressed size guard
	var total uint64
	for _, e := range zipEntries {
		// Overflow guard: if adding would overflow uint64, reject immediately.
		if total > math.MaxUint64-e.UncompressedSize64 {
			return fmt.Errorf("%w: total decompressed size overflows uint64",
				ErrSizeLimitExceeded)
		}
		total += e.UncompressedSize64
		if total > uint64(DefaultMaxDecompressedSize) {
			return fmt.Errorf("%w: total decompressed size %d exceeds limit %d",
				ErrSizeLimitExceeded, total, DefaultMaxDecompressedSize)
		}
	}

	// 3. Compression ratio guard (per entry)
	for _, e := range zipEntries {
		if e.CompressedSize64 == 0 {
			// Skip: folder entries or empty files have no compressed size.
			// Divide-by-zero is avoided here.
			continue
		}
		// Also guard against compressed > uncompressed (degenerate case, not a bomb)
		if e.CompressedSize64 > e.UncompressedSize64 {
			continue
		}
		ratio := float64(e.UncompressedSize64) / float64(e.CompressedSize64)
		if ratio > DefaultMaxCompressionRatio {
			return fmt.Errorf("%w: entry %q has compression ratio %.1f (limit %.1f)",
				ErrCompressionRatioExceeded, e.Name, ratio, DefaultMaxCompressionRatio)
		}
	}

	return nil
}
