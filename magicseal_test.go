package magicseal

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDOCX(t *testing.T) {
	data := makeMinimalZip(t, []string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
	})
	if err := Validate(data, FormatDOCX); err != nil {
		t.Errorf("Validate(valid DOCX): unexpected error: %v", err)
	}
}

func TestValidateXLSX(t *testing.T) {
	data := makeMinimalZip(t, []string{
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
	data := makeMinimalZip(t, []string{
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
	data := makeMinimalZip(t, []string{"x"})
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
	data := makeMinimalZip(t, []string{
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
	data := makeMinimalZip(t, []string{"readme.txt", "images/photo.png"})
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
	data := makeMinimalZip(t, []string{
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

func TestValidateReader(t *testing.T) {
	data := makeMinimalZip(t, []string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
	})
	r := bytes.NewReader(data)
	if err := ValidateReader(r, int64(len(data)), FormatDOCX); err != nil {
		t.Errorf("ValidateReader(valid DOCX): unexpected error: %v", err)
	}
}

func TestValidateReaderUnsupportedFormat(t *testing.T) {
	data := makeMinimalZip(t, []string{"x"})
	r := bytes.NewReader(data)
	err := ValidateReader(r, int64(len(data)), "unknown")
	if err == nil {
		t.Fatal("ValidateReader(unknown format): expected error, got nil")
	}
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("ValidateReader(unknown format): error does not wrap ErrUnsupportedFormat: %v", err)
	}
}

func TestValidateReaderWrongFormat(t *testing.T) {
	data := makeMinimalZip(t, []string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
	})
	r := bytes.NewReader(data)
	err := ValidateReader(r, int64(len(data)), FormatXLSX)
	if err == nil {
		t.Fatal("ValidateReader(wrong format): expected error, got nil")
	}
	if !errors.Is(err, ErrMissingEntry) {
		t.Errorf("ValidateReader(wrong format): expected ErrMissingEntry, got: %v", err)
	}
}

func TestValidateODT(t *testing.T) {
	data := makeMinimalZip(t, []string{
		"content.xml",
		"META-INF/manifest.xml",
		"mimetype",
	})
	if err := Validate(data, FormatODT); err != nil {
		t.Errorf("Validate(valid ODT): unexpected error: %v", err)
	}
}

func TestValidateODS(t *testing.T) {
	data := makeMinimalZip(t, []string{
		"content.xml",
		"META-INF/manifest.xml",
		"mimetype",
	})
	if err := Validate(data, FormatODS); err != nil {
		t.Errorf("Validate(valid ODS): unexpected error: %v", err)
	}
}

func TestValidateJAR(t *testing.T) {
	data := makeMinimalZip(t, []string{
		"META-INF/MANIFEST.MF",
	})
	if err := Validate(data, FormatJAR); err != nil {
		t.Errorf("Validate(valid JAR): unexpected error: %v", err)
	}
}

func TestValidateAPK(t *testing.T) {
	data := makeMinimalZip(t, []string{
		"AndroidManifest.xml",
		"classes.dex",
	})
	if err := Validate(data, FormatAPK); err != nil {
		t.Errorf("Validate(valid APK): unexpected error: %v", err)
	}
}

func TestValidateXPI(t *testing.T) {
	data := makeMinimalZip(t, []string{
		"manifest.json",
	})
	if err := Validate(data, FormatXPI); err != nil {
		t.Errorf("Validate(valid XPI): unexpected error: %v", err)
	}
}

func TestValidateODTMissingMimetype(t *testing.T) {
	data := makeMinimalZip(t, []string{
		"content.xml",
		"META-INF/manifest.xml",
		// "mimetype" missing
	})
	err := Validate(data, FormatODT)
	if err == nil {
		t.Fatal("Validate(ODT missing mimetype): expected error, got nil")
	}
	if !errors.Is(err, ErrMissingEntry) {
		t.Errorf("Validate(ODT missing mimetype): expected ErrMissingEntry, got: %v", err)
	}
}

func TestValidateAPKUnsigned(t *testing.T) {
	// Unsigned/debug APK — no META-INF/ directory. Must still pass.
	data := makeMinimalZip(t, []string{
		"AndroidManifest.xml",
		"classes.dex",
	})
	if err := Validate(data, FormatAPK); err != nil {
		t.Errorf("Validate(unsigned APK): unexpected error: %v", err)
	}
}

func TestValidateAPKMultiDex(t *testing.T) {
	// Multi-dex APK with multiple classes.dex files.
	// classes.dex is the primary one and must exist.
	data := makeMinimalZip(t, []string{
		"AndroidManifest.xml",
		"classes.dex",
		"classes2.dex",
		"classes3.dex",
	})
	if err := Validate(data, FormatAPK); err != nil {
		t.Errorf("Validate(multi-dex APK): unexpected error: %v", err)
	}
}

