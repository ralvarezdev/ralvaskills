# Alternative Engines — MySQL & SQLite

Notes for engines other than the PostgreSQL primary. Load this reference when targeting MySQL 9 or SQLite 3.53. Patterns in the main `SKILL.md` apply unchanged unless called out here.

## MySQL 9

- **Charset:** `utf8mb4` with `utf8mb4_0900_ai_ci` collation. **Never** legacy `utf8` (which is 3-byte and rejects emojis).
- **Storage engine:** InnoDB only. Other engines lose transactional integrity.
- **Primary keys:** no `BIGSERIAL` — use `BIGINT UNSIGNED AUTO_INCREMENT`. For UUID v7 keys, store as `BINARY(16)` and generate in application code (no native `uuidv7()` equivalent).
- **JSON:** native JSON type exists, but JSONB-style indexing is weaker than PostgreSQL. Use **generated columns + secondary index** when you need to query into JSON paths.
- **Time zones:** `TIMESTAMP` stores in UTC and converts on read; `DATETIME` is naïve. Prefer `TIMESTAMP` with the session timezone fixed to `+00:00`.
- **CTEs and window functions** are supported in MySQL 8+; no compatibility concern in MySQL 9.

## SQLite 3.53

- **Primary keys:** `INTEGER PRIMARY KEY` becomes the rowid alias — fastest, but you lose UUID v7 benefits (URL non-enumerability, distributed generation). Use `TEXT PRIMARY KEY` with app-generated UUID v7 when those matter.
- **No native enum types** and **no row-level security** — enforce via `CHECK` constraints and application code.
- **JSON1 extension** is built-in (`json_extract`, `->`, `->>`). Indexing via generated columns.
- **Concurrency:** single-writer at a time (database-level lock). Fine for embedded / single-process use; not for multi-writer servers.
- **WAL mode:** enable with `PRAGMA journal_mode = WAL;` for any non-trivial use — allows readers to run concurrently with one writer.
- **Foreign keys:** off by default. Enable per-connection with `PRAGMA foreign_keys = ON;` or in the application's connection setup.
