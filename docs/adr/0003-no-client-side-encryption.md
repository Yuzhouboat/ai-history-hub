# No client-side encryption layer; rely on S3 SSE + private bucket

Chat history can contain pasted secrets, credentials, or proprietary code, so encryption was a real question. We chose to rely on S3 server-side encryption and a private bucket ACL, and deliberately did not add an rclone `crypt` overlay (client-side encryption before upload).

The threat model this defends against — AWS itself, or someone with access to the AWS account, reading a private bucket's contents at rest — is far down the list of realistic risks for a personal backup where the user already controls the AWS account. `crypt` would add a second secret that must never be lost (losing it makes the entire backup unreadable), which is a worse failure mode for a backup tool than the risk it defends against. If the threat model changes (e.g., a shared or less-trusted AWS account), this should be revisited.
