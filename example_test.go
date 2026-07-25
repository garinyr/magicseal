package magicseal_test

import (
	"archive/zip"
	"bytes"
	"fmt"

	"github.com/garinyr/magicseal"
)

// minimalDOCX creates a minimal valid DOCX in memory (ZIP with required entries).
func minimalDOCX() []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	w.Create("word/document.xml")
	w.Create("_rels/.rels")
	w.Create("[Content_Types].xml")
	w.Close()
	return buf.Bytes()
}

func ExampleValidate() {
	data := minimalDOCX()
	err := magicseal.Validate(data, magicseal.FormatDOCX)
	if err != nil {
		fmt.Println("invalid:", err)
		return
	}
	fmt.Println("valid DOCX")
	// Output: valid DOCX
}

func ExampleValidate_unsupported() {
	err := magicseal.Validate([]byte("hello"), "unknown")
	fmt.Println(err)
	// Output: magicseal: unsupported format: "unknown"
}

