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

## Supported Formats

| Format | Extension | MIME |
|--------|-----------|------|
| DOCX   | `.docx`   | `application/vnd.openxmlformats-officedocument.wordprocessingml.document` |
| XLSX   | `.xlsx`   | `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` |
| PPTX   | `.pptx`   | `application/vnd.openxmlformats-officedocument.presentationml.presentation` |

## Validation Layers

Three validation layers with a built-in safety guard between layers 2 and 3:

**Layer 1 — Magic bytes**  
Fast-fail check on `PK\x03\x04` (ZIP header). Non-ZIP files are rejected immediately
without further processing.

**Layer 2 — ZIP structure**  
Parses the ZIP central directory (EOCD-based via `archive/zip.NewReader`), not just
the local file headers. Corrupt, truncated, or garbage-prepended files are detected here.

**Guard — Safety limits** *(runs before layer 3)*  
Defense-in-depth checks to prevent resource exhaustion:
- Maximum entry count (10,000)
- Maximum total decompressed size (256 MB)
- Maximum compression ratio (100:1) — catches zip bombs

**Layer 3 — Required entries**  
Confirms that every mandatory entry for the claimed format exists in the archive
(case-sensitive exact match). A ZIP renamed to `.docx` will fail here because it
lacks `word/document.xml` and the other required Office Open XML entries.

All decisions are made from **central-directory metadata only** — file contents are
never opened or decompressed. The package is safe to use on untrusted input.

## Error Handling

All seven sentinel errors support `errors.Is()` for programmatic inspection:

```go
err := magicseal.Validate(data, magicseal.FormatDOCX)
if errors.Is(err, magicseal.ErrMagicMismatch)            { /* not a ZIP file */ }
if errors.Is(err, magicseal.ErrInvalidZip)               { /* corrupted ZIP */  }
if errors.Is(err, magicseal.ErrMissingEntry)             { /* missing entry */  }
if errors.Is(err, magicseal.ErrCompressionRatioExceeded) { /* zip bomb */       }
if errors.Is(err, magicseal.ErrSizeLimitExceeded)        { /* too large */      }
if errors.Is(err, magicseal.ErrTooManyEntries)           { /* too many */       }
if errors.Is(err, magicseal.ErrUnsupportedFormat)        { /* unknown format */ }
```

## Design Properties

- **Thread-safe** — the format registry is initialized once at startup and is read-only
  afterwards. `Validate` can be called concurrently from any number of goroutines without
  locks.
- **Metadata-only** — entry contents are never decompressed. The package reads the ZIP
  central directory only.
- **Single read** — input data is read once and the parsed result is reused across all
  validation layers.
- **Zero dependencies** — only the Go standard library.

## Current Limitations

| Capability | Status |
|-----------|--------|
| Custom format registration | Coming in a future release |
| Streaming / `io.ReaderAt` support | Coming in a future release |
| Structured error codes (`ValidationError`) | Coming in a future release |
| Path traversal detection in entry names | Coming in a future release |
| Configurable limits | Coming in a future release |
| Additional built-in formats (ODF, JAR, APK) | Coming in a future release |
| CLI wrapper | Coming in a future release |

## License

MIT — see [LICENSE](./LICENSE)
