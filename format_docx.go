package magicseal

// makeFormatDOCX returns the format definition for Office Open XML Document (.docx).
func makeFormatDOCX() Format {
	return Format{
		Name:       FormatDOCX,
		Extensions: []string{".docx"},
		MIMEType:   "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		MagicBytes: []byte("PK\x03\x04"),
		RequiredEntries: []string{
			"word/document.xml",
			"_rels/.rels",
			"[Content_Types].xml",
		},
	}
}
