package system

import (
	"fmt"
	"os"
	"time"
)

// fakeReferenceTime is the default Now() value when NowValue is unset,
// chosen to be an obviously-fake fixed point so a forgotten override in a
// time-sensitive test is easy to spot.
var fakeReferenceTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

// Command records one invocation made through Fake.Run.
type Command struct {
	Name string
	Args []string
}

// Write records one call made through Fake.WriteFile, in call order — unlike
// Files, which only holds the current content per path, Writes preserves a
// second write to the same path instead of silently losing the first.
type Write struct {
	Path    string
	Content []byte
}

// Fake is an in-memory System for tests: it records every command and file
// write instead of touching the real OS.
type Fake struct {
	Commands []Command
	Writes   []Write
	Files    map[string][]byte

	// RunFunc, if set, computes the result for each Run call. Commands are
	// recorded regardless of whether RunFunc is set.
	RunFunc func(name string, args ...string) (CommandResult, error)

	// InteractiveCommands records every RunInteractive invocation.
	InteractiveCommands []Command

	// RunInteractiveFunc, if set, computes the result for each
	// RunInteractive call. Commands are recorded regardless of whether
	// RunInteractiveFunc is set.
	RunInteractiveFunc func(name string, args ...string) error

	// HostnameValue is returned by Hostname. Defaults to "fake-host" if
	// unset.
	HostnameValue string

	// HomeDirValue is returned by UserHomeDir. Defaults to "/home/fake" if
	// unset.
	HomeDirValue string

	// ExecutableValue is returned by Executable. Defaults to
	// "/usr/local/bin/claude-backup" if unset.
	ExecutableValue string

	// NowValue is returned by Now. Defaults to fakeReferenceTime if unset.
	NowValue time.Time

	// Removals records every path deleted via RemoveFile, in call order.
	Removals []string
}

// NewFake returns an empty Fake System.
func NewFake() *Fake {
	return &Fake{Files: map[string][]byte{}}
}

func (f *Fake) Run(name string, args ...string) (CommandResult, error) {
	f.Commands = append(f.Commands, Command{Name: name, Args: args})
	if f.RunFunc != nil {
		return f.RunFunc(name, args...)
	}
	return CommandResult{}, nil
}

func (f *Fake) RunInteractive(name string, args ...string) error {
	f.InteractiveCommands = append(f.InteractiveCommands, Command{Name: name, Args: args})
	if f.RunInteractiveFunc != nil {
		return f.RunInteractiveFunc(name, args...)
	}
	return nil
}

func (f *Fake) Hostname() (string, error) {
	if f.HostnameValue != "" {
		return f.HostnameValue, nil
	}
	return "fake-host", nil
}

func (f *Fake) UserHomeDir() (string, error) {
	if f.HomeDirValue != "" {
		return f.HomeDirValue, nil
	}
	return "/home/fake", nil
}

func (f *Fake) WriteFile(path string, content []byte, _ os.FileMode) error {
	if f.Files == nil {
		f.Files = map[string][]byte{}
	}
	copied := append([]byte(nil), content...)
	f.Files[path] = copied
	f.Writes = append(f.Writes, Write{Path: path, Content: copied})
	return nil
}

func (f *Fake) ReadFile(path string) ([]byte, error) {
	content, ok := f.Files[path]
	if !ok {
		return nil, fmt.Errorf("%s: %w", path, os.ErrNotExist)
	}
	return content, nil
}

func (f *Fake) FileExists(path string) (bool, error) {
	_, ok := f.Files[path]
	return ok, nil
}

func (f *Fake) RemoveFile(path string) error {
	if _, ok := f.Files[path]; !ok {
		return fmt.Errorf("%s: %w", path, os.ErrNotExist)
	}
	delete(f.Files, path)
	f.Removals = append(f.Removals, path)
	return nil
}

func (f *Fake) Executable() (string, error) {
	if f.ExecutableValue != "" {
		return f.ExecutableValue, nil
	}
	return "/usr/local/bin/claude-backup", nil
}

func (f *Fake) Now() time.Time {
	if !f.NowValue.IsZero() {
		return f.NowValue
	}
	return fakeReferenceTime
}
