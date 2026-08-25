package magicseal

// makeFormatXPI returns the format definition for Firefox Extension (.xpi).
//
// Only manifest.json is required — Firefox dropped support for legacy
// install.rdf-based extensions since Firefox 57 (2017). In the real world,
// XPI == WebExtension == manifest.json only. AND-based RequiredEntries
// (requiring both manifest.json AND install.rdf) would reject all valid XPIs.
func makeFormatXPI() Format {
	return Format{
		Name:       FormatXPI,
		Extensions: []string{".xpi"},
		MIMEType:   "application/x-xpinstall",
		MagicBytes: []byte("PK\x03\x04"),
		RequiredEntries: []string{
			"manifest.json",
		},
	}
}
