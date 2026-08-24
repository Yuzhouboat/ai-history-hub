package cli

import (
	"errors"
	"strings"
	"testing"
	"time"

	"claude-backup/internal/system"
)

func TestDarwinScheduler_InstallGeneratesPlistAndLoadsAgent(t *testing.T) {
	fake := system.NewFake()
	fake.ExecutableValue = "/usr/local/bin/claude-backup"
	sched := darwinScheduler{}

	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("install: %v", err)
	}

	wantPath := "/home/fake/Library/LaunchAgents/com.claude-backup.sync.plist"
	content, err := fake.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("expected plist written at %s: %v", wantPath, err)
	}
	for _, want := range []string{
		"<key>Label</key>",
		"<string>com.claude-backup.sync</string>",
		"<string>/usr/local/bin/claude-backup</string>",
		"<string>sync</string>",
		"<key>StartCalendarInterval</key>",
		"<integer>3</integer>",
		"<integer>0</integer>",
	} {
		if !strings.Contains(string(content), want) {
			t.Errorf("plist content missing %q:\n%s", want, content)
		}
	}

	var loaded bool
	for _, c := range fake.Commands {
		if c.Name == "launchctl" && len(c.Args) >= 2 && c.Args[0] == "load" && c.Args[len(c.Args)-1] == wantPath {
			loaded = true
		}
	}
	if !loaded {
		t.Errorf("commands = %+v, want a launchctl load of the generated plist", fake.Commands)
	}
}

func TestDarwinScheduler_RerunningInstallUpdatesInPlace(t *testing.T) {
	fake := system.NewFake()
	fake.ExecutableValue = "/usr/local/bin/claude-backup"
	sched := darwinScheduler{}

	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("first install: %v", err)
	}
	fake.ExecutableValue = "/opt/claude-backup/claude-backup"
	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("second install: %v", err)
	}

	wantPath := "/home/fake/Library/LaunchAgents/com.claude-backup.sync.plist"
	writesToPlist := 0
	for _, w := range fake.Writes {
		if w.Path == wantPath {
			writesToPlist++
		}
	}
	if writesToPlist != 2 {
		t.Fatalf("got %d writes to plist, want exactly 2 (no duplication)", writesToPlist)
	}

	unloads, loads := 0, 0
	for _, c := range fake.Commands {
		if c.Name != "launchctl" {
			continue
		}
		switch c.Args[0] {
		case "unload":
			unloads++
		case "load":
			loads++
		}
	}
	if unloads != 2 || loads != 2 {
		t.Errorf("launchctl unload/load counts = %d/%d, want 2/2 (unload-then-load each run, no duplicate schedule)", unloads, loads)
	}

	content, err := fake.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(content), "/opt/claude-backup/claude-backup") {
		t.Errorf("plist = %q, want updated executable path", content)
	}
}

func TestDarwinScheduler_UninstallUnloadsAndRemovesPlist(t *testing.T) {
	fake := system.NewFake()
	sched := darwinScheduler{}
	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("install: %v", err)
	}

	if err := sched.uninstall(fake, "/home/fake"); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	wantPath := "/home/fake/Library/LaunchAgents/com.claude-backup.sync.plist"
	exists, err := fake.FileExists(wantPath)
	if err != nil {
		t.Fatalf("FileExists: %v", err)
	}
	if exists {
		t.Error("expected plist removed after uninstall")
	}

	var unloaded bool
	for _, c := range fake.Commands {
		if c.Name == "launchctl" && len(c.Args) >= 2 && c.Args[0] == "unload" && c.Args[len(c.Args)-1] == wantPath {
			unloaded = true
		}
	}
	if !unloaded {
		t.Errorf("commands = %+v, want a launchctl unload of the plist", fake.Commands)
	}
}

func TestDarwinScheduler_UninstallBeforeInstallIsNoop(t *testing.T) {
	fake := system.NewFake()
	sched := darwinScheduler{}

	if err := sched.uninstall(fake, "/home/fake"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.Commands) != 0 {
		t.Errorf("commands = %+v, want no launchctl invocation when nothing was installed", fake.Commands)
	}
}

func TestDarwinScheduler_UninstallDoesNotTouchConfigOrData(t *testing.T) {
	fake := system.NewFake()
	sched := darwinScheduler{}
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

func TestDarwinScheduler_StatusReportsInactiveBeforeInstall(t *testing.T) {
	fake := system.NewFake()
	sched := darwinScheduler{}

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

func TestDarwinScheduler_StatusReportsActiveAndNextRunAfterInstall(t *testing.T) {
	fake := system.NewFake()
	fake.NowValue = time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	sched := darwinScheduler{}
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
	want := time.Date(2026, 8, 24, 3, 0, 0, 0, time.UTC)
	if !got.NextRun.Equal(want) {
		t.Errorf("NextRun = %v, want %v", got.NextRun, want)
	}

	var listed bool
	for _, c := range fake.Commands {
		if c.Name == "launchctl" && len(c.Args) >= 2 && c.Args[0] == "list" && c.Args[1] == "com.claude-backup.sync" {
			listed = true
		}
	}
	if !listed {
		t.Errorf("commands = %+v, want a launchctl list query for the label", fake.Commands)
	}
}

func TestDarwinScheduler_StatusReportsInactiveWhenLaunchctlReportsUnloaded(t *testing.T) {
	fake := system.NewFake()
	sched := darwinScheduler{}
	if err := sched.install(fake, "/home/fake"); err != nil {
		t.Fatalf("install: %v", err)
	}
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		if name == "launchctl" && len(args) > 0 && args[0] == "list" {
			return system.CommandResult{ExitCode: 1}, errors.New("Could not find service")
		}
		return system.CommandResult{}, nil
	}

	got, err := sched.status(fake, "/home/fake")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Active {
		t.Error("expected inactive when launchctl reports the label unloaded")
	}
	if !got.NextRun.IsZero() {
		t.Errorf("NextRun = %v, want zero value when inactive (nothing will actually fire)", got.NextRun)
	}
}
