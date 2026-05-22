# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| postgresql | 18.4 | **Primary** target engine; native `uuidv7()`, declarative partitioning, RLS |
| mysql | 9.7 | Secondary — covered with inline notes where it diverges |
| sqlite | 3.53 | Secondary — embedded / single-process use |
| golang-migrate | 4.19 | Migration runner for Go projects — raw `.sql` files, forward-only |
| alembic | 1.18 | Migration runner for Python projects — raw `.sql` files supported |
| pgbouncer | 1.24 | Connection pooler in front of Postgres (transaction mode for OLTP) |

## Notes

- **Soft delete is the default.** Hard delete only for GDPR/compliance requirements or append-only event tables.
- **Surrogate UUID v7 primary keys** + `UNIQUE` on natural identifiers. Avoid `BIGSERIAL` unless storage cost is genuinely the bottleneck on FK-heavy tables.
- **Forward-only migrations.** No down migrations in production.
- **`timestamptz` over `timestamp`** on PostgreSQL. Always store with timezone.
- **JSONB only for genuinely schemaless data.** Anything structured belongs in normalized columns.

_Last reviewed: 2026-05-20_
_Skill version at last review: 1.0.0_