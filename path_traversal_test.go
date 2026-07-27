package magicseal

import (
	"errors"
	"strings"
	"testing"
)

func TestCheckPathTraversalClean(t *testing.T) {
	entries := []zipEntry{
		{Name: "word/document.xml"},
		{Name: "_rels/.rels"},
		{Name: "[Content_Types].xml"},
		{Name: "images/photo.png"},
	}
	if err := checkPathTraversal(entries); err != nil {
		t.Errorf("checkPathTraversal(clean): unexpected error: %v", err)
	}
}

func TestCheckPathTraversalNoEntries(t *testing.T) {
	if err := checkPathTraversal(nil); err != nil {
		t.Errorf("checkPathTraversal(nil): unexpected error: %v", err)
	}
	if err := checkPathTraversal([]zipEntry{}); err != nil {
		t.Errorf("checkPathTraversal(empty): unexpected error: %v", err)
	}
}

func TestCheckPathTraversalForwardSlash(t *testing.T) {
	entries := []zipEntry{
		{Name: "a/b/c"},
		{Name: "../evil.exe"},
	}
	err := checkPathTraversal(entries)
	if err == nil {
		t.Fatal("checkPathTraversal(../): expected error, got nil")
	}
	if !errors.Is(err, ErrPathTraversal) {
		t.Errorf("checkPathTraversal(../): error does not wrap ErrPathTraversal: %v", err)
	}
}

func TestCheckPathTraversalBackslash(t *testing.T) {
	entries := []zipEntry{
		{Name: `..\windows\system32\evil.dll`},
	}
	err := checkPathTraversal(entries)
	if err == nil {
		t.Fatal("checkPathTraversal(..\\): expected error, got nil")
	}
	if !errors.Is(err, ErrPathTraversal) {
		t.Errorf("checkPathTraversal(..\\): error does not wrap ErrPathTraversal: %v", err)
	}
}

func TestCheckPathTraversalAbsolutePath(t *testing.T) {
	entries := []zipEntry{
		{Name: "/etc/passwd"},
	}
	err := checkPathTraversal(entries)
	if err == nil {
		t.Fatal("checkPathTraversal(absolute): expected error, got nil")
	}
	if !errors.Is(err, ErrPathTraversal) {
		t.Errorf("checkPathTraversal(absolute): error does not wrap ErrPathTraversal: %v", err)
	}
}

func TestCheckPathTraversalDeeplyNested(t *testing.T) {
	// Deep relative escape: a/../../etc/passwd
	entries := []zipEntry{
		{Name: "a/b/../../../etc/passwd"},
	}
	err := checkPathTraversal(entries)
	if err == nil {
		t.Fatal("checkPathTraversal(deep ../): expected error, got nil")
	}
}

func TestCheckPathTraversalNullByte(t *testing.T) {
	entries := []zipEntry{
		{Name: "safe.txt\x00.exe"},
	}
	err := checkPathTraversal(entries)
	if err == nil {
		t.Fatal("checkPathTraversal(null byte): expected error, got nil")
	}
	if !errors.Is(err, ErrPathTraversal) {
		t.Errorf("checkPathTraversal(null byte): error does not wrap ErrPathTraversal: %v", err)
	}
}

func TestCheckPathTraversalDoubleDotEnd(t *testing.T) {
	// Entry name ".." should be caught by normalization.
	entries := []zipEntry{
		{Name: ".."},
	}
	err := checkPathTraversal(entries)
	if err == nil {
		t.Fatal("checkPathTraversal('..'): expected error, got nil")
	}
	if !errors.Is(err, ErrPathTraversal) {
		t.Errorf("checkPathTraversal('..'): error does not wrap ErrPathTraversal: %v", err)
	}
}

func TestCheckPathTraversalPrefixDoubleDot(t *testing.T) {
	// Name that starts with ".." but not as a path segment, e.g. "..something"
	// This is NOT traversal — ".." must be a complete path segment.
	entries := []zipEntry{
		{Name: "..something.txt"},
	}
	if err := checkPathTraversal(entries); err != nil {
		t.Errorf("checkPathTraversal('..something'): should be OK, got: %v", err)
	}
}

func TestCheckPathTraversalMultipleSuspicious(t *testing.T) {
	// Multiple entries, one suspicious — should catch it.
	entries := []zipEntry{
		{Name: "safe.txt"},
		{Name: "data/config.xml"},
		{Name: "../malware.sh"},
	}
	err := checkPathTraversal(entries)
	if err == nil {
		t.Fatal("checkPathTraversal(mixed): expected error, got nil")
	}
	if !errors.Is(err, ErrPathTraversal) {
		t.Errorf("checkPathTraversal(mixed): error does not wrap ErrPathTraversal: %v", err)
	}
	if !strings.Contains(err.Error(), "../") {
		t.Errorf("checkPathTraversal(mixed): error message should mention the suspicious pattern: %v", err)
	}
}

func TestValidateRejectsPathTraversal(t *testing.T) {
	// Full Validate flow should reject files with traversal entries.
	data := makeMinimalZip(t, []string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
		"../evil.bin",
	})
	err := Validate(data, FormatDOCX)
	if err == nil {
		t.Fatal("Validate(../ entry): expected error, got nil")
	}
	if !errors.Is(err, ErrPathTraversal) {
		t.Errorf("Validate(../ entry): error does not wrap ErrPathTraversal: %v", err)
	}
}
