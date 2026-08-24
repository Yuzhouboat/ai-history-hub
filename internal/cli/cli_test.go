package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"claude-backup/internal/cli"
	"claude-backup/internal/system"
)

func TestRun_NoArgsPrintsUsage(t *testing.T) {
	fake := system.NewFake()
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, nil, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "claude-backup") {
		t.Errorf("usage output = %q, want it to mention claude-backup", stdout.String())
	}
	if len(fake.Commands) != 0 {
		t.Errorf("expected no commands run for bare usage, got %v", fake.Commands)
	}
}

func TestRun_HelpFlagPrintsUsage(t *testing.T) {
	fake := system.NewFake()
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"--help"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("usage output = %q, want it to contain Usage:", stdout.String())
	}
}

func TestRun_UnknownSubcommandErrors(t *testing.T) {
	fake := system.NewFake()
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"bogus"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("expected an error for an unknown subcommand")
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("stderr = %q, want usage on unknown subcommand", stderr.String())
	}
}
