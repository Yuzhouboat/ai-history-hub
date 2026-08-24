# 05: status command

**What to build:** A user can run `claude-backup status` to confirm their backup schedule is actually active and see when it last succeeded and when it will next run, on whichever platform (macOS or Linux) they're on.

**Blocked by:** 03, 04

**Status:** done

- [x] `status` reports whether the schedule is currently loaded/active, querying the platform scheduler (`launchctl` on darwin, `systemctl --user` on linux)
- [x] `status` reports the last recorded sync time, sourced from the sync log written by `sync` (ticket 02)
- [x] `status` reports the next scheduled run time, derived from the platform scheduler where available
- [x] `status` produces a clear, distinct message if run before `install` has ever completed (no schedule to report on)
- [x] Both platform branches (darwin querying `launchctl`, linux querying `systemctl --user`) are covered
- [x] All scheduler queries and log reads are verified through the fake `System` — no real launchd/systemd interaction in tests

## Comments

Implemented via TDD in `internal/cli/status.go`, reusing the `scheduler.status` method both tickets 03/04 already built rather than adding a separate query path. "Next scheduled run" is computed deterministically from the fixed daily schedule constant shared with plist/unit generation (`nextDailyRun` in `scheduler.go`) plus a new `System.Now()` primitive, rather than parsed from live scheduler output — launchd has no simple stable command for it, and while systemd's `NextElapseUSecRealtime` property could supply it on Linux, computing it locally keeps both platforms' status logic identical and avoids depending on a duration-string format to parse; a legitimate "more literally platform-derived" alternative for Linux specifically, noted here rather than pursued. "Loaded/active" still does query the real scheduler (`launchctl list <label>` / `systemctl --user is-active <timer>`) per the ticket's ask, and `NextRun` is only populated when that query reports active — otherwise it's the zero value ("not scheduled"), since reporting a next-run time nothing will actually fire would be misleading. "Last recorded sync time" scans the sync log (ticket 02) backwards for the most recent `sync done` line; reports "never" if none exists. Pre-install behavior mirrors `sync`'s existing "run install first" error rather than inventing new phrasing.

Post-implementation code review (`/code-review`) caught two real bugs, fixed before commit: (1) `nextDailyRun` originally hardcoded `time.UTC`, but both `StartCalendarInterval` and unqualified `OnCalendar` are interpreted in the host's *local* time zone by their respective schedulers — fixed by computing in `now.Location()` instead, so the reported next-run matches what will actually fire on any host regardless of its time zone; (2) `status` originally computed `NextRun` whenever the schedule file existed, independent of the active check, so an unloaded/inactive schedule could still show a future next-run time — fixed by gating `NextRun` on `Active`.
