package system

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// Real is the System implementation backed by the actual OS.
type Real struct{}

// NewReal returns a System that runs real commands and touches the real
// filesystem.
func NewReal() *Real {
	return &Real{}
}

func (r *Real) Run(name string, args ...string) (CommandResult, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := CommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
	}

	return result, err
}

func (r *Real) RunInteractive(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *Real) Hostname() (string, error) {
	return os.Hostname()
}

func (r *Real) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (r *Real) WriteFile(path string, content []byte, perm os.FileMode) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, content, perm)
}

func (r *Real) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (r *Real) FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
