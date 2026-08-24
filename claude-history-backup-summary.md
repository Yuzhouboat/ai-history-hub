# Claude History Export & Backup — Discussion Summary

## 1. claude.ai (web chat) export
- Official export: **Settings → Privacy → Export Data** (web or Claude Desktop only, not mobile).
- Anthropic emails a download link, valid ~24 hours; delivers a ZIP of JSON conversation data.
- Available on Free, Pro, and Max plans. Deleted conversations aren't included.
- **No public API** to trigger or automate this export — it requires an authenticated browser session.
- Semi-automation is possible only for the *download* step (e.g., an email filter + script that grabs the link), but the export must still be manually triggered each time.
- Third-party browser extensions exist (e.g., Claude Exporter, AI Toolbox) that scrape claude.ai's internal endpoints — unofficial, use at your own risk regarding account access.

## 2. Team vs Enterprise (org-level export options)
| Feature | Team | Enterprise |
|---|---|---|
| Manual data export | Only Primary Owner can trigger | Yes |
| Compliance API (programmatic chat/file access) | ❌ No | ✅ Yes |
| Audit logs | ❌ No | ✅ Yes |
| Minimum seats | 2 (self-serve) | 20 (self-serve), 50 (sales-assisted) |
| Pricing | ~$20–25/seat/mo (Standard), ~$100–150/seat/mo (Premium) | ~$20/seat/mo + all usage billed separately at API rates |

- Enterprise is **not realistically accessible for a solo individual** — 20-seat minimum means paying for 19 unused seats plus metered usage on top, with no included token allowance.
- For a heavy solo user, **Claude Max** ($100–$200/mo flat) is the more sensible option vs. Enterprise.

## 3. Claude Code local session history
- Sessions stored locally by default, no account/export step needed:
  - Per-session transcripts: `~/.claude/projects/<project>/<session-id>.jsonl`
  - Global prompt index: `~/.claude/history.jsonl`
- Built-in tools:
  - `/export` — manual, per-session, saves to clipboard or a text file
  - `/history` — lists recent sessions
  - `claude -r <session-id>` / `claude -c` — resume a session
  - `claude -p --output-format json` — scriptable, stable structured output (safer than parsing raw JSONL, whose format can change between releases)
- **Default auto-deletion**: transcripts older than 30 days are deleted unless you raise `cleanupPeriodDays` in `~/.claude/settings.json`.
- Third-party tool mentioned: `claude-conversation-extractor` (`claude-extract --all`) for batch Markdown/JSON/HTML export.

## 4. Automating Claude Code backups
Two complementary layers recommended:

### a) Event-driven: `SessionEnd` hook
- Claude Code hooks fire on lifecycle events (`SessionStart`, `SessionEnd`, `Stop`, `PreToolUse`, etc.), configured in `~/.claude/settings.json`.
- A `SessionEnd` hook receives JSON on stdin (including `transcript_path`) and can immediately push that session's transcript to a remote destination (S3, rclone remote, custom server, etc.) right when a session ends.
- **Gotcha**: `Ctrl+C` only *suspends* a session — it does **not** trigger `SessionEnd`. Only `/exit` or closing the terminal does. So hook-only capture can miss sessions.

### b) Scheduled: daily resync (chosen approach)
- Decided against real-time hooks alone; going with a **daily cron/scheduled sync** of the whole `~/.claude/projects/` directory as a catch-all, since it's simpler and more robust than per-session hooks.
- Recommended tool: `rclone sync` (supports 40+ cloud backends: Drive, Dropbox, S3, etc.).
- Example script:
  ```bash
  #!/bin/bash
  # ~/.claude-backup/sync.sh
  LOGFILE="$HOME/.claude-backup/sync.log"
  echo "$(date): starting sync" >> "$LOGFILE"
  rclone sync "$HOME/.claude/projects/" "remote:claude-code-backups/$(hostname)/" \
    --log-file="$LOGFILE" --log-level INFO
  echo "$(date): done" >> "$LOGFILE"
  ```
- Native OS scheduling recommended over hand-rolled daemons:
  - **macOS**: `launchd` — `.plist` in `~/Library/LaunchAgents/`, using `StartCalendarInterval`
  - **Linux**: `systemd` — user `.service` + `.timer` in `~/.config/systemd/user/`, using `OnCalendar` + `Persistent=true` (catches up missed runs, e.g. laptop asleep at scheduled time)

## 5. From script → portable background tool
Explored three levels of packaging, landed on the simplest that fits:

1. **Script + native scheduler (recommended baseline)** — the sync script above, installed via `launchd`/`systemd`, distributed via a small git repo + one-line installer script. No app framework needed.
2. **Full native app (`.app` bundle, Swift/SwiftUI menu bar agent)** — considered but deemed overkill for a daily file-sync job; real engineering overhead (Xcode, code signing, Gatekeeper) for something that's fundamentally a scheduled `rsync`.
3. **CLI tool (final direction)** — a single compiled binary (Go recommended) exposing subcommands:
   ```bash
   claude-backup install    # writes + loads launchd/systemd config
   claude-backup sync       # runs one sync manually
   claude-backup status     # shows last/next run
   claude-backup uninstall  # removes the schedule
   ```
   - Go chosen for: single static binary, no runtime dependencies, trivial cross-compilation (`GOOS=darwin`/`GOOS=linux`) from one codebase.
   - `install` subcommand self-manages writing/loading the launchd plist (Mac) or systemd timer (Linux) — no manual config editing.
   - Distribution options ranked by effort: Homebrew tap > GitHub Releases binary > `go install`.
   - This approach gives real portability across Mac/Linux without app-signing or packaging overhead.

## Next step (not yet done)
Scaffold the actual Go CLI project — `main.go`, launchd/systemd template generation, and a basic README — was offered but not yet built.
