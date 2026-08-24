// Package system is the single seam through which claude-backup's
// subcommand logic touches the outside world: running external commands
// (rclone, launchctl, systemctl) and reading/writing files (scheduler
// config, the backup config, the sync log). Subcommand logic depends only
// on the System interface; Real backs it with the actual OS for production
// use, Fake backs it with an in-memory store for tests.
package system

import "os"

// CommandResult is the outcome of running an external command through
// System.Run.
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// System abstracts every point of contact with the outside world that
// claude-backup's orchestration logic needs.
type System interface {
	// Run executes an external command and captures its result. A non-zero
	// exit is reported both via a non-nil error and via ExitCode on the
	// returned CommandResult.
	Run(name string, args ...string) (CommandResult, error)

	// RunInteractive executes an external command with stdin, stdout, and
	// stderr connected directly to the calling process's own, for wizards
	// that need real terminal interaction (e.g. the `rclone config`
	// wizard). It reports only whether the command succeeded.
	RunInteractive(name string, args ...string) error

	// WriteFile writes content to path, creating parent directories as
	// needed.
	WriteFile(path string, content []byte, perm os.FileMode) error

	// ReadFile reads the content at path. It returns an error wrapping
	// os.ErrNotExist if path has not been written.
	ReadFile(path string) ([]byte, error)

	// FileExists reports whether path has been written.
	FileExists(path string) (bool, error)

	// Hostname reports the local machine's hostname.
	Hostname() (string, error)

	// UserHomeDir reports the current user's home directory.
	UserHomeDir() (string, error)
}
