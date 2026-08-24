# 01: Project scaffold & System seam

**What to build:** The `claude-backup` Go project exists, builds to a binary, and recognizes its four subcommands (`install`, `sync`, `status`, `uninstall`) via CLI dispatch, each currently stubbed. The `System` interface — the single seam abstracting all external command execution and file I/O — is defined, backed by a real implementation (`os/exec` + `os`) and an in-memory fake for tests. Every later ticket's subcommand logic will depend only on this interface for its side effects.

**Blocked by:** None (can start immediately)

**Status:** done

- [x] `go build` produces a `claude-backup` binary
- [x] Running the binary with no args or `--help` lists the four subcommands
- [x] Each subcommand is routed to distinct (currently stub) logic
- [x] `System` interface exists with methods covering: running an external command (name + args, capturing stdout/stderr/exit) and reading/writing a file
- [x] A real `System` implementation exists using `os/exec` and `os`
- [x] An in-memory fake `System` implementation exists for tests, recording invoked commands and file writes for assertions
- [x] At least one test demonstrates a stub subcommand exercised through the fake `System`, establishing the pattern later tickets will follow

## Comments

Checkboxes/status updated retroactively — the work landed in commit 68e57e4 but the ticket file was never marked done at the time. Ticket 02 (commit ae2cd23) built directly on this seam, confirming it holds up as designed.
