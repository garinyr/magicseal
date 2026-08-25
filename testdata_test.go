package magicseal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// --- Testdata fixture validation tests ---

func TestTestdata_ValidFormats(t *testing.T) {
	tests := []struct {
		path   string
		format FormatName
	}{
		{"testdata/valid/docx.zip", FormatDOCX},
		{"testdata/valid/xlsx.zip", FormatXLSX},
		{"testdata/valid/pptx.zip", FormatPPTX},
		{"testdata/valid/odt.zip", FormatODT},
		{"testdata/valid/ods.zip", FormatODS},
		{"testdata/valid/jar.zip", FormatJAR},
		{"testdata/valid/apk.zip", FormatAPK},
		{"testdata/valid/apk-unsigned.zip", FormatAPK},
		{"testdata/valid/apk-multidex.zip", FormatAPK},
		{"testdata/valid/xpi.zip", FormatXPI},
	}
	for _, tt := range tests {
		t.Run(string(tt.format)+"/"+filepath.Base(tt.path), func(t *testing.T) {
			if err := ValidateFromPath(tt.path, tt.format); err != nil {
				t.Errorf("ValidateFromPath(%s, %s): unexpected error: %v", tt.path, tt.format, err)
			}
		})
	}
}

// --- Benchmark tests ---

func BenchmarkValidateInMemory(b *testing.B) {
	data := makeMinimalZip(b, []string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Validate(data, FormatDOCX); err != nil {
			b.Fatalf("Validate: %v", err)
		}
	}
}

func BenchmarkValidateReader(b *testing.B) {
	data := makeMinimalZip(b, []string{
		"word/document.xml",
		"_rels/.rels",
		"[Content_Types].xml",
	})
	r := &sliceReaderAt{data: data}
	size := int64(len(data))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ValidateReader(r, size, FormatDOCX); err != nil {
			b.Fatalf("ValidateReader: %v", err)
		}
	}
}

// sliceReaderAt implements io.ReaderAt for a byte slice.
type sliceReaderAt struct {
	data []byte
}

func (s *sliceReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(s.data)) {
		return 0, nil
	}
	n := copy(p, s.data[off:])
	return n, nil
}

// --- Test fixture generation ---

func TestGenerateTestdata(t *testing.T) {
	if os.Getenv("TESTDATA_GENERATE") != "1" {
		t.Skip("set TESTDATA_GENERATE=1 to regenerate test fixtures")
	}

	gen := func(path string, entries []string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", filepath.Dir(path), err)
		}
		data := makeMinimalZip(t, entries)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", path, err)
		}
		fmt.Println("  created", path)
	}

	// Valid formats
	gen("testdata/valid/docx.zip", []string{"word/document.xml", "_rels/.rels", "[Content_Types].xml"})
	gen("testdata/valid/xlsx.zip", []string{"xl/workbook.xml", "_rels/.rels", "[Content_Types].xml", "xl/_rels/workbook.xml.rels"})
	gen("testdata/valid/pptx.zip", []string{"ppt/presentation.xml", "_rels/.rels", "[Content_Types].xml", "ppt/_rels/presentation.xml.rels"})
	gen("testdata/valid/odt.zip", []string{"content.xml", "META-INF/manifest.xml", "mimetype"})
	gen("testdata/valid/ods.zip", []string{"content.xml", "META-INF/manifest.xml", "mimetype"})
	gen("testdata/valid/jar.zip", []string{"META-INF/MANIFEST.MF"})
	gen("testdata/valid/apk.zip", []string{"AndroidManifest.xml", "classes.dex"})
	gen("testdata/valid/apk-unsigned.zip", []string{"AndroidManifest.xml", "classes.dex"})
	gen("testdata/valid/apk-multidex.zip", []string{"AndroidManifest.xml", "classes.dex", "classes2.dex", "classes3.dex"})
	gen("testdata/valid/xpi.zip", []string{"manifest.json"})

	// Invalid: renamed ZIP claimed as DOCX
	gen("testdata/invalid/rename/plain-zip.docx", []string{"readme.txt", "images/photo.png"})

	// Invalid: corrupt central directory
	{
		valid := makeMinimalZip(t, []string{"word/document.xml", "_rels/.rels", "[Content_Types].xml"})
		corrupt := make([]byte, len(valid))
		copy(corrupt, valid)
		for i := len(corrupt) - 15; i < len(corrupt); i++ {
			corrupt[i] ^= 0xFF
		}
		if err := os.MkdirAll("testdata/invalid/corrupt-cd", 0o755); err != nil {
			t.Fatalf("MkdirAll(corrupt-cd): %v", err)
		}
		if err := os.WriteFile("testdata/invalid/corrupt-cd/broken.zip", corrupt, 0o644); err != nil {
			t.Fatalf("WriteFile(broken.zip): %v", err)
		}
		fmt.Println("  created testdata/invalid/corrupt-cd/broken.zip")
	}

	// Invalid: missing required entry
	gen("testdata/invalid/missing-entry/docx-missing.zip", []string{"_rels/.rels", "[Content_Types].xml"})

	// Invalid: path traversal
	gen("testdata/invalid/path-traversal/dotdot-slash.zip", []string{"word/document.xml", "../evil.exe"})

	// Edge: JAR without MANIFEST.MF
	gen("testdata/edge/jar-no-manifest/no-manifest.jar", []string{"com/example/Main.class"})
}
