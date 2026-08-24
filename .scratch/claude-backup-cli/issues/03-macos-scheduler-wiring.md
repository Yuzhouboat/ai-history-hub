# 03: macOS scheduler wiring

**What to build:** On macOS, `install` sets up a daily automatic backup via `launchd`, and `uninstall` cleanly removes it. This is what turns ticket 02's manual `sync` into an unattended daily backup on Mac.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] On `GOOS=darwin`, `install` generates a `launchd` `.plist` under `~/Library/LaunchAgents/` that invokes `sync` on a daily schedule via `StartCalendarInterval`
- [ ] `install` loads the generated agent (equivalent of `launchctl load`) so the schedule is active immediately after setup
- [ ] Re-running `install` on darwin updates the existing plist/schedule in place (e.g. a changed remote or exclusion list) rather than erroring or duplicating the scheduled job
- [ ] `uninstall` on darwin unloads the agent and removes the generated plist file
- [ ] `uninstall` does not touch already-uploaded S3 data or the local exclusion config
- [ ] All plist content generation and `launchctl` invocations are verified through the fake `System` — asserting exact file content and exact command/args, no real launchd interaction in tests
