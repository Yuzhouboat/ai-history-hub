# Additive-only backup: never delete from the S3 remote

Claude Code auto-deletes local session transcripts older than 30 days by default (`cleanupPeriodDays` in `~/.claude/settings.json`, unset on the reference machine). A naive mirror — `rclone sync`, which prunes the destination to match the source — would propagate that local deletion to S3 and erase the backup at exactly the moment it becomes the only remaining copy.

`claude-backup` therefore uses `rclone copy`, not `rclone sync`: it uploads new or changed files but never deletes an S3 object because the corresponding local file is gone. This is a one-way, additive-only backup, not a mirror. See the "Backup" glossary entry in [CONTEXT.md](../../CONTEXT.md).

## Consequences

S3 storage grows monotonically (no deletion path exists in this tool). Given the source data is small (~5MB across all projects at time of writing) and S3 storage is cheap, this is an accepted trade-off rather than something requiring a retention/pruning feature.
