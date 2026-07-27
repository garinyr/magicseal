# MagicSeal

[![Go Reference](https://pkg.go.dev/badge/github.com/garinyr/magicseal.svg)](https://pkg.go.dev/github.com/garinyr/magicseal)
[![CI](https://github.com/garinyr/magicseal/actions/workflows/ci.yml/badge.svg)](https://github.com/garinyr/magicseal/actions/workflows/ci.yml)

Deep file format validator for ZIP-based container formats — verifies that a file
is what it claims to be, not just a renamed ZIP or a malicious payload.

**Why this exists:** a file with a `.docx` extension and `PK\x03\x04` magic bytes
could be anything — an empty ZIP, a zip bomb, or a random archive renamed to slip past
a naive content-type check. MagicSeal goes beyond the first four bytes: it validates
the internal ZIP structure and confirms that the mandatory entries for the claimed
format actually exist, all from metadata without ever decompressing file contents.

## Requirements

Go 1.22 or later.

## Installation

```sh
go get github.com/garinyr/magicseal
```

## Usage

### Basic validation

```go
import "github.com/garinyr/magicseal"

data, _ := os.ReadFile("document.docx")
if err := magicseal.Validate(data, magicseal.FormatDOCX); err != nil {
    // handle invalid file
}
```

```go
// Convenience helper — reads the file for you.
if err := magicseal.ValidateFromPath("spreadsheet.xlsx", magicseal.FormatXLSX); err != nil {
    // handle invalid file
}
```

### Streaming (large files)

For files >100MB, use `ValidateReader` to avoid loading everything into memory:

```go
f, _ := os.Open("large.xlsx")
defer f.Close()
fi, _ := f.Stat()
if err := magicseal.ValidateReader(f, fi.Size(), magicseal.FormatXLSX); err != nil {
    // handle invalid file
}
```

### Custom format registration

```go
magicseal.RegisterFormat(magicseal.Format{
    Name:            "xpi",
    Extensions:      []string{".xpi"},
    MagicBytes:      []byte("PK\x03\x04"),
    RequiredEntries: []string{"manifest.json"},
})

// List all registered formats
for _, f := range magicseal.ListFormats() {
    fmt.Println(f.Name, f.Extensions)
}

// Remove a format
magicseal.UnregisterFormat("xpi")
```

### Configurable limits

```go
err := magicseal.ValidateWithConfig(data, magicseal.FormatDOCX, &magicseal.ValidatorConfig{
    MaxDecompressedSize: 128 << 20, // 128 MB
    MaxCompressionRatio: 50.0,      // tighten ratio guard
})
```

### Structured error handling

```go
err := magicseal.Validate(data, magicseal.FormatDOCX)

// Sentinel errors — backward compatible (same as MVP 1)
if errors.Is(err, magicseal.ErrMissingEntry) {
    // handle missing entry
}

// Structured error — richer inspection
var ve *magicseal.ValidationError
if errors.As(err, &ve) {
    fmt.Printf("code=%d format=%s details=%s\n", ve.Code, ve.Format, ve.Details)
}
```

### CLI

```sh
# Install
go install github.com/garinyr/magicseal/cmd/magicseal@latest

# Check a file
magicseal check document.docx --as docx

# JSON output
magicseal check file.apk --as apk --json

# List registered formats
magicseal formats
```

Exit codes: `0` = valid, `1` = invalid, `2` = error.

## Supported Formats

| Format | Extension | Key Required Entries |
|--------|-----------|---------------------|
| DOCX   | `.docx`   | `word/document.xml`, `_rels/.rels`, `[Content_Types].xml` |
| XLSX   | `.xlsx`   | `xl/workbook.xml`, `_rels/.rels`, `[Content_Types].xml`, `xl/_rels/workbook.xml.rels` |
| PPTX   | `.pptx`   | `ppt/presentation.xml`, `_rels/.rels`, `[Content_Types].xml`, `ppt/_rels/presentation.xml.rels` |
| ODT    | `.odt`    | `content.xml`, `META-INF/manifest.xml`, `mimetype` |
| ODS    | `.ods`    | `content.xml`, `META-INF/manifest.xml`, `mimetype` |
| JAR    | `.jar`    | `META-INF/MANIFEST.MF` |
| APK    | `.apk`    | `AndroidManifest.xml`, `classes.dex` (unsigned/debug APKs supported) |
| XPI    | `.xpi`    | `manifest.json` (WebExtension, Firefox 57+) |

## Validation Layers

**Layer 1 — Magic bytes**: Fast-fail on non-ZIP files.

**Guard — Path traversal**: All entry names checked for `../`, `..\`, absolute paths, null byte injection. Platform-consistent via forward-slash semantics.

**Layer 2 — ZIP structure**: Central directory parsing via `archive/zip.NewReader`.

**Guard — Safety limits**: Entry count (default 10k), total decompressed size (default 256MB), compression ratio (default 100:1).

**Layer 3 — Required entries**: Exact case-sensitive match of mandatory entries.

All decisions from **central-directory metadata only** — file contents are never decompressed.

## Error Handling

Ten sentinel errors support `errors.Is()`:

| Sentinel | Meaning |
|----------|---------|
| `ErrMagicMismatch` | Not a ZIP file |
| `ErrInvalidZip` | Corrupted ZIP structure |
| `ErrMissingEntry` | Required entry absent |
| `ErrCompressionRatioExceeded` | Zip bomb (high ratio) |
| `ErrSizeLimitExceeded` | Total size exceeds limit |
| `ErrTooManyEntries` | Entry count exceeds limit |
| `ErrUnsupportedFormat` | Unknown format |
| `ErrPathTraversal` | Suspicious entry name |
| `ErrFormatAlreadyRegistered` | Duplicate format registration |
| `ErrFormatNotRegistered` | Unregistering unknown format |

`ValidationError` wraps sentinels with machine-readable `Code` (type `ErrorCode`), `Format`, and `Details` fields. `errors.Is()` continues to work — backward compatible.

## Dropped Formats

**CRX (Chrome Extension)**: `.crx` files use a custom binary header before ZIP data starts — not valid ZIP from byte 0. Not covered by MagicSeal's architecture. See `docs/task/magicseal-mvp2.md` for details.

## Design Properties

- **Thread-safe** — registry protected by `sync.RWMutex`, safe for concurrent `RegisterFormat`/`UnregisterFormat`/`Validate`.
- **Metadata-only** — entry contents are never decompressed.
- **Single read** — input read once, result reused across layers.
- **Mutable but safe** — formats can be registered at runtime (MVP 2).
- **Zero dependencies** — standard library only.

## License

MIT — see [LICENSE](./LICENSE)
