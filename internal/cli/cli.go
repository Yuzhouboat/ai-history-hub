// Package cli dispatches claude-backup's subcommands. Every subcommand
// receives a system.System as its only way to affect the outside world, so
// tests can drive the whole CLI through system.Fake.
package cli

import (
	"errors"
	"fmt"
	"io"

	"claude-backup/internal/system"
)

// ErrNotImplemented is returned by a subcommand that is recognized but not
// yet built.
var ErrNotImplemented = errors.New("not yet implemented")

// Run dispatches args[0] to the matching subcommand. Usage/help goes to
// stdout; errors go to stderr.
func Run(sys system.System, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	switch args[0] {
	case "-h", "--help", "help":
		printUsage(stdout)
		return nil
	case "install":
		return runInstall(sys, args[1:], stdout, stderr)
	case "sync":
		return runSync(sys, args[1:], stdout, stderr)
	case "status":
		return runStatus(sys, args[1:], stdout, stderr)
	case "uninstall":
		return runUninstall(sys, args[1:], stdout, stderr)
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "claude-backup - back up Claude Code chat history to S3")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  claude-backup install [--remote <name>] [--exclude <project>]...")
	fmt.Fprintln(w, "  claude-backup sync")
	fmt.Fprintln(w, "  claude-backup status")
	fmt.Fprintln(w, "  claude-backup uninstall")
}
