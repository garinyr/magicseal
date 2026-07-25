// Package magicseal validates ZIP-based container formats (DOCX, XLSX, PPTX)
// using layered inspection: magic bytes → ZIP structure → required entries.
//
// Design principles:
//   - Format definitions are data, not control flow
//   - File is read once, parsed result reused across all layers
//   - Single public API entry point
//   - Metadata-only inspection — entry contents are never opened
//   - Thread-safe by default
package magicseal

import "fmt"

// FormatName is a typed constant identifying a supported format.
type FormatName string

const (
	FormatDOCX FormatName = "docx"
	FormatXLSX FormatName = "xlsx"
	FormatPPTX FormatName = "pptx"
)

var builtins = []Format{
	makeFormatDOCX(),
	makeFormatXLSX(),
	makeFormatPPTX(),
}

// formatRegistry holds all registered formats. Built once at init(), read-only afterward.
// This is safe for concurrent reads without a lock.
var formatRegistry = make(map[FormatName]Format)

func init() {
	for _, f := range builtins {
		formatRegistry[f.Name] = f
	}
}

// Format defines the structure of a supported container format.
type Format struct {
	Name            FormatName
	Extensions      []string
	MIMEType        string
	MagicBytes      []byte
	RequiredEntries []string
}

// lookupFormat returns the registered Format for the given name, or an error.
func lookupFormat(name FormatName) (Format, error) {
	f, ok := formatRegistry[name]
	if !ok {
		return Format{}, fmt.Errorf("%w: %q", ErrUnsupportedFormat, name)
	}
	return f, nil
}

// RegisterFormat adds a custom format to the registry.
// This is reserved for MVP 2.
func RegisterFormat(f Format) error {
	return fmt.Errorf("format registration not available in MVP 1")
}

// ListFormats returns all registered format names.
// This is reserved for MVP 2.
func ListFormats() []FormatName { return nil }
