---
name: commit-author
description: Analyzes code diffs and generates concise, meaningful Git commit messages following the Conventional Commits specification.
---
# Git Commit Standards

## 1. Structure & Format
- Conventional Commits: Every commit message must strictly follow this format: <type>(<optional scope>): <subject>.
- Allowed Types: - feat: A new feature or capability.
  - fix: A bug fix.
  - refactor: Code changes that neither fix a bug nor add a feature (e.g., renaming variables, simplifying logic).
  - docs: Documentation changes only.
  - chore: Maintenance tasks, dependency updates, or build process changes.
  - perf: A code change that improves performance.
  - test: Adding missing tests or correcting existing tests.

## 2. The Subject Line
- Imperative Mood: Write the subject line as a command. Use add, change, fix, remove instead of added, changing, fixed, or removes.
- Length Limit: The subject line must be 50 characters or less.
- Formatting: Do not capitalize the first letter. Do not end the subject line with a period.

## 3. The Body (The "Why")
- Focus on Intent: The git diff already shows *what* changed. The body of the commit message must explain *why* the change was made, the problem it solves, or the architectural reasoning behind it.
- Line Wrapping: Hard wrap all lines in the message body at 72 characters.
- No Line-by-Line Summaries: Never write a bulleted list of every file that was modified. Summarize the overarching logical change.

## 4. Footers & Breaking Changes
- Breaking Changes: If the commit breaks backward compatibility (like modifying a public API contract, removing an exported function, or changing a database schema), the footer MUST begin with BREAKING CHANGE: followed by a clear explanation and migration path.
- References: Note any related issues, tickets, or pull requests at the very bottom (e.g., Resolves #12).