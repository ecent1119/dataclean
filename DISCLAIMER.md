# Disclaimer

## Intended Use

dataclean is a **local development data management tool**. It creates, restores, and manages snapshots of Docker Compose data volumes to help developers quickly iterate on local development and testing environments.

## Explicit Scope

dataclean is intentionally scoped for **local development and testing**.

It operates on **Docker volumes on your local machine** and is designed for rapid iteration, not data backup or production use.

## Destructive Operations

**dataclean performs destructive operations on Docker volumes.**

The `restore` and `reset` commands will:
- **DELETE existing data** in target volumes
- **Replace** volume contents with snapshot data (restore) or empty state (reset)

These operations are:
- **Irreversible** without a prior snapshot
- **Gated by confirmation** unless `--force` is specified
- **Auto-backed up** by default before execution (configurable)

## What This Is Not

* Not a backup solution
* Not for production data
* Not a disaster recovery tool
* Not a database migration tool
* Not a data synchronization service

## Safety Features

dataclean includes safety measures:

1. **Confirmation prompts**: Destructive operations require typing "yes" or `--force`
2. **Dry run mode**: Preview operations with `--dry-run`
3. **Auto-backup**: Creates snapshot before restore/reset (enabled by default)
4. **Local-only**: Only operates on Docker volumes on your machine

## No Warranties

dataclean is provided "as is" without warranty of any kind. The authors and distributors:

- Make no claims about data integrity or reliability
- Are not responsible for data loss arising from use
- Do not guarantee compatibility with any specific Docker version
- Recommend always maintaining separate backups of important data

## Your Responsibility

By using dataclean, you acknowledge that:

1. You understand it performs destructive operations
2. You have appropriate backups of any important data
3. You are operating on local development data only
4. You accept full responsibility for any data loss

## Third-Party Software

dataclean interacts with third-party software including:

| Software | Purpose |
|----------|---------|
| Docker | Container and volume management |
| Alpine Linux | Temporary containers for volume operations |

These are not included with dataclean. You must have Docker installed and running.

## License

This software is provided under the MIT License. See [LICENSE](LICENSE) for full terms.
