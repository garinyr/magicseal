package magicseal

// makeFormatXLSX returns the format definition for Office Open XML Spreadsheet (.xlsx).
func makeFormatXLSX() Format {
	return Format{
		Name:       FormatXLSX,
		Extensions: []string{".xlsx"},
		MIMEType:   "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		MagicBytes: []byte("PK\x03\x04"),
		RequiredEntries: []string{
			"xl/workbook.xml",
			"_rels/.rels",
			"[Content_Types].xml",
			"xl/_rels/workbook.xml.rels",
		},
	}
}
