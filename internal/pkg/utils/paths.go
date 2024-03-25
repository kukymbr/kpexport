package utils

import (
	"os"
	"strings"
)

// FixSeparators replaces all path separators to the OS-correct.
func FixSeparators(path string) string {
	if path == "" {
		return ""
	}

	sepToReplace := '/'
	if os.PathSeparator == sepToReplace {
		sepToReplace = '\\'
	}

	return strings.ReplaceAll(path, string(sepToReplace), string(os.PathSeparator))
}
