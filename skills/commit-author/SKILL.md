---
name: commit-author
description: Analyzes diffs to generate concise Conventional Commits messages.
---

# Git Commit Standards

## 1. Format & Types
* **Structure:** `<type>(<optional scope>): <subject>`
* **Allowed Types:** `feat`, `fix`, `refactor`, `docs`, `chore`, `perf`, `test`.

## 2. Subject Line
* **Rules:** Use imperative mood (e.g., `add`, not `added`), maximum 50 characters, lowercase start, and no trailing period.

## 3. Body
* **Content:** Explain *why* the change was made, not *what* changed. Do not write line-by-line file summaries.
* **Formatting:** Hard wrap lines at 72 characters.

## 4. Footers
* **Breaking Changes:** Must begin with `BREAKING CHANGE:` followed by a migration path.
* **References:** Append related issues at the bottom (e.g., `Resolves #12`).

## 5. Execution Protocol
* **Consent:** Ask if the user will commit manually or delegate to you. **Never commit without explicit permission.**
* **Signing:** Delegated commits must be GPG signed if configured, unless the user explicitly opts out.