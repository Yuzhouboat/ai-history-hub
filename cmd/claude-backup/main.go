// Command claude-backup backs up Claude Code's local chat history to S3.
package main

import (
	"fmt"
	"os"

	"claude-backup/internal/cli"
	"claude-backup/internal/system"
)

func main() {
	sys := system.NewReal()
	if err := cli.Run(sys, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
