package magicseal

import (
	"errors"
	"testing"
)

func TestCheckEntriesAllPresent(t *testing.T) {
	entries := []zipEntry{
		{Name: "word/document.xml"},
		{Name: "_rels/.rels"},
		{Name: "[Content_Types].xml"},
		{Name: "docProps/app.xml"}, // extra, not required
	}
	required := []string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
	}
	if err := checkEntries(entries, required); err != nil {
		t.Errorf("checkEntries(all present): unexpected error: %v", err)
	}
}

func TestCheckEntriesMissingEntry(t *testing.T) {
	entries := []zipEntry{
		{Name: "word/document.xml"},
		// "_rels/.rels" missing
		{Name: "[Content_Types].xml"},
	}
	required := []string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
	}
	err := checkEntries(entries, required)
	if err == nil {
		t.Fatal("checkEntries(missing): expected error, got nil")
	}
	if !errors.Is(err, ErrMissingEntry) {
		t.Errorf("checkEntries(missing): error does not wrap ErrMissingEntry: %v", err)
	}
}

func TestCheckEntriesCaseSensitive(t *testing.T) {
	// ZIP spec: entry names are case-sensitive.
	// "Word/Document.xml" ≠ "word/document.xml"
	entries := []zipEntry{
		{Name: "Word/Document.xml"},
		{Name: "_rels/.rels"},
		{Name: "[Content_Types].xml"},
	}
	required := []string{
		"word/document.xml", // lowercase — should NOT match "Word/Document.xml"
	}
	err := checkEntries(entries, required)
	if err == nil {
		t.Fatal("checkEntries(case mismatch): expected error, got nil — comparison must be case-sensitive")
	}
	if !errors.Is(err, ErrMissingEntry) {
		t.Errorf("checkEntries(case mismatch): error does not wrap ErrMissingEntry: %v", err)
	}
}

func TestCheckEntriesEmptyRequired(t *testing.T) {
	entries := []zipEntry{{Name: "a"}}
	if err := checkEntries(entries, nil); err != nil {
		t.Errorf("checkEntries(nil required): unexpected error: %v", err)
	}
	if err := checkEntries(entries, []string{}); err != nil {
		t.Errorf("checkEntries(empty required): unexpected error: %v", err)
	}
}

func TestCheckEntriesNoEntries(t *testing.T) {
	err := checkEntries(nil, []string{"x"})
	if err == nil {
		t.Fatal("checkEntries(no entries, has required): expected error, got nil")
	}
	if !errors.Is(err, ErrMissingEntry) {
		t.Errorf("checkEntries(no entries): error does not wrap ErrMissingEntry: %v", err)
	}
}
