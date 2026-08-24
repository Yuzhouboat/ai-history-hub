# 03: macOS scheduler wiring

**What to build:** On macOS, `install` sets up a daily automatic backup via `launchd`, and `uninstall` cleanly removes it. This is what turns ticket 02's manual `sync` into an unattended daily backup on Mac.

**Blocked by:** 02

**Status:** done

- [x] On `GOOS=darwin`, `install` generates a `launchd` `.plist` under `~/Library/LaunchAgents/` that invokes `sync` on a daily schedule via `StartCalendarInterval`
- [x] `install` loads the generated agent (equivalent of `launchctl load`) so the schedule is active immediately after setup
- [x] Re-running `install` on darwin updates the existing plist/schedule in place (e.g. a changed remote or exclusion list) rather than erroring or duplicating the scheduled job
- [x] `uninstall` on darwin unloads the agent and removes the generated plist file
- [x] `uninstall` does not touch already-uploaded S3 data or the local exclusion config
- [x] All plist content generation and `launchctl` invocations are verified through the fake `System` — asserting exact file content and exact command/args, no real launchd interaction in tests

## Comments

Implemented via TDD as `darwinScheduler` in `internal/cli/scheduler_launchd.go`, tested directly against `system.Fake` in `scheduler_launchd_test.go` rather than through `cli.Run` — that's what makes the darwin path testable at all on a non-darwin dev/CI machine, since dispatch to a platform scheduler happens by `runtime.GOOS` (see `internal/cli/scheduler.go`'s `schedulerFor`), which the test suite can't override when going through the top-level `cli.Run` entrypoint. `install` re-runs are made idempotent by unconditionally issuing a best-effort `launchctl unload` before every `launchctl load -w` (harmless no-op the first time, and what makes re-running update the existing schedule in place instead of erroring on an already-loaded label). `uninstall` is a no-op (not an error) when the plist was never written, since nothing scheduler-side needs cleaning up — this also covers ticket 05's status command reporting cleanly on an uninstalled schedule. The `System` seam grew three more primitives for this: `Executable()` (so the generated plist points at the real binary), `Now()` (deterministic "next run" computation in tests), and `RemoveFile()` (uninstall needs a way to delete a file, which didn't exist before this ticket).
