# my-awesome-skills

**Note:** This is my personal toolkit for daily development, made public for anyone who wants to use, fork, or learn from it.

A growing collection of opinionated, Staff-level skills for OpenCode and Claude Code. 

By default, AI coding agents produce functional but generic code. They nest logic too deeply, ignore architectural patterns, and write verbose commit messages. These skills enforce strict standards—functioning as an automated Senior Engineer that demands clean, maintainable, and highly optimized code.

## Skill Library

The following specialized skills are currently available:

### Code Quality & Architecture

- **`logic-cleaner`**: Language-agnostic refactoring engine that flattens nested conditionals using guard clauses, simplifies boolean logic, and eliminates magic numbers.
- **`go-architect`**: Enforces strict Go standards including memory-aligned structs, typed enums (no raw strings), interface design philosophy, and concurrency safety.
- **`python-architect`**: Enforces modern Python 3.10+ standards with strict type hints, dataclass/Pydantic usage, and safe concurrency patterns.
- **`improve-codebase-architecture`**: Identifies architectural friction points and proposes deepening opportunities guided by domain language (CONTEXT.md) and architecture decisions (docs/adr/).

### Development Workflows

- **`commit-author`**: Generates concise, meaningful commit messages following the Conventional Commits specification from git diffs.
- **`tdd`**: Test-driven development workflow with red-green-refactor loop emphasizing behavior verification through public interfaces.
- **`caveman`**: Ultra-compressed communication mode that reduces token usage by ~75% while preserving technical accuracy.
- **`grill-me`**: Stress-tests plans and designs through relentless questioning until reaching shared understanding.

### Domain & Documentation

- **`ubiquitous-language`**: Extracts and formalizes DDD-style glossaries from conversations, identifying ambiguities and proposing canonical terminology. (From Matt Pocock)
- **`demo-script-architect`**: Designs presenter-centric demo scripts with narrative flow, visual guidance, and progressive capability reveals.

## Recommended Official Claude Code Skills

These official Claude Code skills complement the custom skills above and are highly recommended:

- **`frontend-design`**: Create distinctive, production-grade frontend interfaces with high design quality. Generates creative, polished code that avoids generic AI aesthetics.
- **`xlsx`**: Work with spreadsheet files (.xlsx, .xlsm, .csv, .tsv) for data manipulation, analysis, formatting, and conversion tasks.

## Attribution

Several skills in this repository are adapted from [Matt Pocock's skills](https://github.com/mattpocock/skills):
- `ubiquitous-language`
- `tdd`
- `caveman`
- `improve-codebase-architecture`

The original implementations have been customized and integrated into this toolkit.

## Installation

These skills are designed as drop-in modules for OpenCode and Claude Code.

### Global Installation

Clone this repository and copy the skill folders to your global skills directory.

**Windows:**
```powershell
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.config\opencode\skills\"
Copy-Item -Path .\skills\* -Destination "$env:USERPROFILE\.config\opencode\skills\" -Recurse -Force

New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.claude\skills\"
Copy-Item -Path .\skills\* -Destination "$env:USERPROFILE\.claude\skills\" -Recurse -Force