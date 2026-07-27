package magicseal

import (
	"bytes"
	"errors"
	"testing"
)

func TestCheckMagicValid(t *testing.T) {
	f := makeFormatDOCX()
	data := []byte("PK\x03\x04\x00\x00\x00\x00")
	if err := checkMagic(bytes.NewReader(data), int64(len(data)), f); err != nil {
		t.Errorf("checkMagic(valid ZIP header): unexpected error: %v", err)
	}
}

func TestCheckMagicEmptyData(t *testing.T) {
	f := makeFormatDOCX()
	data := []byte{}
	err := checkMagic(bytes.NewReader(data), int64(len(data)), f)
	if err == nil {
		t.Fatal("checkMagic(empty data): expected error, got nil")
	}
	if !errors.Is(err, ErrMagicMismatch) {
		t.Errorf("checkMagic(empty data): error does not wrap ErrMagicMismatch: %v", err)
	}
}

func TestCheckMagicTooShort(t *testing.T) {
	f := makeFormatDOCX()
	data := []byte("PK") // only 2 bytes, need 4
	err := checkMagic(bytes.NewReader(data), int64(len(data)), f)
	if err == nil {
		t.Fatal("checkMagic(short data): expected error, got nil")
	}
	if !errors.Is(err, ErrMagicMismatch) {
		t.Errorf("checkMagic(short data): error does not wrap ErrMagicMismatch: %v", err)
	}
}

func TestCheckMagicWrongBytes(t *testing.T) {
	f := makeFormatDOCX()
	data := []byte("MZ\x00\x00") // DOS/PE header, not ZIP
	err := checkMagic(bytes.NewReader(data), int64(len(data)), f)
	if err == nil {
		t.Fatal("checkMagic(wrong bytes): expected error, got nil")
	}
	if !errors.Is(err, ErrMagicMismatch) {
		t.Errorf("checkMagic(wrong bytes): error does not wrap ErrMagicMismatch: %v", err)
	}
}

func TestCheckMagicNonZip(t *testing.T) {
	// Plain text file renamed to .docx — Layer 1 should reject immediately.
	f := makeFormatDOCX()
	data := []byte("Hello, this is a text file disguised as docx")
	err := checkMagic(bytes.NewReader(data), int64(len(data)), f)
	if err == nil {
		t.Fatal("checkMagic(non-ZIP): expected error, got nil")
	}
	if !errors.Is(err, ErrMagicMismatch) {
		t.Errorf("checkMagic(non-ZIP): error does not wrap ErrMagicMismatch: %v", err)
	}
}

func TestCheckMagicNilData(t *testing.T) {
	f := makeFormatDOCX()
	err := checkMagic(bytes.NewReader(nil), 0, f)
	if err == nil {
		t.Fatal("checkMagic(nil): expected error, got nil")
	}
	if !errors.Is(err, ErrMagicMismatch) {
		t.Errorf("checkMagic(nil): error does not wrap ErrMagicMismatch: %v", err)
	}
}
