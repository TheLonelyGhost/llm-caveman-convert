package proxy

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ValidateBinary checks that bin exists and is executable.
func ValidateBinary(bin string) error {
	var path string
	if strings.ContainsRune(bin, '/') {
		path = bin
	} else {
		var err error
		path, err = exec.LookPath(bin)
		if err != nil {
			return fmt.Errorf("caveman binary %q not found in PATH: %w", bin, err)
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("caveman binary %q: %w", path, err)
	}
	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("caveman binary %q is not executable", path)
	}
	return nil
}
