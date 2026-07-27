package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/garinyr/magicseal"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "magicseal:", err)
		os.Exit(2)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: magicseal <command> [options]\n  check <file> --as <format> [--json]\n  formats")
	}

	switch args[0] {
	case "check":
		return cmdCheck(args[1:])
	case "formats":
		return cmdFormats()
	default:
		return fmt.Errorf("unknown command: %q\nusage: magicseal <check|formats>", args[0])
	}
}

func cmdCheck(args []string) error {
	var file, format string
	jsonOut := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--as":
			if i+1 >= len(args) {
				return fmt.Errorf("--as requires a value")
			}
			i++
			format = args[i]
		case "--json":
			jsonOut = true
		default:
			if file == "" {
				file = args[i]
			}
		}
	}

	if file == "" {
		return fmt.Errorf("missing file argument")
	}
	if format == "" {
		return fmt.Errorf("missing --as <format>")
	}

	err := magicseal.ValidateFromPath(file, magicseal.FormatName(format))
	if err == nil {
		if jsonOut {
			fmt.Println(`{"status":"valid"}`)
		} else {
			fmt.Println("valid")
		}
		return nil
	}

	// Error handling
	var ve *magicseal.ValidationError
	if jsonOut {
		output := map[string]interface{}{
			"status": "invalid",
			"error":  err.Error(),
		}
		if errors.As(err, &ve) {
			output["code"] = int(ve.Code)
			output["format"] = string(ve.Format)
		}
		b, _ := json.Marshal(output)
		fmt.Println(string(b))
	} else {
		if errors.As(err, &ve) {
			fmt.Printf("invalid [code=%d]: %v\n", ve.Code, err)
		} else {
			fmt.Println("invalid:", err)
		}
	}

	// Exit code 1 for invalid, 2 for real errors.
	os.Exit(1)
	return nil
}

func cmdFormats() error {
	formats := magicseal.ListFormats()
	for _, f := range formats {
		fmt.Printf("%-8s %s  %s\n", f.Name, strings.Join(f.Extensions, ", "), f.MIMEType)
	}
	return nil
}
