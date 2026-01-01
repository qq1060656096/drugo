package gomod

import (
	"os"
	"path/filepath"
)

// GmodRoot
func GmodRoot(dir string) string {
	if dir == filepath.Dir(dir) {
		return dir
	}
	if GmodNotExist(dir) {
        return GmodRoot(filepath.Dir(dir)) // ✅ 向上查找父目录
	}
	return dir
}

// GmodNotExist
func GmodNotExist(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return os.IsNotExist(err)
}
