# MagicSeal

Deep file format validator for ZIP-based container formats (DOCX, XLSX, PPTX).

[![Go Reference](https://pkg.go.dev/badge/github.com/garinyr/magicseal.svg)](https://pkg.go.dev/github.com/garinyr/magicseal)

## Usage

```go
import "github.com/garinyr/magicseal"

data, _ := os.ReadFile("document.docx")
if err := magicseal.Validate(data, magicseal.FormatDOCX); err != nil {
    // handle invalid file
}
```

## Supported Formats

| Format | Extension | MIME |
|--------|-----------|------|
| DOCX   | `.docx`   | `application/vnd.openxmlformats-officedocument.wordprocessingml.document` |
| XLSX   | `.xlsx`   | `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` |
| PPTX   | `.pptx`   | `application/vnd.openxmlformats-officedocument.presentationml.presentation` |

## How It Works

3-layer validation with safety guards:

1. **Magic bytes** — verify `PK\x03\x04` header (fast-fail for non-ZIP files)
2. **ZIP structure** — parse central directory via `archive/zip` (EOCD-based, not just local headers)
3. **Safety limits** — entry count (10k), decompressed size (256MB), compression ratio (100:1)
4. **Required entries** — ensure mandatory files exist for the claimed format (case-sensitive exact match)

## Error Handling

All errors implement `errors.Is()` support:

```go
err := magicseal.Validate(data, magicseal.FormatDOCX)
if errors.Is(err, magicseal.ErrMagicMismatch)            { /* not a ZIP file */ }
if errors.Is(err, magicseal.ErrInvalidZip)               { /* corrupted ZIP */  }
if errors.Is(err, magicseal.ErrMissingEntry)             { /* missing file */   }
if errors.Is(err, magicseal.ErrCompressionRatioExceeded) { /* zip bomb */       }
if errors.Is(err, magicseal.ErrSizeLimitExceeded)        { /* too large */      }
if errors.Is(err, magicseal.ErrTooManyEntries)           { /* too many entries */ }
```

## MVP 1 Limitations

- Only DOCX, XLSX, PPTX (ODF, JAR, APK → MVP 2)
- No custom format registration (→ MVP 2)
- No streaming/`io.ReaderAt` support (→ MVP 2)
- No structured error codes (→ MVP 2)
- No path traversal detection (→ MVP 2)
- No configurable limits (→ MVP 2)

## License

MIT — see [LICENSE](./LICENSE)
