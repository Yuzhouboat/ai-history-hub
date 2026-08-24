package cli_test

import (
	"bytes"
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
