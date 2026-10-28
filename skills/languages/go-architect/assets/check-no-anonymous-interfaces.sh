#!/bin/bash
set -euo pipefail

# Check for anonymous interface types: an interface written inline
# (x.(interface{ Policy() Policy }), a parameter, a variable) instead of
# declared with `type Name interface`. An anonymous one has no name to document,
# reuse or grep, and the capability it describes is invisible to the package's
# API.
#
# The walker parses the AST, so a multi-line declaration is not mistaken for an
# inline one, and it skips generated code (a "Code generated" marker, .pb.go
# files, gen/ and vendor/). Empty interfaces (interface{} / any) are not
# reported: any is the right spelling for a truly unconstrained value.
#
# go-architect asset — wire into your task runner as an opt-in check, not the
# main lint gate: golangci-lint has no linter for this (see SKILL.md §4, §14).
#
# Usage: check-no-anonymous-interfaces.sh [-tests] [dir]

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

echo "Checking for anonymous interfaces..."
exec go run "${script_dir}/check-no-anonymous-interfaces/main.go" "$@"
