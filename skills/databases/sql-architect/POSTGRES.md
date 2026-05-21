# PostgreSQL 18 — Engine-Specific Defaults

Patterns, types, and features specific to PostgreSQL 18 (the primary engine targeted by `sql-architect`). Load this reference when designing schemas, choosing types, or reaching for a PG-specific feature.

## Types

- **`timestamptz` over `timestamp`** — always store with timezone; convert on display.
- **`text` over `varchar(n)`** — PG doesn't optimize varchar; `text` + a `CHECK (length(...) <= n)` constraint is cleaner when you need a cap.
- **`citext`** for case-insensitive text (emails, slugs) — eliminates a whole class of bugs.
- **Enums:** prefer **lookup tables** over `CREATE TYPE ... AS ENUM`. Enums can't be modified inside a transaction; lookup tables can. Use enum types only for tiny, truly fixed sets (e.g. `('asc', 'desc')`).
- **Native `uuidv7()`** (PG 18) — no extension needed for UUID v7 primary keys. On PG ≤ 17, use `pg_uuidv7` or generate in app code.

## Features worth reaching for

- **Generated columns** for computed values: `GENERATED ALWAYS AS (price * quantity) STORED`.
- **Declarative partitioning** for time-series at scale; the `pg_partman` extension automates partition maintenance.
- **`LISTEN` / `NOTIFY`** for low-volume pub/sub inside one Postgres deployment — fine for cache invalidation, not a substitute for a real message queue.
- **Row-Level Security** for multi-tenancy: `ENABLE ROW LEVEL SECURITY` + policies keyed on a session variable holding `tenant_id`. Defense in depth against app-layer mistakes.
- **`IS DISTINCT FROM`** for null-safe comparison: `WHERE a IS DISTINCT FROM b` treats two NULLs as equal.
- **`INSERT ... ON CONFLICT`** for atomic upserts (already covered in §4 of SKILL.md; restated here as a PG-native feature).

## Plan analysis tooling

- `EXPLAIN (ANALYZE, BUFFERS)` for IO-aware plans.
- `pg_stat_statements` extension for production query profiling.
- `pg_stat_user_indexes` to find unused indexes (`idx_scan = 0`).
- `pg_stat_activity` for live session inspection.
