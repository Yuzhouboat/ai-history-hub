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

// syncAttempt is the most recent completed sync run recorded in the log:
// either a "sync done" or a "sync failed: <reason>" entry. The "sync start"
// line that precedes every attempt is not itself an outcome and is skipped.
type syncAttempt struct {
	Time    time.Time
	Success bool
	Error   string
}

const syncFailedPrefix = "sync failed: "

// lastSyncAttempt scans the sync log for the most recent completed sync
// attempt (success or failure) and reports its outcome. ok is false if the
// log doesn't exist yet or no attempt has completed.
func lastSyncAttempt(sys system.System, homeDir string) (syncAttempt, bool) {
	content, err := sys.ReadFile(logPath(homeDir))
	if err != nil {
		return syncAttempt{}, false
	}

	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		fields := strings.SplitN(lines[i], " ", 2)
		if len(fields) != 2 {
			continue
		}
		ts, err := time.Parse(time.RFC3339, fields[0])
		if err != nil {
			continue
		}
		switch {
		case fields[1] == "sync done":
			return syncAttempt{Time: ts, Success: true}, true
		case strings.HasPrefix(fields[1], syncFailedPrefix):
			return syncAttempt{Time: ts, Success: false, Error: strings.TrimPrefix(fields[1], syncFailedPrefix)}, true
		}
	}
	return syncAttempt{}, false
}
