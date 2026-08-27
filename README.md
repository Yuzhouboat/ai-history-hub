# aiHistoryHub

Tooling for preserving and backing up local AI assistant history. The first tool here is `claude-backup`, a CLI that periodically ships Claude Code's local chat history to S3 via rclone.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/Yuzhouboat/ai-history-hub/master/install.sh | sh
```

This downloads the latest prebuilt `claude-backup` binary for your OS/arch (linux/darwin, amd64/arm64) from [GitHub Releases](https://github.com/Yuzhouboat/ai-history-hub/releases) — no Go toolchain required.

## Usage

```sh
claude-backup --help
```

See `CONTEXT.md` for the domain vocabulary used throughout this project.
