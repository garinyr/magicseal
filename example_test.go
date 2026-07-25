package magicseal_test

import (
	"archive/zip"
	"bytes"
	"fmt"

	"github.com/garinyr/magicseal"
)

// minimalDOCX creates a minimal valid DOCX in memory (ZIP with required entries).
func minimalDOCX() ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	if _, err := w.Create("word/document.xml"); err != nil {
		return nil, err
	}
	if _, err := w.Create("_rels/.rels"); err != nil {
		return nil, err
	}
	if _, err := w.Create("[Content_Types].xml"); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ExampleValidate() {
	data, err := minimalDOCX()
	if err != nil {
		fmt.Println("invalid:", err)
		return
	}

	err = magicseal.Validate(data, magicseal.FormatDOCX)
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
