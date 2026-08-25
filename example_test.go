package magicseal_test

import (
	"archive/zip"
	"bytes"
	"fmt"

	"github.com/garinyr/magicseal"
)

// minimalDOCX creates a minimal valid DOCX in memory (ZIP with required entries).
// It returns an error if the ZIP cannot be constructed.
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
	// Output: magicseal: unsupported format: "unknown" [format=unknown] [code=1]
}

func ExampleRegisterFormat() {
	// Register a custom format
	err := magicseal.RegisterFormat(magicseal.Format{
		Name:            "custom-fmt",
		Extensions:      []string{".customfmt"},
		MagicBytes:      []byte("PK\x03\x04"),
		RequiredEntries: []string{"data.xml"},
	})
	if err != nil {
		fmt.Println("register failed:", err)
		return
	}

	// Clean up
	if err := magicseal.UnregisterFormat("custom-fmt"); err != nil {
		fmt.Println("unregister failed:", err)
		return
	}

	// Verify it's gone
	lookupErr := magicseal.Validate([]byte("PK\x03\x04"), "custom-fmt")
	fmt.Println("after unregister:", lookupErr)
	// Output: after unregister: magicseal: unsupported format: "custom-fmt" [format=custom-fmt] [code=1]
}

func ExampleValidateReader() {
	data, err := minimalDOCX()
	if err != nil {
		fmt.Println("invalid:", err)
		return
	}

	r := bytes.NewReader(data)
	err = magicseal.ValidateReader(r, int64(len(data)), magicseal.FormatDOCX)
	if err != nil {
		fmt.Println("invalid:", err)
		return
	}
	fmt.Println("valid DOCX")
	// Output: valid DOCX
}

func ExampleValidateWithConfig() {
	data, err := minimalDOCX()
	if err != nil {
		fmt.Println("invalid:", err)
		return
	}

	err = magicseal.ValidateWithConfig(data, magicseal.FormatDOCX,
		&magicseal.ValidatorConfig{
			MaxDecompressedSize: 128 << 20, // 128 MB
		})
	if err != nil {
		fmt.Println("invalid:", err)
		return
	}
	fmt.Println("valid DOCX")
	// Output: valid DOCX
}
