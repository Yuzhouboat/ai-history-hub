# Go + rclone, with rclone owning all S3 credentials

`claude-backup` needs to run as a portable single binary across Linux and Mac, and needs to talk to S3. We chose Go (single static cross-compiled binary, no runtime to install on target machines) driving rclone as the actual transfer engine, rather than a Python/boto3 implementation that would own S3 credentials directly.

Credentials are handled entirely by rclone's own remote config, not by `claude-backup`: `install` takes a `--remote <name>`, checks whether it already exists (`rclone listremotes`), and if not, shells out to rclone's own interactive `rclone config` wizard to create it. `claude-backup` never reads, stores, or passes AWS keys itself.

This was a deliberate trade-off: rclone already has a mature, well-tested credential story (including OS keychain integration), and reimplementing that in our own wrapper would be duplicated, worse-tested work with no benefit. The cost is an external dependency: rclone must already be on PATH (we don't auto-install it — see [ADR-0003](./0003-no-client-side-encryption.md) sibling reasoning on trust surface), and `install` fails fast with instructions if it's missing.

## Considered options

- **Python + boto3, owning S3 credentials directly** — rejected: would require re-implementing credential storage/rotation that rclone already solves, and adds a Python runtime dependency on every target machine.
- **Bash + AWS CLI v2** — rejected: weakest option for a "real" distributable CLI tool (subcommands, cross-platform scheduler management), and still pulls in an external CLI dependency.
