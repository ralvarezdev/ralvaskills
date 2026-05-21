---
name: security-reviewer
version: 1.0.0
description: Cross-language security review — injection, auth/authz, secret leakage, insecure defaults, deserialization, CSRF, SSRF, dependency vulns. Produces a Critical/High/Medium/Low findings report keyed to file:line, with fix suggestions and tool references (semgrep, gosec, bandit, gitleaks, Trivy). Use when reviewing a PR for security, auditing a service, or before a security-sensitive release.
---

# Security Reviewer

Reviews code for security issues before they reach production. **Not** a deep penetration test — that's a different discipline. This skill catches the issues that architect skills already encode rules against; it's the safety net.

## 1. When to invoke

- User asks "review this for security", "audit", "is this safe", "any security issues".
- Pre-release review of a service touching auth, payments, PII, or external integration.
- After dependency updates that include security advisories.
- New endpoint, new auth flow, new SQL query — anything where the failure mode is "data leaks" or "user gets owned".

## 2. Output format

A structured findings report — one row per finding.

```markdown
| Severity | Rule | Location | Evidence | Fix |
|---|---|---|---|---|
| Critical | SQL injection | `users/repo.py:42` | `f"SELECT ... {user_id}"` — string interpolation | Use parameter binding: `cur.execute("SELECT ... %s", (user_id,))` |
| High | Hardcoded secret | `config.go:18` | `apiKey := "sk-..."` | Move to env via `viper`; rotate the exposed key |
| Medium | Missing CORS allow-list | `main.py:34` | `allow_origins=["*"]` with credentials | Restrict to explicit origins per rest-api-architect §10 |
| Low | Outdated dep | `go.mod` | `golang.org/x/net v0.20.0` | Bump to current; Renovate will handle |
| Info | Defense in depth | `auth.go:88` | No rate limit on /login | Add per-IP rate limit; failures audited |
```

**Severity scale:**

- **Critical** — exploitable now, data loss / RCE / full account takeover possible.
- **High** — exploitable with realistic effort, partial compromise.
- **Medium** — exploitable in specific conditions, or significant defense gap.
- **Low** — best-practice gap, low impact alone but contributes to risk.
- **Info** — observation, hardening suggestion, defense in depth.

End the report with a one-sentence summary: *"3 Critical, 2 High, 5 Medium; not safe to ship until Critical/High addressed."*

## 3. Review approach

Two passes:

1. **Tool pass** — run static analyzers, then read their output. They catch most low-hanging fruit fast.
2. **Read pass** — read the actual diff (or files in scope), checking the categories in §4. Tools miss intent.

The read pass is where most real findings come from. Tools surface patterns; humans understand context.

## 4. What to check — by category

### Injection

- **SQL injection.** Per [sql-architect §4](../../databases/sql-architect/SKILL.md#4-query-patterns) and [§8](../../databases/sql-architect/SKILL.md#8-security): every query uses parameter binding. Grep for f-strings, format-strings, concatenation in SQL. `WHERE id = '${userId}'` is a fail; `WHERE id = $1` is a pass.
- **Command injection.** Any `subprocess.run(..., shell=True)`, `os.system`, `exec.Command("sh", "-c", userInput)`. Use list-form args.
- **Template injection.** `Jinja2(autoescape=False)` or `html/template` bypassed via `template.HTML(userInput)`.
- **LDAP / NoSQL / regex injection.** Less common but real — anything that builds a query from user input without escaping is a candidate.

### Auth & authz

- **Per [rest-api-architect §11](../../protocols/rest-api-architect/SKILL.md#11-authentication-patterns)** — Pattern A (in-house JWT) or B (external IdP). Reviewers verify the implementation matches.
- **Argon2id** for password hashing (not bcrypt, never MD5/SHA1).
- **JWT signing algorithm pinned** (`jwt.WithValidMethods([]string{"HS256"})` Go / `algorithms=["HS256"]` Python) — never `algorithms=["HS256", "none"]`, never accept the alg the token claims.
- **`aud` and `iss` verified** for external IdP tokens.
- **Tokens not in URLs**, only in `Authorization: Bearer` headers.
- **401 vs 403** distinction maintained.
- **No tenant-id from request body**; derive from authenticated context. `customer_id` in a JSON payload is a take-over vector.
- **Authorization checked per endpoint**, not assumed by middleware. Per [fastapi-architect §6](../../frameworks/fastapi-architect/SKILL.md#6-authentication--authorization--fastapi-implementation) / [gin-architect §7](../../frameworks/gin-architect/SKILL.md#7-authentication--authorization--gin-implementation) / [nethttp-architect §8](../../frameworks/nethttp-architect/SKILL.md#8-authentication--authorization--stdlib-implementation): scopes/roles enforced per-route.

### Secrets

- **No secrets in code.** Run `gitleaks` (already pinned via [repo-tooling-architect](../../tooling/repo-tooling-architect/STACK.md)) over the diff + the whole history. `gitleaks detect --staged`.
- **No secrets in env files committed to git.** Only `.env.local` (gitignored) holds secrets.
- **No secrets in Dockerfile layers.** Per [docker-architect §4](../../infra/docker-architect/SKILL.md#4-image-security): build-time secrets via `--mount=type=secret`; runtime via orchestrator.
- **No secrets in logs.** Per [observability-architect §7](../../infra/observability-architect/SKILL.md#7-what-not-to-emit): redaction filters configured.
- **Rotation.** If a secret was ever committed, rotate it — even if "we removed it later". Git history is forever.

### Insecure defaults

- **CORS:** never `Access-Control-Allow-Origin: *` for authenticated endpoints. Explicit allow-list.
- **HTTPS only.** Reject plaintext at the load balancer / `Strict-Transport-Security` header set.
- **Security headers** present: `X-Content-Type-Options: nosniff`, `Content-Security-Policy` (for HTML), `Referrer-Policy: no-referrer`.
- **Cookies:** `HttpOnly`, `Secure`, `SameSite=Lax` (or `Strict` where compatible).
- **Default-deny in IAM.** Roles grant specific permissions; nothing is "allow all" except the operator break-glass account.
- **Container runs as non-root** (per [docker-architect §1](../../infra/docker-architect/SKILL.md#1-dockerfile-fundamentals)).

### Deserialization and parsing

- **No `pickle.loads` on untrusted data.** Python's `pickle` is RCE-by-design.
- **No `yaml.load(s)` — use `yaml.safe_load(s)`.**
- **JSON: strict** — `json.NewDecoder(r.Body).DisallowUnknownFields()` (Go), `model_config = ConfigDict(extra="forbid")` (Pydantic). Per [fastapi-architect §3](../../frameworks/fastapi-architect/SKILL.md#3-pydantic-schemas--separate-request-and-response) / [nethttp-architect §3](../../frameworks/nethttp-architect/SKILL.md#3-request-validation).
- **XML: disable external entities** (XXE) — `defusedxml` (Python) or `xml.disable_external_entities` style configuration.

### CSRF, SSRF, IDOR

- **CSRF:** for cookie-authenticated browser endpoints, double-submit token or SameSite=Strict cookies. Bearer-token APIs don't need CSRF protection.
- **SSRF:** outbound HTTP from server-side code that takes a user-supplied URL must validate the URL — block private IP ranges (`10.0.0.0/8`, `192.168.0.0/16`, `127.0.0.0/8`, `169.254.169.254` AWS metadata), enforce HTTPS, restrict to allow-list of domains.
- **IDOR (insecure direct object reference):** every resource lookup checks ownership. `GET /v1/orders/{id}` returns 404 if the order isn't the caller's, *not* 403 (don't leak existence).

### Rate limiting and abuse

- **Rate limit at the edge** (gateway / CDN / WAF) plus per-endpoint critical-path limits.
- **Per-caller**, not per-IP — IPs are unreliable identity (per [rest-api-architect §14](../../protocols/rest-api-architect/SKILL.md#14-rate-limiting)).
- **Login endpoints rate-limited aggressively** to prevent credential stuffing. Lock the account after N failures within a window.
- **Account enumeration:** signup/login error messages don't disclose whether a user exists. "Invalid credentials" — not "user not found".

### Dependency hygiene

- **`trivy image` + `trivy fs`** in CI per [docker-architect §10](../../infra/docker-architect/SKILL.md#10-vulnerability-scanning--trivy). Fail on HIGH/CRITICAL.
- **Renovate-managed updates** per [repo-tooling-architect §7](../../tooling/repo-tooling-architect/SKILL.md#7-dependency-updates--renovate-default-dependabot-acceptable). Security updates land immediately.
- **Lockfiles committed.** `go.sum`, `uv.lock`, `package-lock.json` — never edit by hand.
- **No `:latest` tags** in Dockerfiles (per [docker-architect](../../infra/docker-architect/SKILL.md)). Digest-pin in production.

## 5. Tooling

Run these on the diff or service before the read pass. Findings go into the report alongside human-found issues.

| Tool | Language | Catches |
|---|---|---|
| `gitleaks` | All | Secrets in code and git history |
| `semgrep --config=auto` | All | Multi-language SAST patterns (1500+ rules out of the box) |
| `gosec` | Go | Go-specific issues (`G101`-`G505`) |
| `bandit` | Python | Python-specific issues (B101-B610) |
| `trivy fs .` | All | Filesystem CVE scan, license check, IaC scan |
| `npm audit` / `pip-audit` | Node / Python | Direct dependency advisories |

Per [repo-tooling-architect §5](../../tooling/repo-tooling-architect/SKILL.md#5-pre-commit-hooks--minimal-opt-in): keep these in CI, **not** in pre-commit hooks (they're slow). `gitleaks` is the exception — fast enough to belong in pre-commit.

## 6. What this skill does NOT do

- **Penetration testing.** Black-box attack simulation against a running service is a different skill set and requires explicit authorization.
- **Threat modeling.** Architecture-level "what attacker types can affect this system" lives in [improve-codebase-architecture](../../workflows/improve-codebase-architecture/SKILL.md) territory or a dedicated workshop.
- **Compliance audit.** SOC 2 / HIPAA / PCI-DSS specifics need a compliance specialist.

This skill stops at code- and configuration-level findings. Findings that imply deeper work flag that explicitly: *"recommendation: full threat-model session with the team that owns Payments."*

## 7. Cross-skill ties

- [rest-api-architect §10–11](../../protocols/rest-api-architect/SKILL.md#10-auth--security-headers) — REST security conventions reviewers verify against.
- [sql-architect §4 & §8](../../databases/sql-architect/SKILL.md#4-query-patterns) — parameter binding, RLS for multi-tenancy.
- [docker-architect §4 & §10](../../infra/docker-architect/SKILL.md#4-image-security) — image security baseline + Trivy scanning.
- [observability-architect §7](../../infra/observability-architect/SKILL.md#7-what-not-to-emit) — PII / secret redaction in signals.
- [repo-tooling-architect §5 & §7](../../tooling/repo-tooling-architect/SKILL.md#5-pre-commit-hooks--minimal-opt-in) — gitleaks in pre-commit, Renovate for security updates.
- [commit-author](../../workflows/commit-author/SKILL.md) — security fix commits get the `fix(scope): ...` type with `BREAKING CHANGE` footer if API behavior changes.
