package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"claude-backup/internal/cli"
	"claude-backup/internal/system"
)

func installedFake(t *testing.T, remote string, exclude []string) *system.Fake {
	t.Helper()
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	fake.HostnameValue = "test-host"
	content := `{"remote":"` + remote + `"`
	if len(exclude) > 0 {
		content += `,"exclude":["` + strings.Join(exclude, `","`) + `"]`
	}
	content += `}`
	if err := fake.WriteFile("/home/fake/.claude-backup/config.json", []byte(content), 0o600); err != nil {
		t.Fatalf("seeding config: %v", err)
	}
	return fake
}

func TestSync_BeforeInstallProducesClearError(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/fake"
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"sync"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("expected an error when sync runs before install")
	}
	if !strings.Contains(err.Error(), "install") {
		t.Errorf("error = %q, want it to mention running install first", err.Error())
	}
	for _, c := range fake.Commands {
		if c.Name == "rclone" {
			t.Errorf("expected no rclone invocation before install, got %+v", c)
		}
	}
}

func TestSync_RunsRcloneCopyNeverSync(t *testing.T) {
	fake := installedFake(t, "s3remote", nil)
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"sync"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.Commands) == 0 {
		t.Fatal("expected an rclone command to run")
	}
	cmd := fake.Commands[0]
	if cmd.Name != "rclone" {
		t.Fatalf("command = %+v, want rclone", cmd)
	}
	if len(cmd.Args) == 0 || cmd.Args[0] != "copy" {
		t.Fatalf("args = %v, want first arg 'copy' (never 'sync')", cmd.Args)
	}
	for _, a := range cmd.Args {
		if a == "sync" {
			t.Fatalf("args = %v, must never invoke rclone sync (additive-only)", cmd.Args)
		}
	}
}

func TestSync_TargetsHostnameScopedRemotePath(t *testing.T) {
	fake := installedFake(t, "s3remote", nil)
	var stdout, stderr bytes.Buffer

	if err := cli.Run(fake, []string{"sync"}, &stdout, &stderr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := fake.Commands[0]
	wantDest := "s3remote:claude-code-backups/test-host/"
	found := false
	for _, a := range cmd.Args {
		if a == wantDest {
			found = true
		}
	}
	if !found {
		t.Errorf("args = %v, want it to include destination %q", cmd.Args, wantDest)
	}
}

func TestSync_CoversProjectsTreeAndPromptIndex(t *testing.T) {
	fake := installedFake(t, "s3remote", nil)
	var stdout, stderr bytes.Buffer

	if err := cli.Run(fake, []string{"sync"}, &stdout, &stderr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := fake.Commands[0]
	joined := strings.Join(cmd.Args, " ")
	if !strings.Contains(cmd.Args[1], "/home/fake/.claude") {
		t.Errorf("source arg = %q, want it scoped under ~/.claude", cmd.Args[1])
	}
	if !strings.Contains(joined, "/projects/**") {
		t.Errorf("args = %q, want a filter including the projects tree", joined)
	}
	if !strings.Contains(joined, "/history.jsonl") {
		t.Errorf("args = %q, want a filter including the global prompt index", joined)
	}
}

func TestSync_AppliesExclusionListAsFilters(t *testing.T) {
	fake := installedFake(t, "s3remote", []string{"secret-project"})
	var stdout, stderr bytes.Buffer

	if err := cli.Run(fake, []string{"sync"}, &stdout, &stderr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := fake.Commands[0]
	joined := strings.Join(cmd.Args, " ")
	if !strings.Contains(joined, "secret-project") {
		t.Errorf("args = %q, want an exclude filter for secret-project", joined)
	}

	excludeIdx, includeIdx := -1, -1
	for i, a := range cmd.Args {
		if strings.Contains(a, "secret-project") {
			excludeIdx = i
		}
		if a == "+ /projects/**" {
			includeIdx = i
		}
	}
	if excludeIdx == -1 || includeIdx == -1 || excludeIdx > includeIdx {
		t.Errorf("args = %v, want the project-specific exclude filter before the general projects include", cmd.Args)
	}
}

func TestSync_AppendsTimestampedStartAndDoneLogLines(t *testing.T) {
	fake := installedFake(t, "s3remote", nil)
	var stdout, stderr bytes.Buffer

	if err := cli.Run(fake, []string{"sync"}, &stdout, &stderr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := fake.ReadFile("/home/fake/.claude-backup/sync.log")
	if err != nil {
		t.Fatalf("ReadFile log: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d log lines, want 2 (start + done): %q", len(lines), content)
	}
	if !strings.Contains(lines[0], "start") {
		t.Errorf("first log line = %q, want it to mention start", lines[0])
	}
	if !strings.Contains(lines[1], "done") {
		t.Errorf("second log line = %q, want it to mention done", lines[1])
	}
}

func TestSync_RcloneFailureIsSurfacedAsError(t *testing.T) {
	fake := installedFake(t, "s3remote", nil)
	wantErr := errors.New("connection refused")
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		return system.CommandResult{ExitCode: 1, Stderr: "connection refused"}, wantErr
	}
	var stdout, stderr bytes.Buffer

	err := cli.Run(fake, []string{"sync"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("expected an error when rclone copy fails")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want it to wrap %v", err, wantErr)
	}
}
