# dataclean

**Free and open source** — Local Dev Data Reset & Snapshot Tool

Snapshot, restore, and reset Docker Compose data volumes for rapid development iteration.

## Quick Start

```bash
# In a directory with docker-compose.yaml
dataclean snapshot before-migration   # Save current state
dataclean list                        # Show snapshots
dataclean restore before-migration    # Restore saved state
dataclean reset                       # Wipe to empty state
```

## Features

- **Auto-detection**: Finds volumes from `compose.yaml` / `docker-compose.yaml`
- **Smart datastore detection**: Recognizes Postgres, MySQL, Redis, MongoDB, Neo4j
- **Safe by default**: Destructive operations require `--force` or confirmation
- **Auto-backup**: Creates backup before restore/reset operations
- **Project-local**: Snapshots stored in `.dataclean/` (gitignore-friendly)
- **Snapshot tagging**: Add tags for organization and filtering
- **Metadata support**: Add descriptions and custom metadata to snapshots
- **Include/exclude filters**: Snapshot only specific volumes
- **Size reporting**: See snapshot sizes and datastore size breakdowns
- **Retention policies**: Auto-cleanup old snapshots

## Installation

### macOS (Homebrew)

```bash
brew install stackgen-cli/tap/dataclean
```

### Manual Installation

Download from [releases](https://github.com/stackgen-cli/dataclean/releases) and add to PATH:

```bash
# macOS/Linux
chmod +x dataclean
sudo mv dataclean /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/stackgen-cli/dataclean
cd dataclean
make build
```

## Commands

### `dataclean snapshot [name]`

Create a named snapshot of current data state.

```bash
dataclean snapshot                    # auto-named: snapshot-2024-01-15-143052
dataclean snapshot before-migration   # named: before-migration
dataclean snapshot --dry-run          # preview only
dataclean snapshot --tag production   # add a tag
dataclean snapshot --description "Before schema v2 migration"
dataclean snapshot --include postgres_data --include redis_data
dataclean snapshot --exclude tmp_cache
```

### `dataclean restore <name>`

Restore data from a named snapshot. **Destructive** - replaces current data.

```bash
dataclean restore before-migration          # prompts for confirmation
dataclean restore before-migration --force  # skip confirmation
dataclean restore before-migration --dry-run
```

### `dataclean reset`

Wipe all volumes to empty state. **Destructive** - deletes all data.

```bash
dataclean reset          # prompts for confirmation
dataclean reset --force  # skip confirmation
dataclean reset --dry-run
```

### `dataclean list`

Show all available snapshots.

```bash
dataclean list
```

Output:

```
NAME               CREATED           SIZE    VOLUMES
before-migration   2024-01-15 14:30  45.2 MB 3
fresh-install      2024-01-14 09:15  12.1 MB 3
```

## Configuration

dataclean works with zero configuration by auto-detecting from `compose.yaml`.

For advanced use, create `.dataclean.yaml`:

```yaml
# Optional: explicit compose file path
compose_file: docker-compose.yaml

# Optional: only snapshot these volumes (default: all)
include_volumes:
  - postgres_data
  - redis_data

# Optional: exclude these volumes
exclude_volumes:
  - tmp_cache

# Optional: override datastore type detection
datastore_hints:
  custom_volume: postgres

# Optional: custom snapshot directory
snapshot_dir: .dataclean

# Optional: auto-backup before restore/reset (default: true)
backup_before_restore: true
```

## Supported Datastores

| Datastore | Detection | Native Tools |
|-----------|-----------|--------------|
| PostgreSQL | `postgres:*` images, `/var/lib/postgresql` | `pg_dump` / `pg_restore` |
| MySQL/MariaDB | `mysql:*`, `mariadb:*` images | `mysqldump` / `mysql` |
| Redis | `redis:*` images | `redis-cli --rdb` |
| MongoDB | `mongo:*` images | `mongodump` / `mongorestore` |
| Neo4j | `neo4j:*` images | Volume backup |
| Generic | Any other volume | `tar` archive |

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation prompts |
| `--dry-run` | | Preview without making changes |
| `--quiet` | `-q` | Minimal output (for CI/scripts) |
| `--config` | | Specify config file path |

## Example Workflow

```bash
# Start fresh development environment
docker compose up -d
dataclean snapshot fresh-install

# Work on feature, add test data
# ... development ...

# Need to test migration? Save current state
dataclean snapshot before-migration

# Run migration
./migrate.sh

# Something went wrong? Restore
dataclean restore before-migration --force

# Start completely fresh
dataclean reset --force
dataclean restore fresh-install --force
```

## Snapshot Storage

Snapshots are stored in `.dataclean/` in your project directory:

```
.dataclean/
├── before-migration/
│   ├── metadata.yaml
│   ├── myproject_postgres_data.tar.gz
│   └── myproject_redis_data.tar.gz
└── fresh-install/
    ├── metadata.yaml
    └── ...
```

Add to `.gitignore`:

```
.dataclean/
```

## Part of the Docker Dev Toolchain

dataclean is part of the local development toolchain:

1. **[stackgen](https://github.com/ecent1119/stackgen)** → Generate local dev stacks
2. **[envgraph](https://github.com/ecent1119/envgraph)** → Inspect & validate config/env
3. **dataclean** → Manage local data state

## Support This Project

**dataclean is free and open source.**

If this tool saved you time, consider sponsoring:

[![Sponsor on GitHub](https://img.shields.io/badge/Sponsor-❤️-red?logo=github)](https://github.com/sponsors/ecent1119)

Your support helps maintain and improve this tool.

## License

MIT License. See [LICENSE](LICENSE).

## Disclaimer

**For local development and testing only.**

This tool performs destructive operations on Docker volumes. Always ensure you have backups of important data. See [DISCLAIMER.md](DISCLAIMER.md).
