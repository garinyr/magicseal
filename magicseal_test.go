package magicseal

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDOCX(t *testing.T) {
	data := makeMinimalZip([]string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
	})
	if err := Validate(data, FormatDOCX); err != nil {
		t.Errorf("Validate(valid DOCX): unexpected error: %v", err)
	}
}

func TestValidateXLSX(t *testing.T) {
	data := makeMinimalZip([]string{
		"xl/workbook.xml",
		"_rels/.rels",
		"[Content_Types].xml",
		"xl/_rels/workbook.xml.rels",
	})
	if err := Validate(data, FormatXLSX); err != nil {
		t.Errorf("Validate(valid XLSX): unexpected error: %v", err)
	}
}

func TestValidatePPTX(t *testing.T) {
	data := makeMinimalZip([]string{
		"ppt/presentation.xml",
		"_rels/.rels",
		"[Content_Types].xml",
		"ppt/_rels/presentation.xml.rels",
	})
	if err := Validate(data, FormatPPTX); err != nil {
		t.Errorf("Validate(valid PPTX): unexpected error: %v", err)
	}
}

func TestValidateUnsupportedFormat(t *testing.T) {
	data := makeMinimalZip([]string{"x"})
	err := Validate(data, "unknown")
	if err == nil {
		t.Fatal("Validate(unknown format): expected error, got nil")
	}
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("Validate(unknown format): error does not wrap ErrUnsupportedFormat: %v", err)
	}
}

func TestValidateWrongFormat(t *testing.T) {
	// Data is a valid DOCX, but claimed as XLSX — should fail on missing XLSX entries.
	data := makeMinimalZip([]string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
	})
	err := Validate(data, FormatXLSX)
	if err == nil {
		t.Fatal("Validate(wrong format): expected error (missing XLSX entries), got nil")
	}
	if !errors.Is(err, ErrMissingEntry) {
		t.Errorf("Validate(wrong format): expected ErrMissingEntry, got: %v", err)
	}
}

func TestValidateRenamedZip(t *testing.T) {
	// A plain ZIP (no Office entries) claimed as DOCX — must reject.
	data := makeMinimalZip([]string{"readme.txt", "images/photo.png"})
	err := Validate(data, FormatDOCX)
	if err == nil {
		t.Fatal("Validate(renamed ZIP): expected error, got nil")
	}
	if !errors.Is(err, ErrMissingEntry) {
		t.Errorf("Validate(renamed ZIP): expected ErrMissingEntry, got: %v", err)
	}
}

func TestValidateNonZip(t *testing.T) {
	data := []byte("Hello, I am a text file")
	err := Validate(data, FormatDOCX)
	if err == nil {
		t.Fatal("Validate(non-ZIP): expected error, got nil")
	}
	if !errors.Is(err, ErrMagicMismatch) {
		t.Errorf("Validate(non-ZIP): expected ErrMagicMismatch, got: %v", err)
	}
}

func TestValidateEmptyData(t *testing.T) {
	err := Validate([]byte{}, FormatDOCX)
	if err == nil {
		t.Fatal("Validate(empty): expected error, got nil")
	}
	if !errors.Is(err, ErrMagicMismatch) {
		t.Errorf("Validate(empty): expected ErrMagicMismatch, got: %v", err)
	}
}

func TestValidateFromPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.docx")
	data := makeMinimalZip([]string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
	})
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := ValidateFromPath(path, FormatDOCX); err != nil {
		t.Errorf("ValidateFromPath(valid): unexpected error: %v", err)
	}
}

func TestValidateFromPathNotFound(t *testing.T) {
	err := ValidateFromPath("/nonexistent/file.docx", FormatDOCX)
	if err == nil {
		t.Fatal("ValidateFromPath(missing): expected error, got nil")
	}
}
