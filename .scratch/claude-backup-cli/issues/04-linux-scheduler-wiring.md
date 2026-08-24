# 04: Linux scheduler wiring

**What to build:** On Linux, `install` sets up a daily automatic backup via a `systemd` user timer, and `uninstall` cleanly removes it, including catching up a run that was missed while the machine was off or asleep. This is what turns ticket 02's manual `sync` into an unattended daily backup on Linux.

**Blocked by:** 02 (parallel with 03, not blocked by it)

**Status:** ready-for-agent

- [ ] On `GOOS=linux`, `install` generates a `systemd` user `.service` + `.timer` pair under `~/.config/systemd/user/` that invokes `sync` on a daily schedule via `OnCalendar`
- [ ] The generated timer sets `Persistent=true` so a run missed while the machine was off/asleep at the scheduled time fires on next boot/login instead of being silently skipped
- [ ] `install` enables and starts the timer (equivalent of `systemctl --user enable --now`) so the schedule is active immediately after setup
- [ ] Re-running `install` on linux updates the existing unit/timer/schedule in place rather than erroring or duplicating the scheduled job
- [ ] `uninstall` on linux disables/stops the timer and removes the generated unit and timer files
- [ ] `uninstall` does not touch already-uploaded S3 data or the local exclusion config
- [ ] All unit/timer content generation and `systemctl` invocations are verified through the fake `System` — asserting exact file content and exact command/args, no real systemd interaction in tests
