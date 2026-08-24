package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"claude-backup/internal/cli"
	"claude-backup/internal/system"
)

func TestUninstall_RemovesScheduleWiredUpByInstall(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		return system.CommandResult{Stdout: "s3remote:\n"}, nil
	}
	var stdout, stderr bytes.Buffer
	if err := cli.Run(fake, []string{"install", "--remote", "s3remote"}, &stdout, &stderr); err != nil {
		t.Fatalf("install: %v", err)
	}

	err := cli.Run(fake, []string{"uninstall"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for path := range fake.Files {
		if strings.Contains(path, "systemd") || strings.Contains(path, "LaunchAgents") {
			t.Errorf("expected scheduler file %q removed after uninstall", path)
		}
	}
}

func TestUninstall_LeavesConfigAndDataUntouched(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		return system.CommandResult{Stdout: "s3remote:\n"}, nil
	}
	var stdout, stderr bytes.Buffer
	if err := cli.Run(fake, []string{"install", "--remote", "s3remote"}, &stdout, &stderr); err != nil {
		t.Fatalf("install: %v", err)
	}

	if err := cli.Run(fake, []string{"uninstall"}, &stdout, &stderr); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	exists, err := fake.FileExists("/home/fake/.claude-backup/config.json")
	if err != nil {
		t.Fatalf("FileExists: %v", err)
	}
	if !exists {
		t.Error("expected config.json to survive uninstall")
	}
}

func TestUninstall_BeforeInstallIsNotAnError(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"uninstall"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
