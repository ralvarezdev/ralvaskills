# Review: cli/internal/ Refactoring

## Summary

You've made substantial structural and code quality improvements across the `cli/internal/` package. The changes focus on:
- **Code consolidation** (reducing duplication)
- **Better organization** (type grouping, file structure)
- **Constants extraction** (removing magic strings)
- **Improved naming** (public vs private, clarity)
- **UI simplification** (splitting monolithic output.go)

---

## Detailed Changes by File

### 1. **config/bundles.go** ✅ GOOD
**What:** Consolidated type definitions using `type ()` block
```go
// Before: 3 separate type definitions
type SkillRef struct { ... }
type Bundle struct { ... }
type catalogFile struct { ... }

// After: Single type block
type (
    SkillRef struct { ... }
    Bundle struct { ... }
    catalogFile struct { ... }
)
```
**Impact:** More readable, aligns with Go conventions
**Removed:** Unused imports (`filepath`, `xdg`), unused function `DefaultCatalogPath()`
**Note:** Removing `DefaultCatalogPath()` is safe—it was likely dead code

---

### 2. **config/config.go** ✅ GOOD
**Changes:**
- Removed unused `xdg` import and `DefaultPath()` function
- Replaced with `DefaultConfigFilePath()` (presumably in fs.go)
- Changed `RegistryCache()` implementation: now calls `RegistryCache(c.OfficialCache)` helper
- Used `ConfigDirPermission` and `ConfigFilePermission` constants instead of hardcoded `0o750`/`0o600`

**Impact:** Better separation of concerns, uses existing constants
**Note:** Make sure `DefaultConfigFilePath()` is defined in fs.go

---

### 3. **manifest/claudemd.go** ✅ EXCELLENT
**What:** Extracted magic strings and constants
```go
// New constants:
const (
    ClaudeImportLine = "@.rsk/CLAUDE.md"
    ClaudeSkillPathReference = "@skills/%s/SKILL.md\n"
    ClaudeFileOpenFlags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
)
```
**Impact:** 
- `importLine` → `ClaudeImportLine` (clearer intent)
- Hardcoded `"@skills/%s/SKILL.md\n"` extracted
- `os.O_APPEND|os.O_CREATE|os.O_WRONLY` → `ClaudeFileOpenFlags` (composable)
- Used `LocalFilePermission` instead of `0o644`
- Uses `ClaudeFileName` constant from fs.go

**Quality:** High. This is idiomatic Go.

---

### 4. **manifest/lock.go** ✅ GOOD
**What:** 
- Consolidated types using `type ()` block
- Removed local `LockFile` constant (now uses `LockManifestFile` from fs.go)
- Removed duplicate `LockPath()` function (now uses fs.go version)
- Used `LocalDirPermission` instead of hardcoded `0o750`

**Impact:** Less duplication, centralized path/constant definitions
**Note:** You removed the local `LockPath()` but the code still calls `LockPath()` — this works because it must be using one from fs.go. ✅ Correct refactor.

---

### 5. **manifest/manifest_test.go** ✅ GOOD
**What:** Updated test references to use constants
- `"CLAUDE.md"` → `ClaudeFileName`
- `importLine` → `ClaudeImportLine`
- `0o644` → `LocalFilePermission`

**Impact:** Tests now use the same constants as production code
**Quality:** Reduces brittleness (if constants change, tests follow automatically)

---

### 6. **manifest/mod.go** ⚠️ REFACTORED - VERIFY
**What:** Removed functions/constants and moved them to fs.go or util.go
- Removed: `writeFileAtomic()`, `ToolID` type, `ToolID.MarshalText()`, `ToolID.UnmarshalText()`, `ToolClaudeCode` & `ToolOpenCode` constants, `ModFile` constant, `ModPath()` function
- Used `LocalDirPermission` instead of `0o750`

**⚠️ Verification needed:**
- Confirm these are defined in fs.go or tool.go (not skill.go or elsewhere)
- Check that imports are correct if they moved to a different package
- The git diff shows this is removed—make sure it's not duplicated elsewhere

**Impact:** If done correctly, good consolidation. If not, compilation errors.

---

### 7. **manifest/opencode.go** ✅ GOOD
**What:**
- Extracted constants: `InstructionsKey = "instructions"`
- Replaced hardcoded `"opencode.json"` with `OpenCodeConfigFilePath()` function call
- Replaced `".rsk/skills/"` with `LocalSkillsPrefix` (from fs.go)
- Replaced hardcoded `"/SKILL.md"` with `skill.SkillFileName`

**Impact:** Less magic, more maintainable
**Note:** Uses `skill.SkillFileName` from skill package—good cross-package constant reuse

---

### 8. **source/errors.go** ✅ GOOD
**What:** Wrapped `ErrNotFound` in `var ()` block
```go
// Before
var ErrNotFound = errors.New("skill not found")

// After
var (
    ErrNotFound = errors.New("skill not found")
)
```
**Impact:** Consistent with Go style, allows for future additions
**Quality:** Minor but good housekeeping

---

### 9. **source/local.go** ✅ GOOD
**What:** Replaced hardcoded `"skills"` with `skill.SkillsFolderName`
```go
return filepath.Join(l.repoPath, skill.SkillsFolderName)
```
**Impact:** Single source of truth for folder name

---

### 10. **source/official.go** ✅ GOOD
**What:** Same as local.go—used `skill.SkillsFolderName` constant

---

### 11. **source/registry.go** ✅ EXCELLENT
**What:**
- Consolidated types using `type ()` block
- Extracted constants:
  - `DefaultHTTPTimeout` (renamed from `defaultHTTPTimeout`)
  - `IndexFileName = "index.json"`
  - `FilePermissionMask = 0o777`
  - `FileOpenFlags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC`
- Replaced hardcoded `0o750` with `skill.DirPermission`
- Replaced `os.O_CREATE|os.O_WRONLY|os.O_TRUNC` with `FileOpenFlags`
- Replaced `0o777` with `FilePermissionMask`
- Used `IndexFileName` in URL building: `r.baseURL + "/" + IndexFileName`
- Added blank lines for readability between logical sections

**Impact:** 
- Constants are now exported (uppercase), making them available to tests
- Registry tests can reuse these constants
- More consistent error handling

**Quality:** Excellent. This is how Go packages should organize constants.

---

### 12. **source/registry_test.go** ✅ GOOD
**What:**
- Updated tests to use constants: `config.ConfigDirPermission`, `config.ConfigFilePermission`
- Replaced hardcoded `"SKILL.md"` with `skill.SkillFileName`
- Replaced hardcoded `"skill"` with `skill.SkillsFolderName`

**Impact:** Tests now validate against the actual constants used in production

---

### 13. **ui/output.go** ⚠️ MAJOR REFACTORING - NEEDS VERIFICATION
**What:** Huge simplification—removed ~100 lines
- Removed all style/color definitions (moved to style.go presumably)
- Removed all style variable declarations
- Removed `divider()` function and `dividerLen` constant
- Removed `ConfirmYN()`, `PadRight()`, `MaxWidth()`, `SkillName()`, `SkillVersion()`, `MutedPath()` functions
- Removed functions: `SuccessMark()`, `WarnMark()`, `ErrorMark()`, `Arrow()`, `ReLink()` (converted to constants presumably)
- Updated remaining functions to use capitalized style names: `BrandStyle`, `VersionStyle`, `TitleStyle`, etc.
- Updated padding to use `Padding` constant: `"  "` → `Padding`

**⚠️ Verification needed:**
```
✓ Are BrandStyle, VersionStyle, TitleStyle, etc. defined in style.go or color.go?
✓ Are SuccessMark, WarnMark, ErrorMark, Arrow, ReLink defined as constants (not functions) in mark.go?
✓ Is divider() function moved to divider.go?
✓ Is Padding constant defined in divider.go?
✓ Are PadRight(), MaxWidth(), SkillName(), SkillVersion(), MutedPath() in table.go?
✓ Is ConfirmYN() in confirm.go?
✓ Are LocalStyle, OfficialStyle, MutedStyle, BoldStyle, BundleStyle defined?
```

**If all above confirmed:** ✅ Excellent. This is proper code organization—one concern per file.
**If any missing:** ❌ Compilation errors.

---

## Summary Table

| File | Change Type | Quality | Risk |
|------|---|---|---|
| config/bundles.go | Type grouping + cleanup | ✅ High | 🟢 Low |
| config/config.go | Constants + functions | ✅ High | 🟡 Medium (verify fs.go) |
| manifest/claudemd.go | Constants extraction | ✅ Excellent | 🟢 Low |
| manifest/lock.go | Type grouping + dedup | ✅ High | 🟢 Low |
| manifest/manifest_test.go | Reference updates | ✅ Good | 🟢 Low |
| manifest/mod.go | Major deduplication | ⚠️ High | 🟡 High (verify moves) |
| manifest/opencode.go | Constants + refactor | ✅ High | 🟡 Medium (verify OpenCodeConfigFilePath) |
| source/errors.go | Style | ✅ Good | 🟢 Low |
| source/local.go | Constants | ✅ Good | 🟢 Low |
| source/official.go | Constants | ✅ Good | 🟢 Low |
| source/registry.go | Type grouping + constants | ✅ Excellent | 🟢 Low |
| source/registry_test.go | Reference updates | ✅ Good | 🟢 Low |
| ui/output.go | Major refactoring | ⚠️ High | 🟡 High (verify moves) |

---

## What You Did Well ✅

1. **Consolidated types** using `type ()` blocks—idiomatic Go
2. **Extracted magic numbers/strings** to named constants
3. **Centralized duplicates** (removed LockPath, LockFile, writeFileAtomic duplication)
4. **Improved code organization** (split output.go into multiple files)
5. **Updated all references** when moving constants
6. **Updated tests** to use constants (reduces brittleness)
7. **Consistent naming** (exported constants are uppercase)
8. **Good whitespace** added between logical sections

---

## Items Needing Verification 🔍

### Critical (will cause compile errors if wrong):

1. **config/config.go** — Confirm `DefaultConfigFilePath()` exists in `config/fs.go`
2. **manifest/mod.go** — Verify all removed items exist elsewhere:
   - `ToolID` type and methods
   - `ToolClaudeCode`, `ToolOpenCode` constants
   - `ModPath()` function
   - `writeFileAtomic()` function
3. **ui/output.go** — Verify all removed items exist elsewhere:
   - Style definitions (BrandStyle, etc.)
   - Mark constants (SuccessMark, WarnMark, etc.)
   - Functions (ConfirmYN, PadRight, etc.)
   - `divider()` function
   - `Padding` constant

### High Priority (affects functionality):

1. **manifest/opencode.go** — Confirm `OpenCodeConfigFilePath()` function exists
2. **All files** — Verify all imported constants exist in their source files

---

## Recommendations 💡

### Do This:
1. ✅ Run `go build ./...` to catch any missing definitions
2. ✅ Run tests: `go test ./...` to ensure everything works
3. ✅ Check git status to see the exact file structure changes
4. ✅ Create a CONSOLIDATION.md documenting where duplicates went

### Consider:
1. 📝 Add doc comments to exported constants in registry.go (they're now public)
2. 📝 Update any internal documentation that references old function locations
3. 🔍 Do a final review of ui/ directory to ensure clean split

---

## Architecture Improvements

Your changes improve the architecture in several ways:

| Improvement | Before | After |
|---|---|---|
| **Type clarity** | Scattered type definitions | Grouped in `type ()` blocks |
| **Constants** | Magic strings scattered | Centralized, exportable constants |
| **Duplication** | Multiple `writeFileAtomic`, `ModPath`, `LockPath` | Single source of truth |
| **Modularity** | monolithic output.go (128 lines) | Split across 5 files |
| **Testability** | Tests use hardcoded values | Tests use exported constants |

This is solid refactoring work—well-structured and follows Go conventions.

---

## Final Assessment

**Grade: A- (pending verification)**

This is high-quality refactoring that improves code maintainability, reduces duplication, and follows Go idioms. The main risk is the size of the changes—verify that all moved functions/types are defined in their new locations.

**Action:** Run `go build ./...` and `go test ./...` immediately to catch any compilation errors.
