# 01: Project scaffold & System seam

**What to build:** The `claude-backup` Go project exists, builds to a binary, and recognizes its four subcommands (`install`, `sync`, `status`, `uninstall`) via CLI dispatch, each currently stubbed. The `System` interface — the single seam abstracting all external command execution and file I/O — is defined, backed by a real implementation (`os/exec` + `os`) and an in-memory fake for tests. Every later ticket's subcommand logic will depend only on this interface for its side effects.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] `go build` produces a `claude-backup` binary
- [ ] Running the binary with no args or `--help` lists the four subcommands
- [ ] Each subcommand is routed to distinct (currently stub) logic
- [ ] `System` interface exists with methods covering: running an external command (name + args, capturing stdout/stderr/exit) and reading/writing a file
- [ ] A real `System` implementation exists using `os/exec` and `os`
- [ ] An in-memory fake `System` implementation exists for tests, recording invoked commands and file writes for assertions
- [ ] At least one test demonstrates a stub subcommand exercised through the fake `System`, establishing the pattern later tickets will follow
