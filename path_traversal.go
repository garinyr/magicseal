package magicseal

import (
	"fmt"
	"strings"
)

// checkPathTraversal validates all ZIP entry names for suspicious patterns.
// This runs against ALL entries (not just required entries) as defense-in-depth.
//
// Detected patterns:
//   - ../ (relative path traversal, forward slash)
//   - ..\ (relative path traversal, Windows-style)
//   - Absolute paths starting with /
//   - Null byte injection
//   - ".." as a complete path segment (checked after splitting by /)
//
// ZIP entry names always use forward-slash per the ZIP specification,
// so path.Split semantics are consistent across all platforms.
func checkPathTraversal(zipEntries []zipEntry) error {
	for _, e := range zipEntries {
		name := e.Name

		// Null byte injection check — file names in ZIP should never contain \x00.
		if strings.ContainsRune(name, 0) {
			return fmt.Errorf("%w: entry %q contains null byte", ErrPathTraversal, name)
		}

		// Windows-style backslash traversal: ..\
		if strings.Contains(name, `..\`) {
			return fmt.Errorf("%w: entry %q contains suspicious pattern %q",
				ErrPathTraversal, name, `..\`)
		}

		// Forward-slash traversal string: ../ anywhere.
		if strings.Contains(name, "../") {
			return fmt.Errorf("%w: entry %q contains suspicious pattern %q",
				ErrPathTraversal, name, "../")
		}

		// Absolute path starting with /
		if strings.HasPrefix(name, "/") {
			return fmt.Errorf("%w: entry %q is an absolute path", ErrPathTraversal, name)
		}

		// Check each path segment for ".." (bare double-dot).
		// Splitting by "/" catches: ".." alone, "deep/../../evil", etc.
		// Does NOT match: "..something.txt", "...config", "..hidden".
		for _, segment := range strings.Split(name, "/") {
			if segment == ".." {
				return fmt.Errorf("%w: entry %q contains \"..\" path segment",
					ErrPathTraversal, name)
			}
		}
	}
	return nil
}
