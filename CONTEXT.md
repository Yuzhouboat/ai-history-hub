# aiHistoryHub

Tooling for preserving and backing up local AI assistant history. The first tool built here is `claude-backup`, a CLI that periodically ships Claude Code's local chat history to S3.

## Language

**Chat history**:
The full local record Claude Code keeps on disk: session transcripts plus the global prompt index. This is the thing `claude-backup` protects.

**Session transcript**:
The record of one Claude Code session, stored locally as one file under `~/.claude/projects/<project>/<session-id>.jsonl`. One project can have many session transcripts.

**claude-backup**:
The CLI tool this repo ships: a Go binary that drives rclone to copy chat history to S3 on a schedule, and self-manages that schedule via `install`/`uninstall`/`status`.

**Backup** (as opposed to _mirror_ or _sync_ in the destructive sense):
In this project, backing up means additive-only replication — new or changed files are copied to S3, but nothing already in S3 is ever removed, even if the local copy is gone. Deliberately distinct from a mirror, which would also propagate local deletions to the remote.
_Avoid_: "sync" alone when the additive-only guarantee matters — say "backup" or "additive-only sync" instead, since "sync" is ambiguous with rclone's own destructive `sync` verb.

**Remote**:
An rclone remote: a named, credentialed connection to a storage backend (here, an S3 bucket), owned entirely by rclone's own config — `claude-backup` never stores or handles the underlying credentials itself.

**Excluded project**:
A project directory under `~/.claude/projects/` the user has opted out of backup. The default is no exclusions — every project is backed up unless explicitly excluded.
