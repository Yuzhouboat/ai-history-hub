package cli

import (
	"errors"
	"os"
	"testing"

	"claude-backup/internal/system"
)

func TestSaveConfigThenLoadConfigRoundTrips(t *testing.T) {
	fake := system.NewFake()
	cfg := Config{Remote: "s3remote", Exclude: []string{"secret-project"}}

	if err := saveConfig(fake, "/home/fake", cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	got, err := loadConfig(fake, "/home/fake")
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if got.Remote != cfg.Remote {
		t.Errorf("remote = %q, want %q", got.Remote, cfg.Remote)
	}
	if len(got.Exclude) != 1 || got.Exclude[0] != "secret-project" {
		t.Errorf("exclude = %v, want [secret-project]", got.Exclude)
	}
}

func TestLoadConfigMissingReturnsNotExist(t *testing.T) {
	fake := system.NewFake()

	_, err := loadConfig(fake, "/home/fake")

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v, want wrapping os.ErrNotExist", err)
	}
}

func TestSaveConfigOverwritesPreviousConfig(t *testing.T) {
	fake := system.NewFake()

	if err := saveConfig(fake, "/home/fake", Config{Remote: "old", Exclude: []string{"a"}}); err != nil {
		t.Fatalf("first saveConfig: %v", err)
	}
	if err := saveConfig(fake, "/home/fake", Config{Remote: "new"}); err != nil {
		t.Fatalf("second saveConfig: %v", err)
	}

	got, err := loadConfig(fake, "/home/fake")
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if got.Remote != "new" {
		t.Errorf("remote = %q, want %q", got.Remote, "new")
	}
	if len(got.Exclude) != 0 {
		t.Errorf("exclude = %v, want empty after re-install without --exclude", got.Exclude)
	}
}
