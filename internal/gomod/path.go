package gomod

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModuleImportPath returns the Go import path for outDir based on the nearest go.mod.
func ModuleImportPath(outDir string) (string, error) {
	abs, err := filepath.Abs(outDir)
	if err != nil {
		return "", err
	}
	dir := abs
	for {
		modPath := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(modPath)
		if err == nil {
			module, err := parseModulePath(data)
			if err != nil {
				return "", err
			}
			rel, err := filepath.Rel(dir, abs)
			if err != nil {
				return "", err
			}
			rel = filepath.ToSlash(rel)
			if rel == "." {
				return module, nil
			}
			return module + "/" + rel, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found for %s", outDir)
}

func parseModulePath(data []byte) (string, error) {
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module directive not found")
}
