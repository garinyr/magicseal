package magicseal

import (
	"errors"
	"testing"
)

func TestErrorsIs_AllSentinels(t *testing.T) {
	// Verify that every sentinel error can be matched with errors.Is()
	// after being wrapped in newValidationError.
	tests := []struct {
		name     string
		sentinel error
		format   FormatName
	}{
		{"ErrUnsupportedFormat", ErrUnsupportedFormat, "unknown"},
		{"ErrMagicMismatch", ErrMagicMismatch, FormatDOCX},
		{"ErrInvalidZip", ErrInvalidZip, FormatDOCX},
		{"ErrMissingEntry", ErrMissingEntry, FormatDOCX},
		{"ErrSizeLimitExceeded", ErrSizeLimitExceeded, FormatDOCX},
		{"ErrCompressionRatioExceeded", ErrCompressionRatioExceeded, FormatDOCX},
		{"ErrTooManyEntries", ErrTooManyEntries, FormatDOCX},
		{"ErrPathTraversal", ErrPathTraversal, FormatDOCX},
		{"ErrFormatAlreadyRegistered", ErrFormatAlreadyRegistered, "dup"},
		{"ErrFormatNotRegistered", ErrFormatNotRegistered, "gone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := newValidationError(tt.sentinel, tt.format, "test details")
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("errors.Is(err, %s): expected true, got false. err=%v", tt.name, err)
			}
		})
	}
}

func TestErrorsAs_ValidationError(t *testing.T) {
	err := newValidationError(ErrMissingEntry, FormatDOCX, "missing: word/document.xml")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatal("errors.As(err, &ValidationError): expected true, got false")
	}
	if ve.Code != CodeMissingEntry {
		t.Errorf("ValidationError.Code: got %d, want %d", ve.Code, CodeMissingEntry)
	}
	if ve.Format != FormatDOCX {
		t.Errorf("ValidationError.Format: got %q, want %q", ve.Format, FormatDOCX)
	}
	if ve.Details != "missing: word/document.xml" {
		t.Errorf("ValidationError.Details: got %q, want %q", ve.Details, "missing: word/document.xml")
	}
	if ve.Err != ErrMissingEntry {
		t.Errorf("ValidationError.Err: got %v, want %v", ve.Err, ErrMissingEntry)
	}
}

func TestErrorsIs_ValidateMissingEntry(t *testing.T) {
	// Validate must return an error that passes errors.Is for the right sentinel.
	data := makeMinimalZip(t, []string{"a.txt"})
	err := Validate(data, FormatDOCX)
	if err == nil {
		t.Fatal("Validate(missing entries): expected error, got nil")
	}
	if !errors.Is(err, ErrMissingEntry) {
		t.Errorf("errors.Is(err, ErrMissingEntry): expected true, got: %v", err)
	}

	// Confirm it also works with errors.As for richer inspection.
	var ve *ValidationError
	if errors.As(err, &ve) {
		if ve.Code != CodeMissingEntry {
			t.Errorf("Code: got %d, want %d", ve.Code, CodeMissingEntry)
		}
		if ve.Format != FormatDOCX {
			t.Errorf("Format: got %q, want %q", ve.Format, FormatDOCX)
		}
	}
}

func TestErrorsIs_ValidateUnsupportedFormat(t *testing.T) {
	err := Validate([]byte("PK\x03\x04"), "ghost")
	if err == nil {
		t.Fatal("Validate(unsupported): expected error, got nil")
	}
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("errors.Is(err, ErrUnsupportedFormat): expected true, got: %v", err)
	}
	var ve *ValidationError
	if errors.As(err, &ve) {
		if ve.Code != CodeUnsupportedFormat {
			t.Errorf("Code: got %d, want %d", ve.Code, CodeUnsupportedFormat)
		}
	}
}

func TestErrorsIs_Registration(t *testing.T) {
	// Registration errors are NOT wrapped in ValidationError — they use sentinels directly.
	err := RegisterFormat(Format{
		Name:       "docx",
		Extensions: []string{".xyz"},
		MagicBytes: []byte("PK\x03\x04"),
	})
	if !errors.Is(err, ErrFormatAlreadyRegistered) {
		t.Errorf("RegisterFormat(dup): expected ErrFormatAlreadyRegistered, got: %v", err)
	}
}
