package cli

import (
	"io"

	"claude-backup/internal/system"
)

// runSync will read the persisted config and run rclone copy (ticket 02).
// Not yet built.
func runSync(sys system.System, args []string, stdout, stderr io.Writer) error {
	return ErrNotImplemented
}
