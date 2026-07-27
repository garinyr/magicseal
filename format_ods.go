package magicseal

// makeFormatODS returns the format definition for ODF Spreadsheet (.ods).
// Same structure as ODT: content.xml, META-INF/manifest.xml, mimetype.
func makeFormatODS() Format {
	return Format{
		Name:       FormatODS,
		Extensions: []string{".ods"},
		MIMEType:   "application/vnd.oasis.opendocument.spreadsheet",
		MagicBytes: []byte("PK\x03\x04"),
		RequiredEntries: []string{
			"content.xml",
			"META-INF/manifest.xml",
			"mimetype",
		},
	}
}
