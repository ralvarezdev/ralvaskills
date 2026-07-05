#!/bin/bash
set -euo pipefail

# Check for decorative separator-style comments (// ──, // ═══, # ──, /* ──, etc.).
# These add visual noise without conveying information — prefer a blank line
# between logical groups, or a short comment stating the group's purpose.
#
# go-architect asset — wire into your task runner as an opt-in check, not the
# main lint gate: golangci-lint has no linter for this (see SKILL.md §14).

EXIT_CODE=0
PATTERN='(//|#|/\*)\s*[─═╔╗╚╝║╠╣╦╩╬╪╫╬─═]*[\s─═]*$'

echo "Checking for separator comments..."

for ext in go py sh js ts tsx jsx; do
  while IFS= read -r line; do
    if [[ -n "$line" ]]; then
      echo "Found separator comment: $line"
      EXIT_CODE=1
    fi
  done < <(grep -rn "$PATTERN" . \
    --include="*.$ext" \
    --exclude-dir="vendor" \
    --exclude-dir="third_party" \
    --exclude-dir="node_modules" \
    --exclude-dir=".git" \
    2>/dev/null || true)
done

if [[ $EXIT_CODE -eq 0 ]]; then
  echo "No separator comments detected"
else
  echo ""
  echo "Fix: remove decorative separators and keep only the content/purpose."
  echo "Use plain blank lines between logical groups instead."
fi

exit $EXIT_CODE
