package magicseal

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"testing"
)

// makeMinimalZip creates a minimal valid ZIP file in memory with the given entry names.
// Entries are stored (no compression) with empty content.
func makeMinimalZip(entryNames []string) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, name := range entryNames {
		w.Create(name)
	}
	w.Close()
	return buf.Bytes()
}

func TestParseZipValid(t *testing.T) {
	data := makeMinimalZip([]string{"word/document.xml", "_rels/.rels", "[Content_Types].xml"})
	entries, err := parseZip(data)
	if err != nil {
		t.Fatalf("parseZip(valid ZIP): unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("parseZip: got %d entries, want 3", len(entries))
	}
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name] = true
	}
	for _, want := range []string{"word/document.xml", "_rels/.rels", "[Content_Types].xml"} {
		if !names[want] {
			t.Errorf("parseZip: missing entry %q", want)
		}
	}
}

func TestParseZipEmptyFile(t *testing.T) {
	_, err := parseZip([]byte{})
	if err == nil {
		t.Fatal("parseZip(empty): expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidZip) {
		t.Errorf("parseZip(empty): error does not wrap ErrInvalidZip: %v", err)
	}
}

func TestParseZipNotZipData(t *testing.T) {
	data := []byte("not a zip file at all")
	_, err := parseZip(data)
	if err == nil {
		t.Fatal("parseZip(non-ZIP): expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidZip) {
		t.Errorf("parseZip(non-ZIP): error does not wrap ErrInvalidZip: %v", err)
	}
}

func TestParseZipTruncated(t *testing.T) {
	// Take a valid ZIP and truncate it — simulate interrupted upload.
	full := makeMinimalZip([]string{"a", "b", "c"})
	truncated := full[:len(full)-20] // chop off the central directory
	_, err := parseZip(truncated)
	if err == nil {
		t.Fatal("parseZip(truncated): expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidZip) {
		t.Errorf("parseZip(truncated): error does not wrap ErrInvalidZip: %v", err)
	}
}

func TestParseZipPrependGarbage(t *testing.T) {
	// Prepend garbage bytes before a valid ZIP — simulates prepend attack.
	valid := makeMinimalZip([]string{"readme.txt"})
	garbage := append([]byte("junk data here "), valid...)
	entries, err := parseZip(garbage)
	// archive/zip.NewReader uses EOCD at the end, so prepend garbage
	// should still parse correctly. This confirms central-directory-based
	// parsing is resilient to prepend attacks.
	if err != nil {
		t.Errorf("parseZip(prepend garbage): archive/zip correctly ignores prepended data: %v", err)
	} else {
		if len(entries) != 1 || entries[0].Name != "readme.txt" {
			t.Errorf("parseZip(prepend garbage): expected 1 entry 'readme.txt', got %v", entries)
		}
	}
}

func TestParseZipCorruptCentralDirectory(t *testing.T) {
	// Corrupt the central directory portion (tail of file) by overwriting bytes.
	valid := makeMinimalZip([]string{"a", "b"})
	corrupt := make([]byte, len(valid))
	copy(corrupt, valid)
	// Scramble the last 10 bytes (EOCD / central directory area)
	for i := len(corrupt) - 10; i < len(corrupt); i++ {
		corrupt[i] ^= 0xFF
	}
	_, err := parseZip(corrupt)
	if err == nil {
		t.Fatal("parseZip(corrupt central dir): expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidZip) {
		t.Errorf("parseZip(corrupt central dir): error does not wrap ErrInvalidZip: %v", err)
	}
}

func TestParseZipPreservesSizes(t *testing.T) {
	// Create a ZIP with a known content to verify size metadata.
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	fw, _ := w.Create("hello.txt")
	content := []byte("hello world")
	fw.Write(content)
	w.Close()

	entries, err := parseZip(buf.Bytes())
	if err != nil {
		t.Fatalf("parseZip: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].UncompressedSize64 != uint64(len(content)) {
		t.Errorf("UncompressedSize64: got %d, want %d", entries[0].UncompressedSize64, len(content))
	}
}

// Helper for tests that need actual OS files (Phase 1.8).
func makeZipFile(t *testing.T, path string, entryNames []string) {
	t.Helper()
	data := makeMinimalZip(entryNames)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("failed to write test fixture %s: %v", path, err)
	}
	t.Cleanup(func() { os.Remove(path) })
}
