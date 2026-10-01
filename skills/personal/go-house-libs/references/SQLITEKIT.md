# sqlitekit

`github.com/ralvarezdev/sqlitekit` — one correct way to open SQLite from Go with the pure Go `modernc.org/sqlite` driver (no cgo): WAL, foreign keys and a busy timeout on every pooled connection, a single-connection default, an in-memory opener for tests and embedded goose migrations. The SQLite counterpart of [PGKIT](PGKIT.md).

## Packages

| Package | Owns |
|---|---|
| `sqlitekit` | `Open`, `OpenMemory`, `DSN`, `Config`; depends on modernc.org/sqlite only |
| `sqlitekit/migrate` | embedded goose migrations per module (`Up`, `Down`, `Version`, `Status`, `Open`, `OpenMemory`) with its own version table and no goose global state; the only package importing goose |

## Recipe

```go
//go:embed migrations/*.sql
var migrations embed.FS

db, err := migrate.Open(ctx, path, sqlitekit.Config{}, migrate.Config{
    FS: migrations, Dir: "migrations", TableName: "app_schema_version",
})
```

`sqlitekit.Config{}` means WAL, `synchronous=NORMAL`, foreign keys on, 5s busy timeout, `MaxOpenConns` 1. Tests: `migrate.OpenMemory` (a uniquely named shared-cache database, isolated per call).

## Rules

- **Pragmas go in the DSN, not in `db.Exec`.** A `PRAGMA` sent with `Exec` reaches only the one pooled connection that ran it. sqlitekit writes every pragma as `_pragma=name(value)`, which the driver applies to each connection.
- **Never use the mattn spellings** (`?_journal_mode=WAL&_busy_timeout=5000`) with modernc: v1.52.0 silently ignores them (the database stays on `journal_mode=delete` with `busy_timeout=0`). vtitan's session recorder shipped with this bug. Assert the pragmas in a test (`PRAGMA journal_mode`, `PRAGMA busy_timeout`).
- **Adopting an existing database:** make the baseline migration idempotent (`CREATE TABLE IF NOT EXISTS`) so files created before goose open cleanly and keep their rows; test it by opening a legacy file. Give each module its own `TableName`.
- **Switching an old database to WAL** adds `-wal` and `-shm` files next to it on first open.
- **Requires modernc v1.60.1 and Go 1.27.1**; adopting it bumps the consumer's driver.
- **One connection by default.** Raise `MaxOpenConns` only when readers must run beside a writer, and then expect `SQLITE_BUSY` handling.

## Not here

`_txlock=immediate` and other driver options, a read/write pool split, backup and `VACUUM INTO` helpers, a `sqlitetest` helper.
