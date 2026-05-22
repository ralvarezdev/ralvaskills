# Security Reviewer — Reference Tables

Findings format, severity rubric, and tooling reference for [SKILL.md](SKILL.md). Loaded on demand.

## 1. Findings report format

```markdown
| Severity | Rule | Location | Evidence | Fix |
|---|---|---|---|---|
| Critical | SQL injection | `users/repo.py:42` | `f"SELECT ... {user_id}"` — string interpolation | Use parameter binding: `cur.execute("SELECT ... %s", (user_id,))` |
| High | Hardcoded secret | `config.go:18` | `apiKey := "sk-..."` | Move to env via `viper`; rotate the exposed key |
| Medium | Missing CORS allow-list | `main.py:34` | `allow_origins=["*"]` with credentials | Restrict to explicit origins per rest-api-architect §10 |
| Low | Outdated dep | `go.mod` | `golang.org/x/net v0.20.0` | Bump to current; Renovate will handle |
| Info | Defense in depth | `auth.go:88` | No rate limit on /login | Add per-IP rate limit; failures audited |
```

End with a one-sentence summary: *"3 Critical, 2 High, 5 Medium; not safe to ship until Critical/High addressed."*

## 2. Severity rubric

- **Critical** — exploitable now, data loss / RCE / full account takeover possible.
- **High** — exploitable with realistic effort, partial compromise.
- **Medium** — exploitable in specific conditions, or significant defense gap.
- **Low** — best-practice gap, low impact alone but contributes to risk.
- **Info** — observation, hardening suggestion, defense in depth.

## 3. Tooling reference

| Tool | Language | Catches |
|---|---|---|
| `gitleaks` | All | Secrets in code and git history |
| `semgrep --config=auto` | All | Multi-language SAST patterns (1500+ rules out of the box) |
| `gosec` | Go | Go-specific issues (`G101`-`G505`) |
| `bandit` | Python | Python-specific issues (B101-B610) |
| `trivy fs .` | All | Filesystem CVE scan, license check, IaC scan |
| `npm audit` / `pip-audit` | Node / Python | Direct dependency advisories |

- Run in CI, **not** in pre-commit hooks (most are too slow). `gitleaks` is the exception — fast enough for pre-commit.
- Per [repo-tooling-architect §5](../../tooling/repo-tooling-architect/SKILL.md#5-pre-commit-hooks--minimal-opt-in).
