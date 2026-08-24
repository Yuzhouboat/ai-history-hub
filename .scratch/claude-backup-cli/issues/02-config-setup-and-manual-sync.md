# 02: Config setup + manual sync

**What to build:** A user can run `claude-backup install --remote <name> [--exclude <project> ...]` to set up their rclone remote and exclusion list, then run `claude-backup sync` at any time to manually back up their full Claude Code chat history (session transcripts plus the global prompt index) to S3, additive-only, with logged activity. No OS scheduler yet — this ticket makes "back up right now" fully real end-to-end.

**Blocked by:** 01

**Status:** done

- [x] `install` fails fast with clear instructions if `rclone` is not found on `PATH`
- [x] `install --remote <name>` reuses the remote as-is if `rclone listremotes` already shows it
- [x] `install --remote <name>` launches the interactive `rclone config` wizard to create the remote if it doesn't already exist
- [x] `install` never reads, stores, or passes AWS credentials itself — remote creation is entirely delegated to rclone
- [x] `install --exclude <project>` (repeatable) persists an exclusion list to a local config file, alongside the resolved remote name
- [x] Re-running `install` with different flags updates the existing config in place rather than erroring or duplicating it
- [x] `sync` reads the persisted config (remote + exclusions) and runs `rclone copy` (never `rclone sync`) covering both the `~/.claude/projects/` tree and the global prompt index file, targeting `<remote>:claude-code-backups/<hostname>/`
- [x] `sync` applies the persisted exclusion list as rclone exclude-filter arguments
- [x] `sync` run before `install` has ever completed produces a clear, distinct error rather than a confusing failure
- [x] `sync` appends timestamped start and done lines to a log file
- [x] All of the above is verified through the fake `System` from ticket 01 — no real filesystem, rclone, or network access in tests

## Comments

Implemented via TDD. Notable seam decision beyond the ticket's literal scope: extended `System` with `RunInteractive` (stdin/stdout/stderr passthrough, needed for the real `rclone config` wizard to actually be interactive — `Run` only captures output) and `Hostname`/`UserHomeDir` (so the hostname-scoped destination path and config/log locations stay hermetically testable through the fake rather than hitting the real OS in tests). Config persists as JSON at `~/.claude-backup/config.json`; sync log at `~/.claude-backup/sync.log`. `rclone copy` scopes to chat history via filters (`+ /projects/**`, `+ /history.jsonl`, `- *`), with per-project excludes placed before the general include so first-match-wins filter semantics apply correctly.
