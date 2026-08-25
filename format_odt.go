package magicseal

// makeFormatODT returns the format definition for ODF Text Document (.odt).
// ODF files use ZIP container with a well-defined entry layout.
// Note: mimetype entry is always STORED (uncompressed) and placed first —
// it automatically passes the compression ratio guard (ratio ≈ 1:1).
func makeFormatODT() Format {
	return Format{
		Name:       FormatODT,
		Extensions: []string{".odt"},
		MIMEType:   "application/vnd.oasis.opendocument.text",
		MagicBytes: []byte("PK\x03\x04"),
		RequiredEntries: []string{
			"content.xml",
			"META-INF/manifest.xml",
			"mimetype",
		},
	}
}
