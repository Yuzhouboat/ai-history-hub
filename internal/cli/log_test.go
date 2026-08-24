package cli

import (
	"strings"
	"testing"

	"claude-backup/internal/system"
)

func TestAppendLogAppendsAcrossCalls(t *testing.T) {
	fake := system.NewFake()

	if err := appendLog(fake, "/home/fake", "sync start"); err != nil {
		t.Fatalf("first appendLog: %v", err)
	}
	if err := appendLog(fake, "/home/fake", "sync done"); err != nil {
		t.Fatalf("second appendLog: %v", err)
	}

	content, err := fake.ReadFile(logPath("/home/fake"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %q", len(lines), content)
	}
	if !strings.Contains(lines[0], "sync start") {
		t.Errorf("first line = %q, want it to contain %q", lines[0], "sync start")
	}
	if !strings.Contains(lines[1], "sync done") {
		t.Errorf("second line = %q, want it to contain %q", lines[1], "sync done")
	}
}
