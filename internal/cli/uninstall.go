package cli

import (
	"io"

	"claude-backup/internal/system"
)

// runUninstall will unload/remove the platform scheduler config (tickets
// 03/04). Not yet built.
func runUninstall(sys system.System, args []string, stdout, stderr io.Writer) error {
	return ErrNotImplemented
}
