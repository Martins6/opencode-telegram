package service

import (
	"os"
	"path/filepath"
	"strings"
)

func mkdir(p string) error               { return os.MkdirAll(p, 0o755) }
func writeFile(p string, b string) error { return os.WriteFile(p, []byte(b), 0o755) }
func contains(s, sub string) bool        { return strings.Contains(s, sub) }
func dirOf(p string) string              { return filepath.Dir(p) }
