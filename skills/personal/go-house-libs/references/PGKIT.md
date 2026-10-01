# pgkit

`github.com/ralvarezdev/pgkit` — small helpers for services built on pgx and sqlc: value conversion, transactions, embedded goose migrations, the Postgres `Idempotency-Key` store, pool metrics and a test container. Each heavy dependency lives in exactly one subpackage.

## Packages

| Package | Owns | Only package importing |
|---|---|---|
| `pgkit` | `Text`/`TextPtr`, `UUID`/`UUIDString`/`OptUUID`, typed-UUID `UUIDOf`/`OptUUIDOf`/`UUIDPtr`, `Timestamptz`/`OptTimestamptz`/`TimePtr`, `Date`/`OptDate`/`DatePtr`/`DateVal`, `NullDecimal`, `MapNoRows`, `MapRows`, `RunInTx` (pgx), `RunInSQLTx` (database/sql) | pgx, shopspring/decimal |
| `pgkit/pgerr` | SQLSTATE classification: `HasCode`, `IsUniqueViolation`, `IsForeignKeyViolation`, `IsRestrictViolation`, codes `UniqueViolation`, `ForeignKeyViolation`, `RestrictViolation`, `LockNotAvailable` | pgconn (via pgx) |
| `pgkit/migrate` | embedded goose migrations with a per-module version table: `Up`, `Down`, `Version`, `Status`, `UpDSN`, `NewProvider` | goose |
| `pgkit/idempotency` | Postgres store for `Idempotency-Key`: `Reserve`/`Complete`/`Release` with fenced claims, takeover of abandoned pending claims, `RunCleanup` | (implements `restkit/idempotency.Store`) |
| `pgkit/pgmetrics` | pool statistics as Prometheus `db_pool_*` metrics: `Register(reg, pool, labels)` | prometheus client |
| `pgkit/pgtest` | throwaway Postgres for tests: `StartContainer`, `NewPool`, `NewDB` per test (`Image` defaults to `postgres:18-alpine`); `Start` for a shared `TestMain` container | testcontainers |

## Reach for it when

- mapping sqlc rows to domain types with nullable columns: the conversion helpers, not local copies;
- a store method must return a domain error for no rows: `MapNoRows(err, ErrNotFound, "get loan")`;
- a store must tell a unique or foreign-key violation from other errors: `pgerr.IsUniqueViolation(err)`, not string matching or a local `errors.As` on `*pgconn.PgError`;
- several writes must commit or roll back together: `RunInTx`;
- a module ships its own schema: embed the migrations and expose `migrate.Up` (identity and webpush do);
- ids are typed `uuid.UUID` (any `~[16]byte`): `UUIDOf`/`OptUUIDOf`/`UUIDPtr[uuid.UUID]`, no uuid import in pgkit; `date` columns: `Date`/`DatePtr`;
- an integration test needs a real database.

## Recipe

```go
// Conversions.
row.Note = pgkit.OptText(p.Note)        // nil -> NULL
id := pgkit.UUIDString(row.ID)          // invalid -> ""
d := pgkit.NullDecimal(amount)          // zero -> NULL
loans := pgkit.MapRows(rows, toLoan)

// Not-found mapping and transactions.
if err != nil { return Loan{}, pgkit.MapNoRows(err, loan.ErrNotFound, "get loan") }
err = pgkit.RunInTx(ctx, pool, func(tx pgx.Tx) error { ... })
err = pgkit.RunInSQLTx(ctx, sqlDB, func(tx *sql.Tx) error { ... }) // *sql.DB or *sql.Conn
if pgerr.IsUniqueViolation(err) { return ErrAlreadyExists }
if pgerr.HasCode(err, pgerr.LockNotAvailable) { /* retry */ }
id := pgkit.UUIDOf(userID)                                      // uuid.UUID -> pgtype.UUID
uid := pgkit.UUIDPtr[uuid.UUID](row.ParentID)                   // NULL -> nil; type argument is explicit

// Module migrations, each with its own version table so several modules share one database.
err = migrate.Up(ctx, sqlDB, migrate.Config{FS: migrations, TableName: "app_schema_version"})

// Tests: one pool per test, or one container per package.
pool := pgtest.NewPool(t, pgtest.Config{Migrate: runMigrations})
dsn, stop, err := pgtest.Start(ctx, cfg) // TestMain; never skips, the caller decides what an error means
```

## Rules

- **`RunInTx` and `RunInSQLTx` always finish the transaction**: commit, or roll back on error or panic (the panic is not recovered), even after the context is cancelled. `RunInSQLTx` begins with default `TxOptions`; for others call `BeginTx` yourself.
- **`Date` keeps the time as given**; `DateVal` maps NULL to the zero time, `DatePtr` to nil.
- **`pgtest` skips in `-short` mode or without Docker** and cleans up through `tb.Cleanup`; `Config.Image` is optional, default `postgres:18-alpine` (`ErrImageRequired` is deprecated and never returned; a digest-pinned reference works; `Extensions` adds `CREATE EXTENSION` for images such as timescaledb).
- **Tests over mocks.** Store tests run against a real container, not a mocked pool.
- **Migrations are goose, forward-only**, one version table per module (`identity_schema_version`, `webpush_schema_version`). Do not share a table between modules.
- **Idempotency schema**: create the table with the embedded migration (`idempotency.Migrations`, run through `migrate`) or apply `idempotency.Schema` yourself. Wire the store into `ginkit.Idempotency`; the kits never import each other's concrete types.
- **House SQL stack**: goose for migrations, squirrel for building queries, sqlc for generated ones, wherever each applies. Do not concatenate SQL strings.

## Not here

Repositories or business rules (those stay in the app's domain packages), HTTP glue (`ginkit`), the idempotency contract types (`restkit/idempotency`).
