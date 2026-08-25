package magicseal

// makeFormatAPK returns the format definition for Android Package (.apk).
//
// Only AndroidManifest.xml and classes.dex are required. META-INF/ is NOT
// required because unsigned/debug APK builds lack that directory entirely —
// they are valid APKs that would be falsely rejected if META-INF/ were listed.
func makeFormatAPK() Format {
	return Format{
		Name:       FormatAPK,
		Extensions: []string{".apk"},
		MIMEType:   "application/vnd.android.package-archive",
		MagicBytes: []byte("PK\x03\x04"),
		RequiredEntries: []string{
			"AndroidManifest.xml",
			"classes.dex",
		},
	}
}
