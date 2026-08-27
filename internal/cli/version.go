package cli

import (
	"fmt"
	"io"
)

// Version is claude-backup's released version. It is overridden at build
// time via -ldflags "-X claude-backup/internal/cli.Version=..." (wired up
// in .goreleaser.yaml); local `go build`/`go run` without that flag keeps
// the "dev" default.
var Version = "dev"

func runVersion(stdout io.Writer) {
	fmt.Fprintf(stdout, "claude-backup %s\n", Version)
}
