// Package magicseal validates ZIP-based container formats (DOCX, XLSX, PPTX)
// using layered inspection: magic bytes → ZIP structure → required entries.
//
// Design principles:
//   - Format definitions are data, not control flow
//   - File is read once, parsed result reused across all layers
//   - Single public API entry point
//   - Metadata-only inspection — entry contents are never opened
//   - Thread-safe by default (safe for concurrent read and write via sync.RWMutex)
package magicseal

import (
	"fmt"
	"sync"
)

// FormatName is a typed constant identifying a supported format.
type FormatName string

const (
	FormatDOCX FormatName = "docx"
	FormatXLSX FormatName = "xlsx"
	FormatPPTX FormatName = "pptx"
	FormatODT  FormatName = "odt"
	FormatODS  FormatName = "ods"
	FormatJAR  FormatName = "jar"
	FormatAPK  FormatName = "apk"
	FormatXPI  FormatName = "xpi"
)

var builtins = []Format{
	makeFormatDOCX(),
	makeFormatXLSX(),
	makeFormatPPTX(),
	makeFormatODT(),
	makeFormatODS(),
	makeFormatJAR(),
	makeFormatAPK(),
	makeFormatXPI(),
}

// formatRegistry holds all registered formats.
// Protected by registryMu — safe for concurrent read and write.
var (
	formatRegistry = make(map[FormatName]Format)
	registryMu     sync.RWMutex
)

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
// Thread-safe: acquires a read lock.
func lookupFormat(name FormatName) (Format, error) {
	registryMu.RLock()
	f, ok := formatRegistry[name]
	registryMu.RUnlock()
	if !ok {
		return Format{}, fmt.Errorf("%w: %q", ErrUnsupportedFormat, name)
	}
	return f, nil
}

// RegisterFormat adds a custom format to the registry.
// Returns ErrFormatAlreadyRegistered if the format name or any extension
// is already registered.
// Thread-safe.
func RegisterFormat(f Format) error {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, ok := formatRegistry[f.Name]; ok {
		return fmt.Errorf("%w: %q", ErrFormatAlreadyRegistered, f.Name)
	}
	for _, ext := range f.Extensions {
		for _, existing := range formatRegistry {
			for _, existingExt := range existing.Extensions {
				if ext == existingExt {
					return fmt.Errorf("%w: extension %q already registered by format %q",
						ErrFormatAlreadyRegistered, ext, existing.Name)
				}
			}
		}
	}
	formatRegistry[f.Name] = f
	return nil
}

// UnregisterFormat removes a format from the registry by name.
// Returns ErrFormatNotRegistered if the name is not found.
// Thread-safe.
func UnregisterFormat(name FormatName) error {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, ok := formatRegistry[name]; !ok {
		return fmt.Errorf("%w: %q", ErrFormatNotRegistered, name)
	}
	delete(formatRegistry, name)
	return nil
}

// ListFormats returns a copy of all registered formats.
// Thread-safe: acquires a read lock and returns a snapshot.
func ListFormats() []Format {
	registryMu.RLock()
	defer registryMu.RUnlock()

	result := make([]Format, 0, len(formatRegistry))
	for _, f := range formatRegistry {
		result = append(result, f)
	}
	return result
}
