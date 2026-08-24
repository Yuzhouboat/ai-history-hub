package cli

import (
	"io"

	"claude-backup/internal/system"
)

// runStatus will report schedule state, last sync, and next run (ticket
// 05). Not yet built.
func runStatus(sys system.System, args []string, stdout, stderr io.Writer) error {
	return ErrNotImplemented
}
