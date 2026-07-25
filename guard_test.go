package magicseal

import (
	"errors"
	"testing"
)

func TestCheckLimitsPass(t *testing.T) {
	entries := []zipEntry{
		{Name: "a.txt", CompressedSize64: 100, UncompressedSize64: 500},
		{Name: "b.txt", CompressedSize64: 200, UncompressedSize64: 1000},
	}
	if err := checkLimits(entries); err != nil {
		t.Errorf("checkLimits(valid): unexpected error: %v", err)
	}
}

func TestCheckLimitsEmptyEntries(t *testing.T) {
	if err := checkLimits(nil); err != nil {
		t.Errorf("checkLimits(nil): unexpected error: %v", err)
	}
	if err := checkLimits([]zipEntry{}); err != nil {
		t.Errorf("checkLimits(empty): unexpected error: %v", err)
	}
}

func TestCheckLimitsTooManyEntries(t *testing.T) {
	entries := make([]zipEntry, DefaultMaxEntryCount+1)
	for i := range entries {
		entries[i] = zipEntry{Name: "x", CompressedSize64: 1, UncompressedSize64: 1}
	}
	err := checkLimits(entries)
	if err == nil {
		t.Fatal("checkLimits(too many): expected error, got nil")
	}
	if !errors.Is(err, ErrTooManyEntries) {
		t.Errorf("checkLimits(too many): error does not wrap ErrTooManyEntries: %v", err)
	}
}

func TestCheckLimitsTooManyEntriesExact(t *testing.T) {
	entries := make([]zipEntry, DefaultMaxEntryCount)
	for i := range entries {
		entries[i] = zipEntry{Name: "x", CompressedSize64: 1, UncompressedSize64: 1}
	}
	if err := checkLimits(entries); err != nil {
		t.Errorf("checkLimits(exact max): unexpected error: %v", err)
	}
}

func TestCheckLimitsSizeExceeded(t *testing.T) {
	// Two entries that together exceed 256 MB.
	entries := []zipEntry{
		{Name: "a.dat", CompressedSize64: 1, UncompressedSize64: 200 << 20},
		{Name: "b.dat", CompressedSize64: 1, UncompressedSize64: 100 << 20},
	}
	err := checkLimits(entries)
	if err == nil {
		t.Fatal("checkLimits(size exceeded): expected error, got nil")
	}
	if !errors.Is(err, ErrSizeLimitExceeded) {
		t.Errorf("checkLimits(size exceeded): error does not wrap ErrSizeLimitExceeded: %v", err)
	}
}

func TestCheckLimitsSizeExactlyLimit(t *testing.T) {
	// Need realistic compressed size to avoid triggering compression ratio check.
	// 256MB / 100 = ~2.68MB compressed minimum at ratio limit.
	// ceil to avoid integer division rounding making ratio slightly over limit
	minCompressed := uint64(DefaultMaxDecompressedSize)/uint64(DefaultMaxCompressionRatio) + 1
	entries := []zipEntry{
		{Name: "a.dat", CompressedSize64: minCompressed, UncompressedSize64: uint64(DefaultMaxDecompressedSize)},
	}
	if err := checkLimits(entries); err != nil {
		t.Errorf("checkLimits(exact size limit): unexpected error: %v", err)
	}
}

func TestCheckLimitsCompressionRatioBomb(t *testing.T) {
	// 1 byte compressed → 10,000 bytes uncompressed = ratio 10000:1
	entries := []zipEntry{
		{Name: "bomb.bin", CompressedSize64: 1, UncompressedSize64: 10_000},
	}
	err := checkLimits(entries)
	if err == nil {
		t.Fatal("checkLimits(high ratio): expected error, got nil")
	}
	if !errors.Is(err, ErrCompressionRatioExceeded) {
		t.Errorf("checkLimits(high ratio): error does not wrap ErrCompressionRatioExceeded: %v", err)
	}
}

func TestCheckLimitsCompressionRatioBelowLimit(t *testing.T) {
	// Ratio 100:1 is exactly the limit — must not be treated as "over".
	entries := []zipEntry{
		{Name: "max.bin", CompressedSize64: 1, UncompressedSize64: 100},
	}
	if err := checkLimits(entries); err != nil {
		t.Errorf("checkLimits(ratio exactly at limit): unexpected error: %v", err)
	}
}

func TestCheckLimitsCompressedBiggerThanUncompressed(t *testing.T) {
	// Degenerate: compressed size is larger than uncompressed (not a bomb).
	entries := []zipEntry{
		{Name: "bad.bin", CompressedSize64: 1000, UncompressedSize64: 100},
	}
	if err := checkLimits(entries); err != nil {
		t.Errorf("checkLimits(compressed > uncompressed): unexpected error: %v", err)
	}
}

func TestCheckLimitsSkipCompressedZero(t *testing.T) {
	// Folder entries have CompressedSize64 == 0 — should be skipped, not cause divide-by-zero.
	entries := []zipEntry{
		{Name: "folder/", CompressedSize64: 0, UncompressedSize64: 0},
		{Name: "empty.txt", CompressedSize64: 0, UncompressedSize64: 0},
	}
	if err := checkLimits(entries); err != nil {
		t.Errorf("checkLimits(zero compressed): unexpected error (possible divide-by-zero): %v", err)
	}
}

func TestCheckLimitsSizeOverflow(t *testing.T) {
	// Two entries near MaxUint64 — sum overflows uint64.
	max := uint64(^uint64(0)) // MaxUint64
	entries := []zipEntry{
		{Name: "a.bin", CompressedSize64: 1, UncompressedSize64: max - 100},
		{Name: "b.bin", CompressedSize64: 1, UncompressedSize64: 200},
	}
	err := checkLimits(entries)
	if err == nil {
		t.Fatal("checkLimits(overflow): expected error, got nil")
	}
	if !errors.Is(err, ErrSizeLimitExceeded) {
		t.Errorf("checkLimits(overflow): error does not wrap ErrSizeLimitExceeded: %v", err)
	}
}
