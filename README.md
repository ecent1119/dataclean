# dataclean

Snapshot, restore, and reset Docker Compose data volumes for rapid development iteration.

---

## The problem

- "Let me just wipe the database and start fresh" → 20 minutes later
- Migration testing requires clean state repeatedly
- Switching branches with different schema versions
- "It worked before I ran that migration"
- Manual volume cleanup is error-prone

---

## What it does

- Auto-detects Docker Compose volumes
- Creates named snapshots of current data state
- Restores snapshots instantly
- Resets to clean baseline
- Safe by default — destructive ops require confirmation

---

## New in v2.0

- **Snapshot tagging** — add tags for organization and filtering
- **Metadata support** — add descriptions and custom metadata
- **Include/exclude filters** — snapshot only specific volumes
- **Size reporting** — see snapshot sizes and datastore breakdowns
- **Retention policies** — auto-cleanup old snapshots

---

## Example output

```bash
$ dataclean snapshot before-migration
📸 Creating snapshot: before-migration

  • myapp_postgres_data (postgres)
  • myapp_redis_data (redis)

✅ Snapshot created: before-migration
   Size: 45.2 MB
   Path: .dataclean/before-migration/

$ dataclean list
NAME               CREATED           SIZE     VOLUMES
before-migration   2024-01-15 14:30  45.2 MB  2
fresh-install      2024-01-14 09:15  12.1 MB  2

$ dataclean restore before-migration --force
🔄 Restoring snapshot...
✅ Restored snapshot: before-migration
```

---

## Commands

| Command | Description |
|---------|-------------|
| `dataclean snapshot [name]` | Create named snapshot |
| `dataclean restore <name>` | Restore from snapshot |
| `dataclean reset` | Wipe to empty state |
| `dataclean list` | Show available snapshots |

---

## Supported datastores

| Datastore | Detection |
|-----------|-----------|
| PostgreSQL | `postgres:*` images |
| MySQL/MariaDB | `mysql:*`, `mariadb:*` images |
| Redis | `redis:*` images |
| MongoDB | `mongo:*` images |
| Neo4j | `neo4j:*` images |
| Generic | Any Docker volume |

---

## Scope

- Local development and testing only
- Destructive operations gated by `--force` or confirmation
- Auto-backup before restore/reset
- No production use
- No telemetry

---

## Get it

👉 [Download on Gumroad](https://ecent.gumroad.com/l/sklwb)

---

## Related tools

| Tool | Purpose |
|------|---------|
| **[stackgen](https://github.com/stackgen-cli/stackgen)** | Generate local dev Docker Compose stacks |
| **[envgraph](https://github.com/stackgen-cli/envgraph)** | Scan and validate environment variable usage |
| **[compose-diff](https://github.com/stackgen-cli/compose-diff)** | Semantic Docker Compose diff |
| **[devcheck](https://github.com/stackgen-cli/devcheck)** | Local project readiness inspector |

---

## License

MIT — this repository contains documentation and examples only.
