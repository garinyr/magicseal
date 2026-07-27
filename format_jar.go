package magicseal

// makeFormatJAR returns the format definition for Java Archive (.jar).
//
// Note: META-INF/MANIFEST.MF is technically optional per the JAR spec
// (the jar tool can skip it with the -M flag), but in practice nearly all
// real-world JAR files ship with a manifest. This may become a soft-warning
// instead of a hard-reject if missing-manifest JARs prove common.
func makeFormatJAR() Format {
	return Format{
		Name:       FormatJAR,
		Extensions: []string{".jar"},
		MIMEType:   "application/java-archive",
		MagicBytes: []byte("PK\x03\x04"),
		RequiredEntries: []string{
			"META-INF/MANIFEST.MF",
		},
	}
}
