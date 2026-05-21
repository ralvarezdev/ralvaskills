# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| python | 3.14 | Language |
| fastapi | 0.136 | Web framework |
| uvicorn | 0.47 | ASGI server (prod: `uvicorn ... --workers N`, dev: `--reload`) |
| pydantic | 2.13 | Request / response models, validation |
| pydantic-settings | 2.14 | Layered settings (env + file) |
| psycopg | 3.3 | Postgres driver — see [sql-architect](../../databases/sql-architect/STACK.md) |
| httpx | 0.28 | Async outbound HTTP client (in `lifespan`) |
| pyjwt | 2.12 | JWT issuance + verification (auth pattern A) |
| argon2-cffi | 25.1 | Password hashing — Argon2id, OWASP default |
| python-multipart | 0.0.29 | OAuth2 password-flow form parsing |
| pytest | 9.0 | Test runner |
| pytest-asyncio | 1.3 | Async test support |

## Notes

- **Inherits the `python-architect` stack.** This file pins what is *specific* or *additionally required* for FastAPI services. For language/tooling pins (`mypy`, `ruff`, `uv`, etc.), see [python-architect/STACK.md](../../languages/python-architect/STACK.md).
- **No ORM by default.** Data access goes through `psycopg 3` + `.sql` files via `importlib.resources` — see [sql-architect](../../databases/sql-architect/SKILL.md).
- **Auth library choice depends on the pattern** ([§6](SKILL.md)):
  - Pattern A (in-house OAuth2 + JWT): `pyjwt` + `argon2-cffi` + `python-multipart`
  - Pattern B (external IdP): only `pyjwt` for JWKS verification; password / multipart not needed
- **Rate limiting / DOS protection:** add `slowapi` (0.1.9) or front with a gateway. Not pinned here because it's deployment-shape-dependent.

_Last reviewed: 2026-05-20_
_Skill version at last review: 1.0.0_
