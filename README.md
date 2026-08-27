# aiHistoryHub

[![release](https://img.shields.io/github/v/release/Yuzhouboat/ai-history-hub)](https://github.com/Yuzhouboat/ai-history-hub/releases)
[![license](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Tooling for preserving and backing up local AI assistant history. The first tool here is `claude-backup`, a CLI that ships Claude Code's local chat history to S3 via [rclone](https://rclone.org/), on a daily schedule it manages for you.

`claude-backup` protects the chat history Claude Code keeps under `~/.claude`: every session transcript plus the global prompt index. Backups are additive-only — files are copied to S3, never deleted from it, even after they're gone locally.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/Yuzhouboat/ai-history-hub/master/install.sh | sh
```

This downloads the latest prebuilt `claude-backup` binary for your OS/arch (linux/darwin, amd64/arm64) from [GitHub Releases](https://github.com/Yuzhouboat/ai-history-hub/releases) — no Go toolchain required. It also installs `rclone` alongside it if `rclone` isn't already on your `PATH`, since `claude-backup` shells out to it for the actual transfer.

Building from source instead:

```sh
git clone https://github.com/Yuzhouboat/ai-history-hub.git
cd ai-history-hub
go build -o claude-backup ./cmd/claude-backup
```

## Usage

```sh
# One-time setup: point claude-backup at an rclone remote (created if it
# doesn't exist yet) and turn on the daily schedule.
claude-backup install --remote s3remote

# Optionally, exclude specific projects from every backup (repeatable):
claude-backup install --remote s3remote --exclude side-project

# Run a backup right now, outside the daily schedule.
claude-backup sync

# Check what's configured, whether the schedule is active, and how the
# last sync went.
claude-backup status

# Turn the daily schedule off. S3 data and local config are left alone.
claude-backup uninstall
```

Run `claude-backup --help` (or `claude-backup <command> -h`) for full usage and flags.

### Requirements

- **rclone**, on `PATH` — `claude-backup` shells out to it for the S3 transfer and owns none of the underlying credentials itself. `install.sh` installs it for you; otherwise see [rclone's install docs](https://rclone.org/install/).
- **An S3-compatible rclone remote** — `claude-backup install --remote <name>` reuses one if it already exists in your rclone config, or launches `rclone config` to create it.
- **Linux (systemd) or macOS (launchd)** for the self-managed daily schedule. Other platforms can still run `claude-backup sync` manually or from their own scheduler.

### Troubleshooting

- **`rclone not found on PATH`** — install rclone (see Requirements above) and re-run `claude-backup install`.
- **A sync failed** — `claude-backup status` reports the last error alongside the last sync time.
- **Schedule shows `not active`** — re-run `claude-backup install` to regenerate and reactivate the platform schedule.

## Design

See `CONTEXT.md` for the domain vocabulary used throughout this project, and `docs/adr/` for the reasoning behind key decisions (rclone-owned credentials, additive-only backup, no client-side encryption).

## License

[MIT](LICENSE)
