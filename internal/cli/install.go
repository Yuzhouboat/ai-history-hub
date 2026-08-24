package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"

	"claude-backup/internal/system"
)

// repeatedFlag collects every occurrence of a repeatable flag, e.g.
// --exclude a --exclude b.
type repeatedFlag struct {
	values []string
}

func (r *repeatedFlag) String() string {
	return strings.Join(r.values, ",")
}

func (r *repeatedFlag) Set(v string) error {
	r.values = append(r.values, v)
	return nil
}

// runInstall validates rclone is present, resolves or creates the rclone
// remote, persists the resolved remote + exclusion list to config, and
// activates the daily backup on the host platform's own scheduler.
func runInstall(sys system.System, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(stderr)
	remote := fs.String("remote", "", "rclone remote name to reuse or create")
	var exclude repeatedFlag
	fs.Var(&exclude, "exclude", "project directory name to exclude from backup (repeatable)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *remote == "" {
		return errors.New("install: --remote is required")
	}

	if err := ensureRemote(sys, *remote, stdout); err != nil {
		return err
	}

	homeDir, err := resolveHomeDir(sys, "install")
	if err != nil {
		return err
	}

	cfg := Config{Remote: *remote, Exclude: exclude.values}
	if err := saveConfig(sys, homeDir, cfg); err != nil {
		return fmt.Errorf("install: saving config: %w", err)
	}

	sched, err := schedulerFor(runtime.GOOS)
	if err != nil {
		return fmt.Errorf("install: selecting scheduler: %w", err)
	}
	if err := sched.install(sys, homeDir); err != nil {
		return fmt.Errorf("install: scheduling daily sync: %w", err)
	}

	fmt.Fprintf(stdout, "claude-backup configured with remote %q\n", *remote)
	return nil
}

// ensureRemote reuses remote if rclone already knows it, otherwise shells
// out to rclone's own interactive config wizard to create it. Credentials
// are handled entirely inside that wizard; claude-backup never sees them.
func ensureRemote(sys system.System, remote string, stdout io.Writer) error {
	result, err := sys.Run("rclone", "listremotes")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("install: rclone not found on PATH; install it from https://rclone.org/install/ and try again")
		}
		return fmt.Errorf("install: checking rclone remotes: %w", err)
	}

	for _, line := range strings.Split(result.Stdout, "\n") {
		if strings.TrimSpace(line) == remote+":" {
			fmt.Fprintf(stdout, "reusing existing rclone remote %q\n", remote)
			return nil
		}
	}

	fmt.Fprintf(stdout, "remote %q not found; launching rclone config to create it\n", remote)
	if err := sys.RunInteractive("rclone", "config"); err != nil {
		return fmt.Errorf("install: rclone config: %w", err)
	}
	return nil
}
