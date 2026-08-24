package cli

import (
	"fmt"
	"io"
	"runtime"

	"claude-backup/internal/system"
)

// runUninstall deactivates and removes the platform scheduler config,
// leaving already-uploaded S3 data and the local exclusion config
// untouched.
func runUninstall(sys system.System, args []string, stdout, stderr io.Writer) error {
	homeDir, err := resolveHomeDir(sys, "uninstall")
	if err != nil {
		return err
	}

	sched, err := schedulerFor(runtime.GOOS)
	if err != nil {
		return fmt.Errorf("uninstall: selecting scheduler: %w", err)
	}
	if err := sched.uninstall(sys, homeDir); err != nil {
		return fmt.Errorf("uninstall: removing schedule: %w", err)
	}

	fmt.Fprintln(stdout, "claude-backup schedule removed")
	return nil
}
