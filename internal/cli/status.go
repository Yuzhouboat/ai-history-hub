package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"claude-backup/internal/system"
)

// runStatus reports whether the daily backup schedule is active, when it
// last succeeded, and when it's next due to run.
func runStatus(sys system.System, args []string, stdout, stderr io.Writer) error {
	homeDir, err := resolveHomeDir(sys, "status")
	if err != nil {
		return err
	}

	if _, err := loadConfig(sys, homeDir); err != nil {
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

	fmt.Fprintf(stdout, "schedule: %s\n", activeLabel(sc.Active))
	fmt.Fprintf(stdout, "last sync: %s\n", lastSyncLabel(sys, homeDir))
	fmt.Fprintf(stdout, "next run: %s\n", nextRunLabel(sc))
	return nil
}

func activeLabel(active bool) string {
	if active {
		return "active"
	}
	return "not active"
}

func lastSyncLabel(sys system.System, homeDir string) string {
	ts, ok := lastSyncDone(sys, homeDir)
	if !ok {
		return "never"
	}
	return ts.Format(time.RFC3339)
}

func nextRunLabel(sc scheduleStatus) string {
	if sc.NextRun.IsZero() {
		return "not scheduled"
	}
	return sc.NextRun.Format(time.RFC3339)
}
