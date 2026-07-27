package magicseal

import (
	"archive/zip"
	"fmt"
	"io"
)

// zipEntry holds metadata extracted from a ZIP central directory entry.
type zipEntry struct {
	Name               string
	CompressedSize64   uint64
	UncompressedSize64 uint64
}

// parseZip parses the ZIP central directory from an io.ReaderAt and returns entry metadata.
// Uses metadata-only inspection — entry contents are never opened (no Open() call).
//
// Uses archive/zip.NewReader which parses the central directory via EOCD,
// not just the local file headers. This means prepend-garbage attacks are
// detected (the EOCD offset won't match).
func parseZip(r io.ReaderAt, size int64) ([]zipEntry, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidZip, err)
	}

	entries := make([]zipEntry, 0, len(zr.File))
	for _, f := range zr.File {
		// Metadata-only: read FileHeader fields, never call Open().
		entries = append(entries, zipEntry{
			Name:               f.Name,
			CompressedSize64:   f.CompressedSize64,
			UncompressedSize64: f.UncompressedSize64,
		})
	}
	return entries, nil
}
