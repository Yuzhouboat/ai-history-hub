package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"claude-backup/internal/system"
)

// linuxServiceUnit and linuxTimerUnit are the systemd user unit names
// claude-backup generates and manages.
const (
	linuxServiceUnit = "claude-backup.service"
	linuxTimerUnit   = "claude-backup.timer"
)

// linuxScheduler wires the daily sync into systemd via a per-user service +
// timer pair under ~/.config/systemd/user.
type linuxScheduler struct{}

func (linuxScheduler) unitDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "systemd", "user")
}

func (l linuxScheduler) servicePath(homeDir string) string {
	return filepath.Join(l.unitDir(homeDir), linuxServiceUnit)
}

func (l linuxScheduler) timerPath(homeDir string) string {
	return filepath.Join(l.unitDir(homeDir), linuxTimerUnit)
}

func linuxServiceContent(execPath string) string {
	return fmt.Sprintf(`[Unit]
Description=claude-backup daily sync

[Service]
Type=oneshot
ExecStart=%s sync
`, execPath)
}

func linuxTimerContent() string {
	return fmt.Sprintf(`[Unit]
Description=claude-backup daily sync timer

[Timer]
OnCalendar=*-*-* %02d:%02d:00
Persistent=true

[Install]
WantedBy=timers.target
`, scheduledHour, scheduledMinute)
}

func (l linuxScheduler) install(sys system.System, homeDir string) error {
	execPath, err := sys.Executable()
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	if err := sys.WriteFile(l.servicePath(homeDir), []byte(linuxServiceContent(execPath)), 0o644); err != nil {
		return fmt.Errorf("writing service unit: %w", err)
	}
	if err := sys.WriteFile(l.timerPath(homeDir), []byte(linuxTimerContent()), 0o644); err != nil {
		return fmt.Errorf("writing timer unit: %w", err)
	}
	if _, err := sys.Run("systemctl", "--user", "daemon-reload"); err != nil {
		return fmt.Errorf("systemctl daemon-reload: %w", err)
	}
	if _, err := sys.Run("systemctl", "--user", "enable", "--now", linuxTimerUnit); err != nil {
		return fmt.Errorf("systemctl enable --now: %w", err)
	}
	return nil
}

func (l linuxScheduler) uninstall(sys system.System, homeDir string) error {
	exists, err := sys.FileExists(l.timerPath(homeDir))
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	if _, err := sys.Run("systemctl", "--user", "disable", "--now", linuxTimerUnit); err != nil {
		return fmt.Errorf("systemctl disable --now: %w", err)
	}
	if err := sys.RemoveFile(l.servicePath(homeDir)); err != nil {
		return fmt.Errorf("removing service unit: %w", err)
	}
	if err := sys.RemoveFile(l.timerPath(homeDir)); err != nil {
		return fmt.Errorf("removing timer unit: %w", err)
	}
	if _, err := sys.Run("systemctl", "--user", "daemon-reload"); err != nil {
		return fmt.Errorf("systemctl daemon-reload: %w", err)
	}
	return nil
}

func (l linuxScheduler) status(sys system.System, homeDir string) (scheduleStatus, error) {
	exists, err := sys.FileExists(l.timerPath(homeDir))
	if err != nil {
		return scheduleStatus{}, err
	}
	if !exists {
		return scheduleStatus{}, nil
	}

	result, runErr := sys.Run("systemctl", "--user", "is-active", linuxTimerUnit)
	if !(runErr == nil && strings.TrimSpace(result.Stdout) == "active") {
		return scheduleStatus{}, nil
	}
	return scheduleStatus{Active: true, NextRun: nextDailyRun(sys.Now())}, nil
}
