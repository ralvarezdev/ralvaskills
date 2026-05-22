# Build Errors - Immediate Action Required

## Summary
Your refactoring introduced undefined references. **23 compilation errors** need fixing.

## Critical Issues

### 1. **Missing: `DefaultConfigFilePath()`**
- **Files affected:** `cmd/init.go:157,158`
- **Error:** `undefined: DefaultConfigFilePath`
- **Fix:** Define this function in `cli/internal/config/fs.go`
  ```go
  func DefaultConfigFilePath() string {
      return filepath.Join(xdg.ConfigHome, ConfigFolderName, ConfigFileName)
  }
  ```

### 2. **Missing: `config.DefaultPath()`**
- **Files affected:** `cmd/init.go:44`
- **Error:** `undefined: config.DefaultPath`
- **Fix:** This was removed from config.go but is still called in cmd layer
  - Either restore it in config.go, OR
  - Update cmd/init.go to use `config.DefaultConfigFilePath()` instead

### 3. **Function Naming: `flagBool` vs `FlagBool`**
- **Files affected:** `cmd/init.go:34`, `cmd/install.go:39,40,41`
- **Error:** `undefined: flagBool (but have FlagBool)`
- **Fix:** Update cmd/flags.go or the function names to be consistent
  - Check if function is exported (FlagBool) but being called lowercase (flagBool)
  - Rename calls from `flagBool()` → `FlagBool()` or vice versa

### 4. **Type Moved: `manifest.ToolID`**
- **Files affected:** `cmd/init.go:126`
- **Error:** `cannot use t.ID (variable of string type manifest.ToolID) as string value in map index`
- **Issue:** ToolID type was moved, but its usage in cmd/init.go wasn't updated
- **Fix:** Check where ToolID ended up (tool.go? manifest.go?)
  - Confirm the type definition location
  - Update imports if needed

### 5. **Missing Field: `ToolDef.id`**
- **Files affected:** `cmd/init.go:133`
- **Error:** `ToolDef has no field or method id`
- **Fix:** Check the ToolDef struct definition
  - Verify field name (should be `id`? `ID`? something else?)
  - Update struct definition or reference

### 6. **Unused Variable: `home`**
- **Files affected:** `cmd/init.go:148`
- **Error:** `declared and not used: home`
- **Fix:** Either use `home` variable or remove its declaration

---

## Files to Check

### `cli/internal/manifest/fs.go`
- [ ] Verify `ToolID` type exists (was it moved here?)
- [ ] Verify `ToolClaudeCode`, `ToolOpenCode` constants exist
- [ ] Verify `writeFileAtomic()` function exists (moved from mod.go)
- [ ] Add `DefaultConfigFilePath()` if missing

### `cli/internal/manifest/tool.go` (or wherever ToolID went)
- [ ] Confirm ToolID definition and its `MarshalText()`, `UnmarshalText()` methods
- [ ] Verify tool constant definitions

### `cli/cmd/init.go`
- [ ] Check all uses of `flagBool` vs `FlagBool`
- [ ] Check all uses of removed constants/types
- [ ] Remove unused `home` variable

### `cli/cmd/flags.go`
- [ ] Verify function naming is consistent

---

## Quick Fix Checklist

```bash
# Step 1: Find all undefined references
cd cli && go build ./... 2>&1 | grep "undefined:"

# Step 2: For each undefined reference, locate where it should be defined
grep -r "func DefaultConfigFilePath" cli/internal/
grep -r "type ToolID" cli/internal/
grep -r "func writeFileAtomic" cli/internal/

# Step 3: After moving/adding definitions, rebuild
cd cli && go build ./...

# Step 4: If all compiles, run tests
cd cli && go test ./...
```

---

## Root Cause Analysis

Your refactoring moved/removed these items from their original locations without updating all references:

| Item | Removed From | Moved To | Cmd Layer Still Uses? |
|------|---|---|---|
| `DefaultPath()` | config.go | ❌ N/A | ✅ YES (init.go:44) |
| `DefaultConfigFilePath()` | ❌ Was never defined | Should be fs.go | ✅ YES (init.go:157,158) |
| `ToolID` type | manifest/mod.go | ❓ tool.go? fs.go? | ✅ YES (init.go:126) |
| `writeFileAtomic()` | manifest/util.go (or mod.go) | ❓ Exists? | ✅ YES (internal use) |

---

## Suggested Order of Fixes

1. **First:** Define `DefaultConfigFilePath()` in config/fs.go
2. **Second:** Fix ToolID location and update imports
3. **Third:** Fix flagBool/FlagBool naming
4. **Fourth:** Verify all moved functions exist in their new locations
5. **Fifth:** Run full build and test suite

---

## After Fixing

Once build succeeds:
```bash
# Full test
go test ./...

# Format check
go fmt ./...

# Lint check
golangci-lint run ./...
```
