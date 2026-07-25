package magicseal

import (
	"errors"
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
		"word/document.xml": {},
		"_rels/.rels":       {},
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
