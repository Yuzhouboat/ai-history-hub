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
	case "-v", "--version", "version":
		runVersion(stdout)
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
	fmt.Fprintln(w, "claude-backup ships the local chat history Claude Code keeps under")
	fmt.Fprintln(w, "~/.claude (session transcripts and the global prompt index) to an S3")
	fmt.Fprintln(w, "bucket via rclone, additive-only and on a daily schedule it manages")
	fmt.Fprintln(w, "for you through the host platform's own scheduler (launchd/systemd).")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  claude-backup install [--remote <name>] [--exclude <project>]...")
	fmt.Fprintln(w, "      Configure the S3 remote and activate the daily backup schedule.")
	fmt.Fprintln(w, "  claude-backup sync")
	fmt.Fprintln(w, "      Run a backup immediately, outside the daily schedule.")
	fmt.Fprintln(w, "  claude-backup status")
	fmt.Fprintln(w, "      Show the configured remote, schedule state, and last sync result.")
	fmt.Fprintln(w, "  claude-backup uninstall")
	fmt.Fprintln(w, "      Deactivate the daily schedule. Leaves S3 data and config untouched.")
	fmt.Fprintln(w, "  claude-backup version")
	fmt.Fprintln(w, "      Print the claude-backup version.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  claude-backup install --remote s3remote")
	fmt.Fprintln(w, "  claude-backup install --remote s3remote --exclude side-project")
	fmt.Fprintln(w, "  claude-backup status")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Run `claude-backup install -h` for install's flags.")
}
