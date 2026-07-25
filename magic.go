package magicseal

import (
	"bytes"
	"fmt"
)

// checkMagic validates the magic bytes at the start of data.
// Layer 1: fast-fail for files that are clearly wrong.
// Returns nil if the first len(f.MagicBytes) bytes of data match f.MagicBytes.
func checkMagic(data []byte, f Format) error {
	if len(data) < len(f.MagicBytes) {
		return fmt.Errorf("%w: data too short (%d bytes), expected at least %d",
			ErrMagicMismatch, len(data), len(f.MagicBytes))
	}
	if !bytes.HasPrefix(data, f.MagicBytes) {
		return fmt.Errorf("%w: expected %x, got %x",
			ErrMagicMismatch, f.MagicBytes, data[:len(f.MagicBytes)])
	}
	return nil
}
