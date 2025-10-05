package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsPDF checks if a file is a PDF
func IsPDF(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".pdf")
}

// IsImage checks if a file is an image
func IsImage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".bmp"
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// GenerateOutputFileName generates a unique output filename
func GenerateOutputFileName(basePath, suffix, ext string) string {
	dir := filepath.Dir(basePath)
	name := strings.TrimSuffix(filepath.Base(basePath), filepath.Ext(basePath))
	return filepath.Join(dir, fmt.Sprintf("%s_%s%s", name, suffix, ext))
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

