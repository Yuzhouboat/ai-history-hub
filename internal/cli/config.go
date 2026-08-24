package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"claude-backup/internal/system"
)

// Config is claude-backup's persisted setup: the rclone remote to use and
// the projects excluded from backup. install writes it; sync reads it.
type Config struct {
	Remote  string   `json:"remote"`
	Exclude []string `json:"exclude,omitempty"`
}

// configDir is where claude-backup keeps everything it persists: the
// config file and the sync log.
func configDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude-backup")
}

func configPath(homeDir string) string {
	return filepath.Join(configDir(homeDir), "config.json")
}

// resolveHomeDir wraps sys.UserHomeDir with a subcommand-scoped error
// message, since both install and sync need it to locate the config dir.
func resolveHomeDir(sys system.System, cmd string) (string, error) {
	homeDir, err := sys.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%s: resolving home directory: %w", cmd, err)
	}
	return homeDir, nil
}

func saveConfig(sys system.System, homeDir string, cfg Config) error {
	content, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	return sys.WriteFile(configPath(homeDir), content, 0o600)
}

func loadConfig(sys system.System, homeDir string) (Config, error) {
	content, err := sys.ReadFile(configPath(homeDir))
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}
