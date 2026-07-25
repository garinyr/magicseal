package magicseal

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"sync"
	"testing"
)

// --- 1.8.4 Integration tests: all "Definisi Selesai" scenarios ---

func TestIntegration_ValidFormats(t *testing.T) {
	tests := []struct {
		name   FormatName
		format Format
	}{
		{FormatDOCX, makeFormatDOCX()},
		{FormatXLSX, makeFormatXLSX()},
		{FormatPPTX, makeFormatPPTX()},
	}
	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			data := makeMinimalZip(tt.format.RequiredEntries)
			if err := Validate(data, tt.name); err != nil {
				t.Errorf("Validate(valid %s): unexpected error: %v", tt.name, err)
			}
		})
	}
}

func TestIntegration_RenamedZip(t *testing.T) {
	// A plain ZIP with non-Office entries claimed as DOCX — Layer 3 must reject.
	data := makeMinimalZip([]string{"notes.txt", "archive.zip"})
	err := Validate(data, FormatDOCX)
	if err == nil {
		t.Fatal("Validate(renamed ZIP as DOCX): expected error, got nil")
	}
	if !errors.Is(err, ErrMissingEntry) {
		t.Errorf("expected ErrMissingEntry, got: %v", err)
	}
}

func TestIntegration_CorruptCentralDirectory(t *testing.T) {
	valid := makeMinimalZip(makeFormatDOCX().RequiredEntries)
	corrupt := make([]byte, len(valid))
	copy(corrupt, valid)
	// Flip bits in the last 15 bytes (EOCD area)
	for i := len(corrupt) - 15; i < len(corrupt); i++ {
		corrupt[i] ^= 0xFF
	}
	err := Validate(corrupt, FormatDOCX)
	if err == nil {
		t.Fatal("Validate(corrupt central dir): expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidZip) {
		t.Errorf("expected ErrInvalidZip, got: %v", err)
	}
}

func TestIntegration_MissingEntry(t *testing.T) {
	// Build a ZIP that has all but one required entry removed.
	required := makeFormatDOCX().RequiredEntries
	data := makeMinimalZip(required[:len(required)-1]) // drop last entry
	err := Validate(data, FormatDOCX)
	if err == nil {
		t.Fatal("Validate(missing entry): expected error, got nil")
	}
	if !errors.Is(err, ErrMissingEntry) {
		t.Errorf("expected ErrMissingEntry, got: %v", err)
	}
}

func TestIntegration_ZipBombSmall(t *testing.T) {
	// Build a ZIP with a single entry that has ratio 1000:1.
	// Use archive/zip directly to create a stored file with manipulated sizes.
	// Since archive/zip.Writer computes real sizes, we create a valid ZIP
	// and verify that high-ratio files are caught by checkLimits.
	entries := []zipEntry{
		{Name: "a.xml", CompressedSize64: 1, UncompressedSize64: 10_000},
		{Name: "b.xml", CompressedSize64: 1, UncompressedSize64: 10_000},
	}
	// We test checkLimits directly for ratio — Validate() integration
	// with a real zip bomb is tested at the guard level.
	err := checkLimits(entries)
	if err == nil {
		t.Fatal("checkLimits(zip bomb ratio): expected error, got nil")
	}
	if !errors.Is(err, ErrCompressionRatioExceeded) {
		t.Errorf("expected ErrCompressionRatioExceeded, got: %v", err)
	}
}

func TestIntegration_TooManyEntries(t *testing.T) {
	entries := make([]zipEntry, DefaultMaxEntryCount+1)
	for i := range entries {
		entries[i] = zipEntry{Name: "x", CompressedSize64: 1, UncompressedSize64: 1}
	}
	err := checkLimits(entries)
	if err == nil {
		t.Fatal("checkLimits(>10k entries): expected error, got nil")
	}
	if !errors.Is(err, ErrTooManyEntries) {
		t.Errorf("expected ErrTooManyEntries, got: %v", err)
	}
}

func TestIntegration_NonZip(t *testing.T) {
	data := []byte("this is plain text, not a ZIP file at all")
	err := Validate(data, FormatDOCX)
	if err == nil {
		t.Fatal("Validate(plain text): expected error, got nil")
	}
	if !errors.Is(err, ErrMagicMismatch) {
		t.Errorf("expected ErrMagicMismatch, got: %v", err)
	}
}

func TestIntegration_EmptyFile(t *testing.T) {
	err := Validate(nil, FormatDOCX)
	if err == nil {
		t.Fatal("Validate(nil): expected error, got nil")
	}
	if !errors.Is(err, ErrMagicMismatch) {
		t.Errorf("expected ErrMagicMismatch, got: %v", err)
	}
}

func TestIntegration_TruncatedZip(t *testing.T) {
	full := makeMinimalZip(makeFormatDOCX().RequiredEntries)
	// Cut off the central directory (last ~40 bytes or so)
	truncateAt := len(full) - 10
	if truncateAt < 4 {
		t.Fatal("fixture too small to truncate")
	}
	data := full[:truncateAt]
	err := Validate(data, FormatDOCX)
	if err == nil {
		t.Fatal("Validate(truncated): expected error, got nil")
	}
	// Must not panic — error should wrap ErrMagicMismatch or ErrInvalidZip
	if !errors.Is(err, ErrMagicMismatch) && !errors.Is(err, ErrInvalidZip) {
		t.Errorf("expected ErrMagicMismatch or ErrInvalidZip, got: %v", err)
	}
}

func TestIntegration_CaseMismatchEntries(t *testing.T) {
	// Build a ZIP where required entries use wrong casing.
	// "Word/Document.xml" ≠ "word/document.xml"
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	if _, err := w.Create("Word/Document.xml"); err != nil {
		t.Fatalf("create Word/Document.xml: %v", err)
	}
	if _, err := w.Create("_Rels/.rels"); err != nil {
		t.Fatalf("create _Rels/.rels: %v", err)
	}
	if _, err := w.Create("[Content_Types].xml"); err != nil {
		t.Fatalf("create [Content_Types].xml: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}

	err := Validate(buf.Bytes(), FormatDOCX)
	if err == nil {
		t.Fatal("Validate(case mismatch): expected error, got nil")
	}
	if !errors.Is(err, ErrMissingEntry) {
		t.Errorf("expected ErrMissingEntry, got: %v", err)
	}
}

// --- 1.8.5 Resource cleanup ---

func TestIntegration_NoResourceLeak(t *testing.T) {
	// Validate operates entirely in-memory on []byte — no file descriptors
	// or external resources are allocated. This test verifies that repeatedly
	// calling Validate does not accumulate any observable resource usage.
	data := makeMinimalZip(makeFormatDOCX().RequiredEntries)
	for i := 0; i < 1000; i++ {
		if err := Validate(data, FormatDOCX); err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
	}
	// ValidateFromPath does open a file — ensure it's closed.
	dir := t.TempDir()
	path := dir + "/test.docx"
	writeFile(t, path, data)
	for i := 0; i < 100; i++ {
		if err := ValidateFromPath(path, FormatDOCX); err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

// --- 1.8.6 Concurrency test ---

func TestConcurrency_Validate(t *testing.T) {
	data := makeMinimalZip(makeFormatDOCX().RequiredEntries)

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Validate(data, FormatDOCX); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent Validate returned error: %v", err)
	}
}

func TestConcurrency_MultipleFormats(t *testing.T) {
	docxData := makeMinimalZip(makeFormatDOCX().RequiredEntries)
	xlsxData := makeMinimalZip(makeFormatXLSX().RequiredEntries)
	pptxData := makeMinimalZip(makeFormatPPTX().RequiredEntries)

	var wg sync.WaitGroup
	errs := make(chan error, 300)

	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			if err := Validate(docxData, FormatDOCX); err != nil {
				errs <- err
			}
		}()
		go func() {
			defer wg.Done()
			if err := Validate(xlsxData, FormatXLSX); err != nil {
				errs <- err
			}
		}()
		go func() {
			defer wg.Done()
			if err := Validate(pptxData, FormatPPTX); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent multi-format Validate: %v", err)
	}
}
