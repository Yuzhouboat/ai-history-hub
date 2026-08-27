package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"claude-backup/internal/system"
)

// runStatus reports the active configuration (remote, exclusions), whether
// the daily backup schedule is active, the outcome of the last sync
// attempt, and when the next run is due.
func runStatus(sys system.System, args []string, stdout, stderr io.Writer) error {
	homeDir, err := resolveHomeDir(sys, "status")
	if err != nil {
		return err
	}

	cfg, err := loadConfig(sys, homeDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("status: no backup configured yet; run `claude-backup install` first")
		}
		return fmt.Errorf("status: loading config: %w", err)
	}

	sched, err := schedulerFor(runtime.GOOS)
	if err != nil {
		return fmt.Errorf("status: selecting scheduler: %w", err)
	}
	sc, err := sched.status(sys, homeDir)
	if err != nil {
		return fmt.Errorf("status: querying schedule: %w", err)
	}

	fmt.Fprintf(stdout, "remote: %s\n", cfg.Remote)
	fmt.Fprintf(stdout, "excluded projects: %s\n", excludedLabel(cfg.Exclude))
	fmt.Fprintf(stdout, "schedule: %s\n", activeLabel(sc.Active))
	fmt.Fprintf(stdout, "last sync: %s\n", lastSyncLabel(sys, homeDir))
	if errMsg, ok := lastSyncErrorLabel(sys, homeDir); ok {
		fmt.Fprintf(stdout, "last error: %s\n", errMsg)
	}
	fmt.Fprintf(stdout, "next run: %s\n", nextRunLabel(sc))
	return nil
}

func excludedLabel(exclude []string) string {
	if len(exclude) == 0 {
		return "none"
	}
	return strings.Join(exclude, ", ")
}

func activeLabel(active bool) string {
	if active {
		return "active"
	}
	return "not active"
}

// lastSyncLabel reports when the last sync attempt completed and, for a
// failed attempt, flags it as such so a stale "last sync" time doesn't read
// as a healthy backup.
func lastSyncLabel(sys system.System, homeDir string) string {
	attempt, ok := lastSyncAttempt(sys, homeDir)
	if !ok {
		return "never"
	}
	if attempt.Success {
		return attempt.Time.Format(time.RFC3339)
	}
	return attempt.Time.Format(time.RFC3339) + " (failed)"
}

// lastSyncErrorLabel returns the error recorded for the last sync attempt,
// if that attempt failed.
func lastSyncErrorLabel(sys system.System, homeDir string) (string, bool) {
	attempt, ok := lastSyncAttempt(sys, homeDir)
	if !ok || attempt.Success {
		return "", false
	}
	return attempt.Error, true
}

func nextRunLabel(sc scheduleStatus) string {
	if sc.NextRun.IsZero() {
		return "not scheduled"
	}
	return sc.NextRun.Format(time.RFC3339)
}
