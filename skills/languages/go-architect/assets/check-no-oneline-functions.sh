#!/bin/bash
set -euo pipefail

# Check for Go function/method declarations whose entire body sits on one line.
# Pattern: func ...(...) ... { <non-empty body> }
# Empty bodies (func Foo() {}) are allowed — this only catches one-liners with logic.
#
# go-architect asset — wire into your task runner as an opt-in check, not the
# main lint gate: golangci-lint has no linter for this, and existing codebases
# may have violations that need a separate cleanup pass (see SKILL.md §14).

EXIT_CODE=0
PATTERN='^[[:space:]]*func\b[^{]*\{[^}]+\}[[:space:]]*$'

echo "Checking for one-line function bodies..."

while IFS= read -r match; do
  if [[ -n "$match" ]]; then
    echo "One-line function: $match"
    EXIT_CODE=1
  fi
done < <(grep -rnE "$PATTERN" . \
  --include="*.go" \
  --exclude="*.pb.go" \
  --exclude="mock_*.go" \
  --exclude-dir="vendor" \
  --exclude-dir="third_party" \
  --exclude-dir=".git" \
  --exclude-dir=".claude" \
  2>/dev/null || true)

if [[ $EXIT_CODE -eq 0 ]]; then
  echo "No one-line function bodies detected"
else
  echo ""
  echo "Fix: split the function body onto its own line(s)."
fi

exit $EXIT_CODE
