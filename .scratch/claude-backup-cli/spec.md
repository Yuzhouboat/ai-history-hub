Status: ready-for-agent

# claude-backup CLI

## Problem Statement

Claude Code keeps a user's full local chat history — session transcripts under `~/.claude/projects/<project>/<session-id>.jsonl` plus the global prompt index — on disk with no built-in export or offsite copy. Local transcripts older than 30 days are auto-deleted by default (`cleanupPeriodDays`), and even without that setting a laptop failure or accidental deletion permanently loses the history. The official web export requires manually triggering a browser session every time and only covers claude.ai chats, not Claude Code sessions. The user wants their Claude Code chat history protected without having to remember to run anything by hand.

## Solution

A single portable Go binary, `claude-backup`, that self-manages a daily OS-scheduled backup (additive-only) of the local chat history to an S3 bucket via rclone. It exposes four subcommands — `install`, `sync`, `status`, `uninstall` — so the user sets it up once and it runs unattended, using rclone's own credential handling rather than the tool owning AWS keys itself.

## User Stories

1. As a solo Claude Code user, I want to install a daily backup schedule with one command, so that I don't have to remember to back up my chat history manually.
2. As a solo Claude Code user, I want claude-backup to reuse my existing rclone remote if one already exists, so that I don't have to re-enter my S3 credentials.
3. As a solo Claude Code user setting up for the first time, I want claude-backup to walk me through creating an rclone remote if none exists, so that I can get backups running without leaving the CLI.
4. As a solo Claude Code user, I want claude-backup to fail fast with clear instructions if rclone isn't installed, so that I know exactly what to do before continuing.
5. As a solo Claude Code user, I want both my session transcripts and the global prompt index backed up, so that my full chat history (per the project's own definition of that term) is protected, not just part of it.
6. As a solo Claude Code user, I want backups to be additive-only, so that Claude Code's automatic 30-day cleanup of local transcripts never deletes my S3 copies.
7. As a solo Claude Code user, I want to exclude specific projects from backup, so that I can opt sensitive or irrelevant projects out.
8. As a solo Claude Code user, I want backups to run automatically every day even if my laptop is asleep at the scheduled time, so that sleep/shutdown doesn't cause me to silently miss backups.
9. As a solo Claude Code user, I want to trigger a sync manually at any time, so that I can back up immediately after an important session instead of waiting for the schedule.
10. As a solo Claude Code user, I want to check the status of my backup schedule, so that I can confirm it's actually running and see when it last succeeded.
11. As a solo Claude Code user, I want to see when the next scheduled backup will run, so that I can verify the schedule is configured correctly.
12. As a solo Claude Code user, I want to uninstall the backup schedule cleanly, so that I can stop backups without leaving stray scheduler config behind.
13. As a solo Claude Code user on macOS, I want the schedule installed via launchd, so that it integrates natively with how my Mac manages background jobs.
14. As a solo Claude Code user on Linux, I want the schedule installed via a systemd user timer, so that it integrates natively with how my Linux system manages background jobs.
15. As a solo Claude Code user on Linux, I want missed runs caught up automatically (e.g. my machine was off at the scheduled time), so that I don't accumulate long gaps between backups.
16. As a solo Claude Code user, I want claude-backup distributed as a single static binary, so that I don't need to install a language runtime to use it.
17. As a solo Claude Code user, I want claude-backup to never read, store, or transmit my AWS credentials itself, so that I can rely on rclone's mature, already-trusted credential handling instead of a new implementation.
18. As a security-conscious user, I want it clear that claude-backup relies on S3 server-side encryption and a private bucket rather than adding its own client-side encryption, so that I understand the actual threat model instead of assuming a guarantee that isn't there.
19. As a solo Claude Code user backing up from multiple machines, I want each machine's backups stored under its own hostname-scoped path in the bucket, so that backups from different machines don't collide or overwrite each other.
20. As a solo Claude Code user, I want sync activity logged with timestamps, so that I can troubleshoot a failed or missing backup after the fact.
21. As a solo Claude Code user, I want re-running `install` to update my existing schedule (e.g. a new remote or exclusion list) rather than fail or duplicate it, so that I can safely change my setup later.
22. As a solo Claude Code user, I want a clear error if `sync` runs before `install`, so that I understand the schedule isn't set up yet rather than seeing a confusing failure.
23. As a developer of claude-backup, I want the CLI's decision logic (what command to run, what config file content to write) testable without touching a real filesystem or a real OS scheduler, so that the test suite is fast and hermetic.
24. As a developer of claude-backup, I want both the macOS and Linux code paths covered by tests, so that a change to one platform's logic can't silently break the other undetected.

## Implementation Decisions

- **Language/distribution**: Go, compiled to a single static binary, cross-compiled for `GOOS=darwin` and `GOOS=linux` (per [ADR-0001](../../docs/adr/0001-go-and-rclone-with-rclone-owned-credentials.md)). Windows is not a target.
- **Transfer engine**: all data movement is delegated to `rclone`, invoked as an external subprocess. `claude-backup` requires `rclone` on `PATH` and does not auto-install it; `install` checks for it first and fails fast with install instructions if absent.
- **Credentials**: owned entirely by rclone's own remote config, never by `claude-backup` (ADR-0001). `install` takes a `--remote <name>` flag: if `rclone listremotes` shows it already exists, reuse it as-is; otherwise shell out to the interactive `rclone config` wizard to create it. `claude-backup` never reads, stores, or passes AWS keys.
- **Sync semantics**: use `rclone copy`, never `rclone sync` — additive-only, no deletion path against the remote, ever (ADR-0002).
- **Encryption**: none added by this tool. No `rclone crypt` overlay is offered by `install`; the tool relies entirely on S3 SSE + a private bucket (ADR-0003).
- **Backup source**: matches the "Chat history" glossary entry in `CONTEXT.md` — the `~/.claude/projects/` tree (session transcripts) plus the global prompt index file. Both are covered by one `rclone copy` invocation against `~/.claude/` scoped to those paths.
- **Excluded projects**: `install` accepts repeatable `--exclude <project-dir-name>` flags. The exclusion list is persisted in a local config file (e.g. `~/.claude-backup/config.json`) that `sync` reads at run time to build rclone's exclude-filter arguments. Default is no exclusions, matching the "Excluded project" glossary entry.
- **Destination path convention**: `<remote>:claude-code-backups/<hostname>/`, so multiple machines backing up to the same bucket never collide.
- **Scheduler — macOS**: a `launchd` agent, `.plist` written to `~/Library/LaunchAgents/`, using `StartCalendarInterval` for a daily trigger.
- **Scheduler — Linux**: a `systemd` user `.service` + `.timer` pair written to `~/.config/systemd/user/`, using `OnCalendar` for the daily trigger and `Persistent=true` so a missed run (machine off/asleep at the scheduled time) fires on next boot/login instead of silently being skipped.
- **Subcommands**:
  - `install [--remote <name>] [--exclude <project> ...]` — validates rclone is present, resolves/creates the remote, writes the platform-appropriate scheduler config, loads it, and persists the exclusion config. Idempotent: re-running with different flags updates the existing schedule/config rather than erroring or duplicating it.
  - `sync` — runs one immediate `rclone copy` using the persisted config (remote + exclusions); this is the same command the schedule invokes. Errors clearly if no config exists yet (i.e. `install` was never run). Appends timestamped start/done lines to a log file.
  - `status` — reports whether the schedule is currently loaded/active, the last recorded sync time, and the next scheduled run time, by querying the platform scheduler (`launchctl` / `systemctl --user`) and/or reading the sync log.
  - `uninstall` — unloads/stops and disables the schedule and removes the generated scheduler config files. Does not touch already-uploaded S3 data (consistent with the additive-only design) and does not delete the local exclusion config, so a later `install` can pick the same exclusions back up.
- **Testing seam**: a single injected `System` interface abstracts every point of contact with the outside world — running an external command (rclone, launchctl/systemctl) and reading/writing a file. All four subcommands depend only on this interface for their side effects; everything else (deciding what plist/unit content to generate per `GOOS`, what rclone arguments to build, how to interpret scheduler/log output for `status`) is pure orchestration logic layered on top of it. A real implementation backs `System` with `os/exec` and `os`; tests use an in-memory fake.

## Testing Decisions

- Tests exercise subcommand orchestration (`install`, `sync`, `status`, `uninstall`) entirely through the fake `System`, asserting on the exact external commands (name + args) and exact file writes (path + content) each subcommand produces — not on internal Go call sequences beyond what's observable through that one interface. This is what "only test external behavior, not implementation details" means for this tool: `System` *is* the external behavior boundary.
- Both platform branches are first-class, not implementation detail: cover macOS (`launchd` plist generation + `launchctl` invocations) and Linux (`systemd` unit/timer generation + `systemctl` invocations) as separate cases for `install`, `status`, and `uninstall`.
- Cover the error paths explicitly: rclone missing from `PATH`, remote not found (triggers the `rclone config` wizard path), a scheduler command failing and surfacing as a clear CLI error, `sync` run before `install`.
- No prior art in this repo — it's greenfield. Build the first suite test-first per `/tdd`, one red-green slice per subcommand behavior, using the fake `System` from the first test onward rather than retrofitting it later.
- Real `rclone`/`launchctl`/`systemd` are never invoked by the automated suite. A manual check against a real test bucket is useful before first real-world use but is not part of the automated test suite.

## Out of Scope

- Auto-installing rclone if it's missing from `PATH` (ADR-0001) — fail fast with instructions instead.
- Client-side encryption / an `rclone crypt` overlay (ADR-0003).
- Remote pruning, retention policies, or any deletion path against S3 (ADR-0002) — storage is expected to grow monotonically by design.
- Windows support.
- Distribution packaging (Homebrew tap, GitHub Releases automation) — a build/release-time concern, not behavior of the CLI itself.
- A real-time `SessionEnd`-hook-based capture path — explicitly considered and rejected in favor of the scheduled daily resync (see `claude-history-backup-summary.md`), since `Ctrl+C` doesn't trigger `SessionEnd` and hook-only capture can miss sessions.
- Restoring from backup (a pull-down/restore command) — this spec covers push-only backup; nothing so far has asked for restore.

## Further Notes

- Background context and the options that were considered and rejected along the way live in `claude-history-backup-summary.md` at the repo root.
- The "Chat history", "Session transcript", "Backup", "Remote", and "Excluded project" terms used throughout this spec are defined in `CONTEXT.md`; use them as defined there in code, tests, and CLI help text rather than drifting to synonyms (e.g. never "sync" alone when the additive-only guarantee matters).
