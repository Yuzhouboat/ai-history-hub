package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"claude-backup/internal/cli"
	"claude-backup/internal/system"
)

func TestStatus_BeforeInstallProducesClearError(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"status"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("expected an error when status runs before install")
	}
	if !strings.Contains(err.Error(), "install") {
		t.Errorf("error = %q, want it to mention running install first", err.Error())
	}
}

func TestStatus_ReportsActiveScheduleAndNextRun(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.NowValue = time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		if name == "systemctl" && len(args) > 0 && args[len(args)-2] == "is-active" {
			return system.CommandResult{Stdout: "active\n"}, nil
		}
		return system.CommandResult{Stdout: "s3remote:\n"}, nil
	}
	var stdout, stderr bytes.Buffer
	if err := cli.Run(fake, []string{"install", "--remote", "s3remote"}, &stdout, &stderr); err != nil {
		t.Fatalf("install: %v", err)
	}
	stdout.Reset()

	err := cli.Run(fake, []string{"status"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "active") {
		t.Errorf("status output = %q, want it to report the schedule active", out)
	}
	if !strings.Contains(out, "2026-08-24T03:00:00Z") {
		t.Errorf("status output = %q, want it to report the next daily run", out)
	}
}

func TestStatus_ReportsLastSyncTimeFromLog(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		return system.CommandResult{Stdout: "s3remote:\n"}, nil
	}
	var stdout, stderr bytes.Buffer
	if err := cli.Run(fake, []string{"install", "--remote", "s3remote"}, &stdout, &stderr); err != nil {
		t.Fatalf("install: %v", err)
	}
	if err := cli.Run(fake, []string{"sync"}, &stdout, &stderr); err != nil {
		t.Fatalf("sync: %v", err)
	}
	stdout.Reset()

	err := cli.Run(fake, []string{"status"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if strings.Contains(out, "never") {
		t.Errorf("status output = %q, want a recorded last sync time after sync ran", out)
	}
}

func TestStatus_ReportsRemoteAndExcludedProjects(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		return system.CommandResult{Stdout: "s3remote:\n"}, nil
	}
	var stdout, stderr bytes.Buffer
	install := []string{"install", "--remote", "s3remote", "--exclude", "side-project"}
	if err := cli.Run(fake, install, &stdout, &stderr); err != nil {
		t.Fatalf("install: %v", err)
	}
	stdout.Reset()

	err := cli.Run(fake, []string{"status"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "remote: s3remote") {
		t.Errorf("status output = %q, want it to report the configured remote", out)
	}
	if !strings.Contains(out, "side-project") {
		t.Errorf("status output = %q, want it to report the excluded project", out)
	}
}

func TestStatus_ReportsExcludedProjectsNoneWhenListEmpty(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		return system.CommandResult{Stdout: "s3remote:\n"}, nil
	}
	var stdout, stderr bytes.Buffer
	if err := cli.Run(fake, []string{"install", "--remote", "s3remote"}, &stdout, &stderr); err != nil {
		t.Fatalf("install: %v", err)
	}
	stdout.Reset()

	err := cli.Run(fake, []string{"status"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "excluded projects: none") {
		t.Errorf("status output = %q, want it to report no excluded projects", stdout.String())
	}
}

func TestStatus_ReportsLastSyncFailureWithError(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		if name == "rclone" && len(args) > 0 && args[0] == "copy" {
			return system.CommandResult{ExitCode: 1}, errors.New("connection refused")
		}
		return system.CommandResult{Stdout: "s3remote:\n"}, nil
	}
	var stdout, stderr bytes.Buffer
	if err := cli.Run(fake, []string{"install", "--remote", "s3remote"}, &stdout, &stderr); err != nil {
		t.Fatalf("install: %v", err)
	}
	_ = cli.Run(fake, []string{"sync"}, &stdout, &stderr)
	stdout.Reset()

	err := cli.Run(fake, []string{"status"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "last sync:") || !strings.Contains(out, "(failed)") {
		t.Errorf("status output = %q, want the last sync flagged as failed", out)
	}
	if !strings.Contains(out, "last error: connection refused") {
		t.Errorf("status output = %q, want it to surface the failure reason", out)
	}
}

func TestStatus_ReportsNeverSyncedWhenNoSyncHasRunYet(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		return system.CommandResult{Stdout: "s3remote:\n"}, nil
	}
	var stdout, stderr bytes.Buffer
	if err := cli.Run(fake, []string{"install", "--remote", "s3remote"}, &stdout, &stderr); err != nil {
		t.Fatalf("install: %v", err)
	}
	stdout.Reset()

	err := cli.Run(fake, []string{"status"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "never") {
		t.Errorf("status output = %q, want it to report no sync has run yet", stdout.String())
	}
}
