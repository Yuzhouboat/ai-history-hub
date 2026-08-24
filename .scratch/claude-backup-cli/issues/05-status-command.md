# 05: status command

**What to build:** A user can run `claude-backup status` to confirm their backup schedule is actually active and see when it last succeeded and when it will next run, on whichever platform (macOS or Linux) they're on.

**Blocked by:** 03, 04

**Status:** ready-for-agent

- [ ] `status` reports whether the schedule is currently loaded/active, querying the platform scheduler (`launchctl` on darwin, `systemctl --user` on linux)
- [ ] `status` reports the last recorded sync time, sourced from the sync log written by `sync` (ticket 02)
- [ ] `status` reports the next scheduled run time, derived from the platform scheduler where available
- [ ] `status` produces a clear, distinct message if run before `install` has ever completed (no schedule to report on)
- [ ] Both platform branches (darwin querying `launchctl`, linux querying `systemctl --user`) are covered
- [ ] All scheduler queries and log reads are verified through the fake `System` — no real launchd/systemd interaction in tests
