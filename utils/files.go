package utils

import (
	"path/filepath"
	"strings"
)

func ReplaceTarballWithIndex(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	base = strings.TrimSuffix(base, ".tar.gz")

	return filepath.Join(dir, base+".index")
}
