package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"claude-backup/internal/system"
)

func logPath(homeDir string) string {
	return filepath.Join(configDir(homeDir), "sync.log")
}

// appendLog appends a timestamped line to the sync log, creating it if it
// doesn't exist yet.
func appendLog(sys system.System, homeDir, line string) error {
	path := logPath(homeDir)
	existing, err := sys.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading log: %w", err)
	}
	entry := fmt.Sprintf("%s %s\n", time.Now().UTC().Format(time.RFC3339), line)
	return sys.WriteFile(path, append(existing, []byte(entry)...), 0o600)
}

// lastSyncDone scans the sync log for the most recent "sync done" entry and
// returns its timestamp. ok is false if the log doesn't exist yet or has no
// successful sync recorded.
func lastSyncDone(sys system.System, homeDir string) (time.Time, bool) {
	content, err := sys.ReadFile(logPath(homeDir))
	if err != nil {
		return time.Time{}, false
	}

	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if !strings.Contains(line, "sync done") {
			continue
		}
		fields := strings.SplitN(line, " ", 2)
		ts, err := time.Parse(time.RFC3339, fields[0])
		if err != nil {
			continue
		}
		return ts, true
	}
	return time.Time{}, false
}
