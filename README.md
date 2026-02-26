# my-awesome-skills 🧠🛠️

**Note:** *This is currently my personal toolkit for daily development, but I've made it public for anyone who wants to use, fork, or learn from it.*

An expanding collection of highly opinionated, Staff-level skills for **OpenCode** and **Claude Code**. 

By default, AI coding agents tend to generate functional but generic code (often referred to as "AI slop"). They nest logic too deeply, ignore memory alignment, and write verbose commit messages. 

These skills act as a strict, automated Senior Engineer—forcing the AI to write clean, maintainable, and highly optimized code. Because this is my evolving personal toolkit, new skills are actively being developed and added as my workflows change.

## 📦 Current Skill Library

This repository currently includes the following specialized skills:

- **`logic-cleaner`**: A language-agnostic, ruthless refactoring engine. Forces the AI to flatten nested `if` statements (using guard clauses), simplify boolean algebra, and eliminate magic numbers.
- **`go-architect`**: Strict standards for modern Go. Enforces memory alignment in structs, strict typing for enums (no raw strings), interface philosophy, and concurrency safety.
- **`python-architect`**: Modern Python 3.10+ standards. Enforces strict type hinting, `dataclasses`/Pydantic for memory optimization, and safe concurrency using Context Managers and TaskGroups.
- **`commit-author`**: A Git history cleaner. Reads the `git diff` and generates concise, meaningful commit messages strictly following the Conventional Commits specification.

## 🚀 Installation

These skills are designed to be drop-in ready for OpenCode (and are fully compatible with Claude Code).

### Global Installation (Recommended)
To make these skills available across all your projects, clone this repository and copy the folders into your global skills directory.

**Windows:**
```powershell
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.config\opencode\skills\"
Copy-Item -Path .\skills\* -Destination "$env:USERPROFILE\.config\opencode\skills\" -Recurse -Force

New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.claude\skills\"
Copy-Item -Path .\skills\* -Destination "$env:USERPROFILE\.claude\skills\" -Recurse -Force