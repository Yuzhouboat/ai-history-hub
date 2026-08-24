package cli

import (
	"io"

	"claude-backup/internal/system"
)

// runInstall will validate rclone, resolve/create the remote, and persist
// the config (ticket 02) and write/load the platform scheduler (tickets
// 03/04). Not yet built.
func runInstall(sys system.System, args []string, stdout, stderr io.Writer) error {
	return ErrNotImplemented
}
