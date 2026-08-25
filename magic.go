package magicseal

import (
	"bytes"
	"fmt"
	"io"
)

// checkMagic validates the magic bytes at the start of data.
// Layer 1: fast-fail for files that are clearly wrong.
// Returns nil if the first len(f.MagicBytes) bytes of data match f.MagicBytes.
func checkMagic(r io.ReaderAt, size int64, f Format) error {
	magicLen := int64(len(f.MagicBytes))
	if size < magicLen {
		return fmt.Errorf("%w: data too short (%d bytes), expected at least %d",
			ErrMagicMismatch, size, magicLen)
	}
	buf := make([]byte, magicLen)
	if _, err := r.ReadAt(buf, 0); err != nil {
		return fmt.Errorf("%w: failed to read magic bytes: %v", ErrMagicMismatch, err)
	}
	if !bytes.HasPrefix(buf, f.MagicBytes) {
		return fmt.Errorf("%w: expected %x, got %x",
			ErrMagicMismatch, f.MagicBytes, buf)
	}
	return nil
}
