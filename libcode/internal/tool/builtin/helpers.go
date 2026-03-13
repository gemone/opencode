// Package builtin provides built-in tool implementations
package builtin

import (
	"path/filepath"
)

// resolvePath resolves a path relative to the working directory.
// If the path is already absolute, it is returned unchanged.
func resolvePath(path, workingDir string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(workingDir, path)
}
