package proxy_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/proxy"
)

func TestValidateBinaryNotInPath(t *testing.T) {
	err := proxy.ValidateBinary("no-such-binary-xyz-12345")
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestValidateBinaryNotExecutable(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "notexec-*")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if err := os.Chmod(f.Name(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := proxy.ValidateBinary(f.Name()); err == nil {
		t.Fatal("expected error for non-executable file")
	}
}

func TestValidateBinaryExecutable(t *testing.T) {
	bin := buildCavemanBinary(t)
	if err := proxy.ValidateBinary(bin); err != nil {
		t.Fatalf("expected no error for valid executable: %v", err)
	}
}

func TestValidateBinaryAbsPathNotFound(t *testing.T) {
	err := proxy.ValidateBinary(filepath.Join(t.TempDir(), "nonexistent"))
	if err == nil {
		t.Fatal("expected error for missing absolute path")
	}
}
