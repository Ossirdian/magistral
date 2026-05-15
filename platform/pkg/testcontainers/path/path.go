package path

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func RepoRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to resolve caller path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", "..")), nil
}
