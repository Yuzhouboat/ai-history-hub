# 04: Linux scheduler wiring

**What to build:** On Linux, `install` sets up a daily automatic backup via a `systemd` user timer, and `uninstall` cleanly removes it, including catching up a run that was missed while the machine was off or asleep. This is what turns ticket 02's manual `sync` into an unattended daily backup on Linux.

**Blocked by:** 02 (parallel with 03, not blocked by it)

**Status:** done

- [x] On `GOOS=linux`, `install` generates a `systemd` user `.service` + `.timer` pair under `~/.config/systemd/user/` that invokes `sync` on a daily schedule via `OnCalendar`
- [x] The generated timer sets `Persistent=true` so a run missed while the machine was off/asleep at the scheduled time fires on next boot/login instead of being silently skipped
- [x] `install` enables and starts the timer (equivalent of `systemctl --user enable --now`) so the schedule is active immediately after setup
- [x] Re-running `install` on linux updates the existing unit/timer/schedule in place rather than erroring or duplicating the scheduled job
- [x] `uninstall` on linux disables/stops the timer and removes the generated unit and timer files
- [x] `uninstall` does not touch already-uploaded S3 data or the local exclusion config
- [x] All unit/timer content generation and `systemctl` invocations are verified through the fake `System` — asserting exact file content and exact command/args, no real systemd interaction in tests

## Comments

Implemented via TDD as `linuxScheduler` in `internal/cli/scheduler_systemd.go` (paired with `darwinScheduler` from ticket 03 behind a shared `scheduler` interface in `internal/cli/scheduler.go`). Unlike the darwin path, linux is also the sandbox's host platform, so this scheduler is exercised both directly against `system.Fake` (`scheduler_systemd_test.go`, for exact unit content and command assertions) and end-to-end through `cli.Run` (`install_test.go`/`uninstall_test.go`/`status_test.go`), since `runtime.GOOS` resolves to `linux` here. `install` always rewrites both unit files, runs `daemon-reload`, then `enable --now` — safe to repeat, so re-running updates the schedule in place without needing to diff old vs. new content. `uninstall` is a no-op (not an error) when the timer file was never written. The daily run time (03:00 UTC) is a shared constant (`scheduledHour`/`scheduledMinute` in `scheduler.go`) used both for `OnCalendar` generation and for status's next-run computation in ticket 05, so the two can't drift apart.
