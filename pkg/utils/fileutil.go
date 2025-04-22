// Helper: AppendUniqueSuffix to add timestamp suffix to filenames.
package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// AppendUniqueSuffix generates a unique filename like: report_20240422_162045.json
func AppendUniqueSuffix(filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s%s", name, timestamp, ext)
}
