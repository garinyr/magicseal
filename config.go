package magicseal

import "time"

// ValidatorConfig holds configurable safety limits for validation.
// Zero-value fields use safe defaults via the Defaults() method.
type ValidatorConfig struct {
	// MaxDecompressedSize is the maximum total decompressed size in bytes.
	// Default: DefaultMaxDecompressedSize (256 MB).
	MaxDecompressedSize int64

	// MaxCompressionRatio is the maximum allowed ratio of uncompressed to compressed size.
	// Default: DefaultMaxCompressionRatio (100.0).
	MaxCompressionRatio float64

	// MaxFileNameLength is the maximum length of a ZIP entry name.
	// Default: 1024.
	MaxFileNameLength int

	// MaxTotalEntries is the maximum number of entries allowed in a ZIP.
	// Default: DefaultMaxEntryCount (10,000).
	MaxTotalEntries int

	// ParseTimeout is an experimental timeout for ZIP parsing.
	// Default: 0 (no timeout). See README for goroutine-leak caveat.
	ParseTimeout time.Duration
}

// Defaults returns a copy of cfg with zero-value fields replaced by safe defaults.
// If cfg is nil, returns a fully defaulted config.
func (cfg *ValidatorConfig) Defaults() *ValidatorConfig {
	if cfg == nil {
		return &ValidatorConfig{
			MaxDecompressedSize: DefaultMaxDecompressedSize,
			MaxCompressionRatio: DefaultMaxCompressionRatio,
			MaxFileNameLength:   1024,
			MaxTotalEntries:     DefaultMaxEntryCount,
		}
	}

	// Return a new copy so callers can't mutate the original via the result.
	result := *cfg
	if result.MaxDecompressedSize <= 0 {
		result.MaxDecompressedSize = DefaultMaxDecompressedSize
	}
	if result.MaxCompressionRatio <= 0 {
		result.MaxCompressionRatio = DefaultMaxCompressionRatio
	}
	if result.MaxFileNameLength <= 0 {
		result.MaxFileNameLength = 1024
	}
	if result.MaxTotalEntries <= 0 {
		result.MaxTotalEntries = DefaultMaxEntryCount
	}
	return &result
}
