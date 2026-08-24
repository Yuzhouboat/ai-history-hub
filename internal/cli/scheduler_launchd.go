package cli

import (
	"fmt"
	"path/filepath"

	"claude-backup/internal/system"
)

// darwinLabel is the launchd agent label claude-backup registers itself
// under, and doubles as the generated plist's filename stem.
const darwinLabel = "com.claude-backup.sync"

// darwinScheduler wires the daily sync into launchd via a per-user launch
// agent under ~/Library/LaunchAgents.
type darwinScheduler struct{}

func (darwinScheduler) plistPath(homeDir string) string {
	return filepath.Join(homeDir, "Library", "LaunchAgents", darwinLabel+".plist")
}

func darwinPlistContent(execPath string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>sync</string>
	</array>
	<key>StartCalendarInterval</key>
	<dict>
		<key>Hour</key>
		<integer>%d</integer>
		<key>Minute</key>
		<integer>%d</integer>
	</dict>
</dict>
</plist>
`, darwinLabel, execPath, scheduledHour, scheduledMinute)
}

func (d darwinScheduler) install(sys system.System, homeDir string) error {
	execPath, err := sys.Executable()
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	path := d.plistPath(homeDir)
	// Best-effort: unload whatever is currently registered under this label
	// so a re-run updates the schedule in place instead of launchctl load
	// erroring on an already-loaded label. Harmless (and expected to fail)
	// when nothing is loaded yet.
	_, _ = sys.Run("launchctl", "unload", path)

	if err := sys.WriteFile(path, []byte(darwinPlistContent(execPath)), 0o644); err != nil {
		return fmt.Errorf("writing launch agent: %w", err)
	}
	if _, err := sys.Run("launchctl", "load", "-w", path); err != nil {
		return fmt.Errorf("launchctl load: %w", err)
	}
	return nil
}

func (d darwinScheduler) uninstall(sys system.System, homeDir string) error {
	path := d.plistPath(homeDir)
	exists, err := sys.FileExists(path)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	if _, err := sys.Run("launchctl", "unload", path); err != nil {
		return fmt.Errorf("launchctl unload: %w", err)
	}
	if err := sys.RemoveFile(path); err != nil {
		return fmt.Errorf("removing launch agent: %w", err)
	}
	return nil
}

func (d darwinScheduler) status(sys system.System, homeDir string) (scheduleStatus, error) {
	path := d.plistPath(homeDir)
	exists, err := sys.FileExists(path)
	if err != nil {
		return scheduleStatus{}, err
	}
	if !exists {
		return scheduleStatus{}, nil
	}

	_, runErr := sys.Run("launchctl", "list", darwinLabel)
	if runErr != nil {
		return scheduleStatus{}, nil
	}
	return scheduleStatus{Active: true, NextRun: nextDailyRun(sys.Now())}, nil
}
