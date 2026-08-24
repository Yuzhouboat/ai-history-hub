package system_test

import (
	"os"
	"path/filepath"
	"testing"

	"claude-backup/internal/system"
)

func TestReal_RunCapturesOutputAndExitCode(t *testing.T) {
	sys := system.NewReal()

	result, err := sys.Run("sh", "-c", "echo out; echo err >&2; exit 3")

	if err == nil {
		t.Fatalf("expected an error for non-zero exit, got nil")
	}
	if result.Stdout != "out\n" {
		t.Errorf("stdout = %q, want %q", result.Stdout, "out\n")
	}
	if result.Stderr != "err\n" {
		t.Errorf("stderr = %q, want %q", result.Stderr, "err\n")
	}
	if result.ExitCode != 3 {
		t.Errorf("exit code = %d, want 3", result.ExitCode)
	}
}

func TestReal_RunSucceeds(t *testing.T) {
	sys := system.NewReal()

	result, err := sys.Run("sh", "-c", "echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("stdout = %q, want %q", result.Stdout, "hello\n")
	}
	if result.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", result.ExitCode)
	}
}

func TestReal_WriteFileThenReadFileRoundTrips(t *testing.T) {
	sys := system.NewReal()
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.json")

	if err := sys.WriteFile(path, []byte(`{"remote":"s3"}`), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := sys.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != `{"remote":"s3"}` {
		t.Errorf("content = %q, want %q", got, `{"remote":"s3"}`)
	}
}

func TestReal_FileExists(t *testing.T) {
	sys := system.NewReal()
	dir := t.TempDir()
	path := filepath.Join(dir, "present.txt")

	exists, err := sys.FileExists(path)
	if err != nil {
		t.Fatalf("FileExists (absent): %v", err)
	}
	if exists {
		t.Errorf("expected FileExists to be false before file is written")
	}

	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	exists, err = sys.FileExists(path)
	if err != nil {
		t.Fatalf("FileExists (present): %v", err)
	}
	if !exists {
		t.Errorf("expected FileExists to be true after file is written")
	}
}
