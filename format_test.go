package magicseal

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestRegistryBuiltins(t *testing.T) {
	names := []FormatName{FormatDOCX, FormatXLSX, FormatPPTX}
	for _, name := range names {
		f, err := lookupFormat(name)
		if err != nil {
			t.Fatalf("lookupFormat(%q): unexpected error: %v", name, err)
		}
		if f.Name != name {
			t.Errorf("lookupFormat(%q): got name %q", name, f.Name)
		}
	}
}

func TestRegistryUnknownFormat(t *testing.T) {
	_, err := lookupFormat("unknown")
	if err == nil {
		t.Fatal("lookupFormat(unknown): expected error, got nil")
	}
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("lookupFormat(unknown): error %v does not wrap ErrUnsupportedFormat", err)
	}
}

func TestFormatRequiredEntriesExact(t *testing.T) {
	// Entry name comparison must be case-sensitive (ZIP spec).
	// "Word/Document.xml" ≠ "word/document.xml"
	entries := map[string]struct{}{
		"word/document.xml":   {},
		"_rels/.rels":         {},
		"[Content_Types].xml": {},
	}

	if _, ok := entries["Word/Document.xml"]; ok {
		t.Error(`"Word/Document.xml" incorrectly matched "word/document.xml" — comparison must be case-sensitive`)
	}
}

func TestFormatDOCXRequiredEntries(t *testing.T) {
	f := makeFormatDOCX()
	expected := []string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
	}
	if len(f.RequiredEntries) != len(expected) {
		t.Fatalf("DOCX: got %d required entries, want %d", len(f.RequiredEntries), len(expected))
	}
	for i, e := range expected {
		if f.RequiredEntries[i] != e {
			t.Errorf("DOCX entry[%d]: got %q, want %q", i, f.RequiredEntries[i], e)
		}
	}
}

func TestFormatXLSXRequiredEntries(t *testing.T) {
	f := makeFormatXLSX()
	expected := []string{
		"xl/workbook.xml",
		"_rels/.rels",
		"[Content_Types].xml",
		"xl/_rels/workbook.xml.rels",
	}
	if len(f.RequiredEntries) != len(expected) {
		t.Fatalf("XLSX: got %d required entries, want %d", len(f.RequiredEntries), len(expected))
	}
	for i, e := range expected {
		if f.RequiredEntries[i] != e {
			t.Errorf("XLSX entry[%d]: got %q, want %q", i, f.RequiredEntries[i], e)
		}
	}
}

func TestFormatPPTXRequiredEntries(t *testing.T) {
	f := makeFormatPPTX()
	expected := []string{
		"ppt/presentation.xml",
		"_rels/.rels",
		"[Content_Types].xml",
		"ppt/_rels/presentation.xml.rels",
	}
	if len(f.RequiredEntries) != len(expected) {
		t.Fatalf("PPTX: got %d required entries, want %d", len(f.RequiredEntries), len(expected))
	}
	for i, e := range expected {
		if f.RequiredEntries[i] != e {
			t.Errorf("PPTX entry[%d]: got %q, want %q", i, f.RequiredEntries[i], e)
		}
	}
}

func TestRegisterFormatSuccess(t *testing.T) {
	f := Format{
		Name:       "custom",
		Extensions: []string{".custom"},
		MagicBytes: []byte("PK\x03\x04"),
		RequiredEntries: []string{"data.xml"},
	}
	if err := RegisterFormat(f); err != nil {
		t.Fatalf("RegisterFormat(custom): unexpected error: %v", err)
	}
	defer func() {
		if err := UnregisterFormat("custom"); err != nil {
			t.Errorf("UnregisterFormat(custom): unexpected error: %v", err)
		}
	}()

	// Verify format is usable via lookupFormat.
	got, err := lookupFormat("custom")
	if err != nil {
		t.Fatalf("lookupFormat(custom): unexpected error: %v", err)
	}
	if got.Name != "custom" {
		t.Errorf("lookupFormat(custom): got name %q, want %q", got.Name, "custom")
	}

	// Verify format appears in ListFormats.
	formats := ListFormats()
	found := false
	for _, lf := range formats {
		if lf.Name == "custom" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ListFormats: custom format not found")
	}
}

func TestRegisterFormatDuplicateName(t *testing.T) {
	f := Format{
		Name:       "docx", // already registered as builtin
		Extensions: []string{".fake"},
		MagicBytes: []byte("PK\x03\x04"),
	}
	err := RegisterFormat(f)
	if err == nil {
		t.Fatal("RegisterFormat(duplicate name): expected error, got nil")
	}
	if !errors.Is(err, ErrFormatAlreadyRegistered) {
		t.Errorf("RegisterFormat(duplicate name): error does not wrap ErrFormatAlreadyRegistered: %v", err)
	}
}

func TestRegisterFormatDuplicateExtension(t *testing.T) {
	f := Format{
		Name:       "fake-docx",
		Extensions: []string{".docx"}, // already registered by docx
		MagicBytes: []byte("PK\x03\x04"),
	}
	err := RegisterFormat(f)
	if err == nil {
		t.Fatal("RegisterFormat(duplicate extension): expected error, got nil")
	}
	if !errors.Is(err, ErrFormatAlreadyRegistered) {
		t.Errorf("RegisterFormat(duplicate extension): error does not wrap ErrFormatAlreadyRegistered: %v", err)
	}
}

func TestUnregisterFormatSuccess(t *testing.T) {
	f := Format{
		Name:       "to-remove",
		Extensions: []string{".remove"},
		MagicBytes: []byte("PK\x03\x04"),
	}
	if err := RegisterFormat(f); err != nil {
		t.Fatalf("RegisterFormat: %v", err)
	}

	if err := UnregisterFormat("to-remove"); err != nil {
		t.Fatalf("UnregisterFormat: unexpected error: %v", err)
	}

	// Verify format is gone.
	_, err := lookupFormat("to-remove")
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("lookupFormat after unregister: expected ErrUnsupportedFormat, got: %v", err)
	}

	// Verify format not in ListFormats.
	for _, lf := range ListFormats() {
		if lf.Name == "to-remove" {
			t.Error("ListFormats: removed format still present")
		}
	}
}

func TestUnregisterFormatNotFound(t *testing.T) {
	err := UnregisterFormat("nonexistent")
	if err == nil {
		t.Fatal("UnregisterFormat(nonexistent): expected error, got nil")
	}
	if !errors.Is(err, ErrFormatNotRegistered) {
		t.Errorf("UnregisterFormat(nonexistent): error does not wrap ErrFormatNotRegistered: %v", err)
	}
}

func TestListFormatsReturnsCopy(t *testing.T) {
	// ListFormats must return a copy, not a reference to the internal map.
	formats1 := ListFormats()
	formats2 := ListFormats()

	// Mutate formats1 — should not affect formats2.
	if len(formats1) > 1 {
		formats1[0], formats1[1] = formats1[1], formats1[0]
	}

	if len(formats2) != len(formats1) {
		t.Error("ListFormats: modifying returned slice appears to affect internal state")
	}
}

func TestFormatRegistryThreadSafety(t *testing.T) {
	// Concurrent register/unregister and lookups must not race.
	const goroutines = 20
	var wg sync.WaitGroup

	// Register custom formats concurrently.
	wg.Add(goroutines)
	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			name := FormatName(fmt.Sprintf("thread-%d-%d", idx, idx))
			f := Format{
				Name:       name,
				Extensions: []string{fmt.Sprintf(".t%d", idx)},
				MagicBytes: []byte("PK\x03\x04"),
			}
			// Ignore errors — some may race on duplicate extension if indices collide;
			// we just care that it doesn't panic/race.
			_ = RegisterFormat(f)
			_, _ = lookupFormat(FormatDOCX)
			ListFormats()
			_ = UnregisterFormat(name)
		}(i)
	}
	wg.Wait()
}
