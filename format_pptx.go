package magicseal

// makeFormatPPTX returns the format definition for Office Open XML Presentation (.pptx).
func makeFormatPPTX() Format {
	return Format{
		Name:       FormatPPTX,
		Extensions: []string{".pptx"},
		MIMEType:   "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		MagicBytes: []byte("PK\x03\x04"),
		RequiredEntries: []string{
			"ppt/presentation.xml",
			"_rels/.rels",
			"[Content_Types].xml",
			"ppt/_rels/presentation.xml.rels",
		},
	}
}
