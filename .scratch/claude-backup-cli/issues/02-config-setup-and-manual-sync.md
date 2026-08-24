# 02: Config setup + manual sync

**What to build:** A user can run `claude-backup install --remote <name> [--exclude <project> ...]` to set up their rclone remote and exclusion list, then run `claude-backup sync` at any time to manually back up their full Claude Code chat history (session transcripts plus the global prompt index) to S3, additive-only, with logged activity. No OS scheduler yet — this ticket makes "back up right now" fully real end-to-end.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] `install` fails fast with clear instructions if `rclone` is not found on `PATH`
- [ ] `install --remote <name>` reuses the remote as-is if `rclone listremotes` already shows it
- [ ] `install --remote <name>` launches the interactive `rclone config` wizard to create the remote if it doesn't already exist
- [ ] `install` never reads, stores, or passes AWS credentials itself — remote creation is entirely delegated to rclone
- [ ] `install --exclude <project>` (repeatable) persists an exclusion list to a local config file, alongside the resolved remote name
- [ ] Re-running `install` with different flags updates the existing config in place rather than erroring or duplicating it
- [ ] `sync` reads the persisted config (remote + exclusions) and runs `rclone copy` (never `rclone sync`) covering both the `~/.claude/projects/` tree and the global prompt index file, targeting `<remote>:claude-code-backups/<hostname>/`
- [ ] `sync` applies the persisted exclusion list as rclone exclude-filter arguments
- [ ] `sync` run before `install` has ever completed produces a clear, distinct error rather than a confusing failure
- [ ] `sync` appends timestamped start and done lines to a log file
- [ ] All of the above is verified through the fake `System` from ticket 01 — no real filesystem, rclone, or network access in tests
