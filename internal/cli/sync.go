package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"claude-backup/internal/system"
)

// runSync backs up the local chat history (the ~/.claude/projects tree and
// the global prompt index) to S3 using the remote and exclusion list
// persisted by install, additive-only (rclone copy, never rclone sync).
func runSync(sys system.System, args []string, stdout, stderr io.Writer) error {
	homeDir, err := resolveHomeDir(sys, "sync")
	if err != nil {
		return err
	}

	cfg, err := loadConfig(sys, homeDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("sync: no backup configured yet; run `claude-backup install` first")
		}
		return fmt.Errorf("sync: loading config: %w", err)
	}

	hostname, err := sys.Hostname()
	if err != nil {
		return fmt.Errorf("sync: resolving hostname: %w", err)
	}

	if err := appendLog(sys, homeDir, "sync start"); err != nil {
		return fmt.Errorf("sync: logging start: %w", err)
	}

	dest := fmt.Sprintf("%s:claude-code-backups/%s/", cfg.Remote, hostname)
	rcloneArgs := rcloneCopyArgs(homeDir, dest, cfg.Exclude)

	result, err := sys.Run("rclone", rcloneArgs...)
	if err != nil {
		_ = appendLog(sys, homeDir, "sync failed: "+err.Error())
		return fmt.Errorf("sync: rclone copy failed: %w (%s)", err, result.Stderr)
	}

	if err := appendLog(sys, homeDir, "sync done"); err != nil {
		return fmt.Errorf("sync: logging done: %w", err)
	}

	fmt.Fprintln(stdout, "sync complete")
	return nil
}

// rcloneCopyArgs builds the `rclone copy` arguments scoping the copy to
// exactly the chat history: the projects tree and the global prompt index,
// with per-project exclude filters applied before the general includes so
// they take precedence (rclone filters match first-rule-wins, top to
// bottom).
func rcloneCopyArgs(homeDir, dest string, excludedProjects []string) []string {
	args := []string{"copy", filepath.Join(homeDir, ".claude"), dest}
	for _, project := range excludedProjects {
		args = append(args, "--filter", "- /projects/"+project+"/**")
	}
	args = append(args,
		"--filter", "+ /projects/**",
		"--filter", "+ /history.jsonl",
		"--filter", "- *",
	)
	return args
}
