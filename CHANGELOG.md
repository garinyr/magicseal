# Changelog

All notable changes to this project will be documented in this file.

## [1.0.0] — 2026-07-27

First stable release. All MVP 1 APIs remain unchanged and backward compatible.

### Added

- **Format Registration API**: `RegisterFormat`, `UnregisterFormat`, `ListFormats` for runtime format management.
- **New built-in formats**: ODT, ODS, JAR, APK (incl. unsigned/debug), XPI (WebExtension).
- **Path traversal detection**: All ZIP entry names checked for `../`, `..\`, absolute paths, and null byte injection. Platform-consistent via forward-slash semantics.
- **Streaming support**: `ValidateReader(io.ReaderAt, int64, FormatName)` for large files (>100MB).
- **Configurable limits**: `ValidateWithConfig`, `ValidateReaderWithConfig`, `ValidatorConfig` struct with safe defaults.
- **Structured errors**: `ValidationError` type with `ErrorCode`, `Format`, and `Details`. `errors.Is()` backward compatible.
- **CLI**: `cmd/magicseal` with `check` (human + JSON) and `formats` commands.
- **New sentinel errors**: `ErrPathTraversal`, `ErrFormatAlreadyRegistered`, `ErrFormatNotRegistered`.
- **Thread-safe mutability**: Registry now uses `sync.RWMutex`, safe for concurrent register/unregister/validate.

### Dropped

- **CRX (Chrome Extension)**: Not ZIP from byte 0 — custom binary header requires pre-processing outside magicseal's architecture.

### Internal

- Refactored `parseZip`, `checkMagic`, `validate` to accept `io.ReaderAt` — enables streaming without loading entire file.
- `checkLimits` now accepts `*ValidatorConfig` for custom limits.

## [0.1.0] — MVP 1

Initial release with DOCX, XLSX, PPTX validation. Three-layer inspection: magic bytes → ZIP structure → required entries. Safety guards for entry count, decompressed size, and compression ratio.
