package system_test

import (
	"errors"
	"os"
	"testing"

	"claude-backup/internal/system"
)

func TestFake_RunRecordsCommands(t *testing.T) {
	fake := system.NewFake()

	if _, err := fake.Run("rclone", "copy", "src", "dst"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := fake.Run("launchctl", "load", "agent.plist"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fake.Commands) != 2 {
		t.Fatalf("got %d recorded commands, want 2", len(fake.Commands))
	}
	if fake.Commands[0].Name != "rclone" || fake.Commands[0].Args[0] != "copy" {
		t.Errorf("first command = %+v, want rclone copy ...", fake.Commands[0])
	}
	if fake.Commands[1].Name != "launchctl" {
		t.Errorf("second command = %+v, want launchctl ...", fake.Commands[1])
	}
}

func TestFake_RunFuncControlsResult(t *testing.T) {
	fake := system.NewFake()
	wantErr := errors.New("rclone not found")
	fake.RunFunc = func(name string, args ...string) (system.CommandResult, error) {
		return system.CommandResult{ExitCode: 127}, wantErr
	}

	result, err := fake.Run("rclone", "listremotes")

	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if result.ExitCode != 127 {
		t.Errorf("exit code = %d, want 127", result.ExitCode)
	}
	if len(fake.Commands) != 1 {
		t.Errorf("expected command still recorded despite RunFunc, got %d", len(fake.Commands))
	}
}

func TestFake_WriteFileThenReadFileRoundTrips(t *testing.T) {
	fake := system.NewFake()

	if err := fake.WriteFile("/config/claude-backup.json", []byte(`{"remote":"s3"}`), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := fake.ReadFile("/config/claude-backup.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != `{"remote":"s3"}` {
		t.Errorf("content = %q, want %q", got, `{"remote":"s3"}`)
	}
}

func TestFake_WriteFileRecordsEachWriteEvenToTheSamePath(t *testing.T) {
	fake := system.NewFake()

	if err := fake.WriteFile("/config/claude-backup.json", []byte(`{"remote":"a"}`), 0o644); err != nil {
		t.Fatalf("first WriteFile: %v", err)
	}
	if err := fake.WriteFile("/config/claude-backup.json", []byte(`{"remote":"b"}`), 0o644); err != nil {
		t.Fatalf("second WriteFile: %v", err)
	}

	if len(fake.Writes) != 2 {
		t.Fatalf("got %d recorded writes, want 2", len(fake.Writes))
	}
	if fake.Writes[0].Path != "/config/claude-backup.json" || string(fake.Writes[0].Content) != `{"remote":"a"}` {
		t.Errorf("first write = %+v, want path+content for remote a", fake.Writes[0])
	}
	if fake.Writes[1].Path != "/config/claude-backup.json" || string(fake.Writes[1].Content) != `{"remote":"b"}` {
		t.Errorf("second write = %+v, want path+content for remote b", fake.Writes[1])
	}

	got, err := fake.ReadFile("/config/claude-backup.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != `{"remote":"b"}` {
		t.Errorf("current content = %q, want the latest write to win", got)
	}
}

func TestFake_ReadFileMissingReturnsNotExist(t *testing.T) {
	fake := system.NewFake()

	_, err := fake.ReadFile("/does/not/exist.json")

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v, want wrapping os.ErrNotExist", err)
	}
}

func TestFake_RunInteractiveRecordsCommand(t *testing.T) {
	fake := system.NewFake()

	if err := fake.RunInteractive("rclone", "config"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fake.InteractiveCommands) != 1 {
		t.Fatalf("got %d recorded interactive commands, want 1", len(fake.InteractiveCommands))
	}
	if fake.InteractiveCommands[0].Name != "rclone" || fake.InteractiveCommands[0].Args[0] != "config" {
		t.Errorf("recorded command = %+v, want rclone config", fake.InteractiveCommands[0])
	}
}

func TestFake_RunInteractiveFuncControlsResult(t *testing.T) {
	fake := system.NewFake()
	wantErr := errors.New("wizard cancelled")
	fake.RunInteractiveFunc = func(name string, args ...string) error {
		return wantErr
	}

	err := fake.RunInteractive("rclone", "config")

	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if len(fake.InteractiveCommands) != 1 {
		t.Errorf("expected command still recorded despite RunInteractiveFunc, got %d", len(fake.InteractiveCommands))
	}
}

func TestFake_HostnameDefault(t *testing.T) {
	fake := system.NewFake()

	got, err := fake.Hostname()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("expected a non-empty default hostname")
	}
}

func TestFake_HostnameOverride(t *testing.T) {
	fake := system.NewFake()
	fake.HostnameValue = "my-laptop"

	got, err := fake.Hostname()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "my-laptop" {
		t.Errorf("hostname = %q, want %q", got, "my-laptop")
	}
}

func TestFake_UserHomeDirDefault(t *testing.T) {
	fake := system.NewFake()

	got, err := fake.UserHomeDir()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("expected a non-empty default home dir")
	}
}

func TestFake_UserHomeDirOverride(t *testing.T) {
	fake := system.NewFake()
	fake.HomeDirValue = "/home/someone"

	got, err := fake.UserHomeDir()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/home/someone" {
		t.Errorf("home dir = %q, want %q", got, "/home/someone")
	}
}

func TestFake_FileExists(t *testing.T) {
	fake := system.NewFake()

	exists, err := fake.FileExists("/config/claude-backup.json")
	if err != nil {
		t.Fatalf("FileExists (absent): %v", err)
	}
	if exists {
		t.Errorf("expected false before write")
	}

	if err := fake.WriteFile("/config/claude-backup.json", []byte("{}"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	exists, err = fake.FileExists("/config/claude-backup.json")
	if err != nil {
		t.Fatalf("FileExists (present): %v", err)
	}
	if !exists {
		t.Errorf("expected true after write")
	}
}
