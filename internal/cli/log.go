package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
