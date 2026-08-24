package cli

import (
	"errors"
	"strings"
	"testing"
	"time"

	"claude-backup/internal/system"
)

func TestLinuxScheduler_InstallGeneratesUnitsAndEnablesTimer(t *testing.T) {
	fake := system.NewFake()
	fake.ExecutableValue = "/usr/local/bin/claude-backup"
	sched := linuxScheduler{}

	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("install: %v", err)
	}

	serviceContent, err := fake.ReadFile("/home/fake/.config/systemd/user/claude-backup.service")
	if err != nil {
		t.Fatalf("expected service unit written: %v", err)
	}
	if !strings.Contains(string(serviceContent), "ExecStart=/usr/local/bin/claude-backup sync") {
		t.Errorf("service unit = %q, want an ExecStart running sync", serviceContent)
	}

	timerContent, err := fake.ReadFile("/home/fake/.config/systemd/user/claude-backup.timer")
	if err != nil {
		t.Fatalf("expected timer unit written: %v", err)
	}
	for _, want := range []string{"OnCalendar=", "03:00:00", "Persistent=true"} {
		if !strings.Contains(string(timerContent), want) {
			t.Errorf("timer unit missing %q:\n%s", want, timerContent)
		}
	}

	var reloaded, enabled bool
	for _, c := range fake.Commands {
		if c.Name != "systemctl" {
			continue
		}
		joined := strings.Join(c.Args, " ")
		if joined == "--user daemon-reload" {
			reloaded = true
		}
		if joined == "--user enable --now claude-backup.timer" {
			enabled = true
		}
	}
	if !reloaded {
		t.Errorf("commands = %+v, want a systemctl --user daemon-reload", fake.Commands)
	}
	if !enabled {
		t.Errorf("commands = %+v, want a systemctl --user enable --now claude-backup.timer", fake.Commands)
	}
}

func TestLinuxScheduler_RerunningInstallUpdatesInPlace(t *testing.T) {
	fake := system.NewFake()
	fake.ExecutableValue = "/usr/local/bin/claude-backup"
	sched := linuxScheduler{}

	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("first install: %v", err)
	}
	fake.ExecutableValue = "/opt/claude-backup/claude-backup"
	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("second install: %v", err)
	}

	writesToService := 0
	for _, w := range fake.Writes {
		if w.Path == "/home/fake/.config/systemd/user/claude-backup.service" {
			writesToService++
		}
	}
	if writesToService != 2 {
		t.Fatalf("got %d writes to service unit, want exactly 2 (no duplication)", writesToService)
	}

	content, err := fake.ReadFile("/home/fake/.config/systemd/user/claude-backup.service")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(content), "/opt/claude-backup/claude-backup") {
		t.Errorf("service unit = %q, want updated executable path", content)
	}
}

func TestLinuxScheduler_UninstallDisablesAndRemovesUnits(t *testing.T) {
	fake := system.NewFake()
	sched := linuxScheduler{}
	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("install: %v", err)
	}

	if err := sched.uninstall(fake, "/home/fake"); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	for _, path := range []string{
		"/home/fake/.config/systemd/user/claude-backup.service",
		"/home/fake/.config/systemd/user/claude-backup.timer",
	} {
		exists, err := fake.FileExists(path)
		if err != nil {
			t.Fatalf("FileExists(%s): %v", path, err)
		}
		if exists {
			t.Errorf("expected %s removed after uninstall", path)
		}
	}

	var disabled bool
	for _, c := range fake.Commands {
		if c.Name == "systemctl" && strings.Join(c.Args, " ") == "--user disable --now claude-backup.timer" {
			disabled = true
		}
	}
	if !disabled {
		t.Errorf("commands = %+v, want a systemctl --user disable --now claude-backup.timer", fake.Commands)
	}
}

func TestLinuxScheduler_UninstallBeforeInstallIsNoop(t *testing.T) {
	fake := system.NewFake()
	sched := linuxScheduler{}

	if err := sched.uninstall(fake, "/home/fake"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.Commands) != 0 {
		t.Errorf("commands = %+v, want no systemctl invocation when nothing was installed", fake.Commands)
	}
}

func TestLinuxScheduler_UninstallDoesNotTouchConfigOrData(t *testing.T) {
	fake := system.NewFake()
	sched := linuxScheduler{}
	if err := saveConfig(fake, "/home/fake", Config{Remote: "s3remote"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("install: %v", err)
	}

	if err := sched.uninstall(fake, "/home/fake"); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	exists, err := fake.FileExists(configPath("/home/fake"))
	if err != nil {
		t.Fatalf("FileExists: %v", err)
	}
	if !exists {
		t.Error("expected config.json to survive uninstall")
	}
}

func TestLinuxScheduler_StatusReportsInactiveBeforeInstall(t *testing.T) {
	fake := system.NewFake()
	sched := linuxScheduler{}

	got, err := sched.status(fake, "/home/fake")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Active {
		t.Error("expected inactive before install")
	}
	if !got.NextRun.IsZero() {
		t.Errorf("NextRun = %v, want zero value before install", got.NextRun)
	}
}

func TestLinuxScheduler_StatusReportsActiveAndNextRunAfterInstall(t *testing.T) {
	fake := system.NewFake()
	fake.NowValue = time.Date(2026, 8, 23, 2, 0, 0, 0, time.UTC)
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		if name == "systemctl" && len(args) > 0 && args[len(args)-2] == "is-active" {
			return system.CommandResult{Stdout: "active\n"}, nil
		}
		return system.CommandResult{}, nil
	}
	sched := linuxScheduler{}
	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("install: %v", err)
	}

	got, err := sched.status(fake, "/home/fake")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Active {
		t.Error("expected active after install")
	}
	want := time.Date(2026, 8, 23, 3, 0, 0, 0, time.UTC)
	if !got.NextRun.Equal(want) {
		t.Errorf("NextRun = %v, want %v", got.NextRun, want)
	}

	var queried bool
	for _, c := range fake.Commands {
		if c.Name == "systemctl" && strings.Join(c.Args, " ") == "--user is-active claude-backup.timer" {
			queried = true
		}
	}
	if !queried {
		t.Errorf("commands = %+v, want a systemctl --user is-active claude-backup.timer query", fake.Commands)
	}
}

func TestLinuxScheduler_StatusReportsInactiveWhenTimerNotActive(t *testing.T) {
	fake := system.NewFake()
	sched := linuxScheduler{}
	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("install: %v", err)
	}
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		if name == "systemctl" && len(args) > 0 && args[len(args)-2] == "is-active" {
			return system.CommandResult{Stdout: "inactive\n", ExitCode: 3}, errors.New("exit status 3")
		}
		return system.CommandResult{}, nil
	}

	got, err := sched.status(fake, "/home/fake")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Active {
		t.Error("expected inactive when systemctl reports the timer inactive")
	}
	if !got.NextRun.IsZero() {
		t.Errorf("NextRun = %v, want zero value when inactive (nothing will actually fire)", got.NextRun)
	}
}
