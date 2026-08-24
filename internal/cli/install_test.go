package cli_test

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"claude-backup/internal/cli"
	"claude-backup/internal/system"
)

func TestInstall_FailsFastWhenRcloneNotFound(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		if name == "rclone" {
			return system.CommandResult{}, exec.ErrNotFound
		}
		return system.CommandResult{}, nil
	}
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"install", "--remote", "s3remote"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("expected an error when rclone is missing")
	}
	if !strings.Contains(err.Error(), "rclone") {
		t.Errorf("error = %q, want it to mention rclone", err.Error())
	}
	exists, _ := fake.FileExists("/home/fake/.claude-backup/config.json")
	if exists {
		t.Error("expected no config written when rclone is missing")
	}
}

func TestInstall_ReusesExistingRemote(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		if name == "rclone" && len(args) > 0 && args[0] == "listremotes" {
			return system.CommandResult{Stdout: "s3remote:\nother:\n"}, nil
		}
		return system.CommandResult{}, nil
	}
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"install", "--remote", "s3remote"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.InteractiveCommands) != 0 {
		t.Errorf("expected no interactive wizard when remote already exists, got %v", fake.InteractiveCommands)
	}
	content, err := fake.ReadFile("/home/fake/.claude-backup/config.json")
	if err != nil {
		t.Fatalf("expected config to be written: %v", err)
	}
	if !strings.Contains(string(content), "s3remote") {
		t.Errorf("config = %q, want it to contain remote name", content)
	}
}

func TestInstall_CreatesRemoteViaInteractiveWizardWhenMissing(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		if name == "rclone" && len(args) > 0 && args[0] == "listremotes" {
			return system.CommandResult{Stdout: "other:\n"}, nil
		}
		return system.CommandResult{}, nil
	}
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"install", "--remote", "s3remote"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.InteractiveCommands) != 1 {
		t.Fatalf("expected the interactive rclone config wizard to run, got %v", fake.InteractiveCommands)
	}
	got := fake.InteractiveCommands[0]
	if got.Name != "rclone" || len(got.Args) == 0 || got.Args[0] != "config" {
		t.Errorf("interactive command = %+v, want rclone config", got)
	}
}

func TestInstall_NeverPassesAWSCredentials(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		return system.CommandResult{Stdout: "s3remote:\n"}, nil
	}
	var stdout, stderr bytes.Buffer

	if err := cli.Run(fake, []string{"install", "--remote", "s3remote"}, &stdout, &stderr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range fake.Commands {
		for _, a := range c.Args {
			lower := strings.ToLower(a)
			if strings.Contains(lower, "key") || strings.Contains(lower, "secret") {
				t.Errorf("command args leaked a credential-shaped value: %+v", c)
			}
		}
	}
}

func TestInstall_PersistsExcludeList(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		return system.CommandResult{Stdout: "s3remote:\n"}, nil
	}
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"install", "--remote", "s3remote", "--exclude", "proj-a", "--exclude", "proj-b"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content, err := fake.ReadFile("/home/fake/.claude-backup/config.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(content), "proj-a") || !strings.Contains(string(content), "proj-b") {
		t.Errorf("config = %q, want both excluded projects", content)
	}
}

func TestInstall_RerunningUpdatesConfigInPlace(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		return system.CommandResult{Stdout: "s3remote:\nnewremote:\n"}, nil
	}
	var stdout, stderr bytes.Buffer

	if err := cli.Run(fake, []string{"install", "--remote", "s3remote", "--exclude", "proj-a"}, &stdout, &stderr); err != nil {
		t.Fatalf("first install: %v", err)
	}
	if err := cli.Run(fake, []string{"install", "--remote", "newremote"}, &stdout, &stderr); err != nil {
		t.Fatalf("second install: %v", err)
	}

	writesToConfig := 0
	for _, w := range fake.Writes {
		if w.Path == "/home/fake/.claude-backup/config.json" {
			writesToConfig++
		}
	}
	if writesToConfig != 2 {
		t.Fatalf("got %d writes to config, want exactly 2 (no duplication)", writesToConfig)
	}

	content, err := fake.ReadFile("/home/fake/.claude-backup/config.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(content), "newremote") {
		t.Errorf("config = %q, want updated remote", content)
	}
	if strings.Contains(string(content), "proj-a") {
		t.Errorf("config = %q, want stale exclude list replaced by the re-run's flags", content)
	}
}

func TestInstall_RequiresRemoteFlag(t *testing.T) {
	fake := system.NewFake()
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"install"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("expected an error when --remote is missing")
	}
	if errors.Is(err, cli.ErrNotImplemented) {
		t.Fatal("install should be implemented now")
	}
}
