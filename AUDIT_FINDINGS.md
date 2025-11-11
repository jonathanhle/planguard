# Planguard Audit Findings

## 🔴 CRITICAL ISSUES (Must Fix Before v1.0)

### 1. GitHub Action Outputs Don't Work
**Location**: `action.yml` + `entrypoint.sh`
**Problem**: Outputs `violations` and `passed` are declared but never set.
**Impact**: Users following ACTION.md examples will get empty values.
**Fix Required**: Modify `entrypoint.sh` to write to `$GITHUB_OUTPUT`

```sh
# Add after planguard runs:
if [ "$EXIT_CODE" -eq 0 ]; then
  echo "passed=true" >> $GITHUB_OUTPUT
else
  echo "passed=false" >> $GITHUB_OUTPUT
fi

# For JSON format, extract violations
if echo "$@" | grep -q "format.*json"; then
  echo "violations=$OUTPUT" >> $GITHUB_OUTPUT
fi
```

### 2. Stderr Not Captured
**Location**: `entrypoint.sh:4`
**Problem**: Only stdout is redirected, stderr messages are lost.
**Impact**: Error messages won't appear in logs or output files.
**Fix**: Change to `> /tmp/planguard-output.txt 2>&1`

### 3. Release Workflow Continues on Existing Tag
**Location**: `.github/workflows/release.yml:41-74`
**Problem**: If tag exists, workflow skips creation but runs GoReleaser anyway, which fails.
**Impact**: Failed releases, confusion.
**Fix**: Add early exit when tag exists.

### 4. README False Claim - Azure Support
**Location**: `README.md:68`
**Problem**: Says "AWS, Azure, and common patterns" but only AWS rules exist.
**Impact**: Users expect Azure rules, won't find them.
**Fix**: Change to "AWS and common patterns"

### 5. README Overselling - "Production-ready"
**Location**: `README.md:3`
**Problem**: Claims production-ready but has critical bugs (outputs, stderr).
**Impact**: Users may use in production with broken features.
**Fix**: Change to "Beta" or "v1.0 Release Candidate"

---

## 🟡 HIGH PRIORITY ISSUES

### 6. Unbenchmarked Performance Claim
**Location**: `README.md:12`
**Problem**: "Scans 1000+ files in seconds" - no benchmarks to support this.
**Impact**: Overselling, potential disappointment.
**Fix**: Either benchmark and prove it, or remove the specific number.

### 7. SARIF Detection Fragile
**Location**: `entrypoint.sh:11`
**Problem**: `grep -q "\-format sarif"` might miss variations.
**Impact**: SARIF file might not be created in edge cases.
**Fix**: Use `grep -q "format.*sarif"` for more robust matching.

### 8. Docker VERSION Build Arg Not Passed
**Location**: `.goreleaser.yml` Docker config
**Problem**: Dockerfile expects VERSION arg but GoReleaser doesn't pass it.
**Impact**: Docker images show version as "dev" instead of actual version.
**Fix**: Add `build_flag_templates: - "--build-arg=VERSION={{ .Version }}"`

---

## 🟢 NICE TO HAVE / POLISH

### 9. Config Default May Confuse Users
**Location**: `action.yml:13`
**Problem**: Default is `.planguard/config.hcl` which may not exist.
**Impact**: Users get error unless they create config.
**Consideration**: This might be intentional to force explicit configuration.
**Suggestion**: Document clearly in README that config is required OR make it truly optional.

### 10. Missing Error Handling for Large Output
**Location**: `entrypoint.sh:8`
**Problem**: No check if `/tmp/planguard-output.txt` is huge.
**Impact**: Could exhaust memory with cat on massive file.
**Fix**: Add size check or use streaming approach.

### 11. HOMEBREW_TAP_TOKEN Not Documented
**Location**: README.md
**Problem**: Release workflow requires this secret but README doesn't mention it.
**Impact**: Contributors won't know to set it up.
**Fix**: Add "For Maintainers" section documenting release process.

---

## 📊 DOCUMENTATION INACCURACIES

### ACTION.md Example 3
**Lines 83-84**: Shows using `violations` and `passed` outputs which don't work.
**Fix**: Either fix outputs or remove this example.

### ACTION.md Example 6
**Shows PR comment feature** that relies on broken outputs.
**Fix**: Mark as "Coming Soon" or implement outputs first.

---

## ✅ WHAT'S WORKING WELL

1. ✅ Multi-stage Docker build (efficient)
2. ✅ VERSION file-driven releases (good approach)
3. ✅ Comprehensive workflow coverage (CI, verify, release)
4. ✅ Good example documentation structure
5. ✅ SARIF support for GitHub Security (innovative)
6. ✅ Homebrew tap setup (good distribution)
7. ✅ Exception management system (well designed)

---

## 🎯 RECOMMENDED FIX ORDER

1. **Fix outputs** (entrypoint.sh + test)
2. **Fix stderr capture** (entrypoint.sh)
3. **Fix release workflow early exit** (release.yml)
4. **Remove "Azure" from README** (README.md)
5. **Change "Production-ready" to "Beta"** (README.md)
6. **Add VERSION build arg to GoReleaser** (.goreleaser.yml)
7. **Remove/update ACTION.md examples using outputs**
8. **Improve SARIF detection** (entrypoint.sh)
9. **Document HOMEBREW_TAP_TOKEN** (README.md)
10. **Remove or validate performance claim** (README.md)

---

## 🔬 TESTING RECOMMENDATIONS

1. **Test action outputs** - verify violations and passed are set correctly
2. **Test stderr capture** - ensure error messages appear in logs
3. **Test SARIF upload** - confirm file is created and uploaded
4. **Test with no config file** - document or fix behavior
5. **Test release workflow** - try with existing and new tags
6. **Benchmark performance** - if keeping the 1000+ files claim

---

## 🚀 PATH TO PRODUCTION-READY

To legitimately call this "production-ready":
- [ ] All critical issues fixed
- [ ] Comprehensive test suite (unit + integration)
- [ ] Benchmarks to validate performance claims
- [ ] Real-world usage by multiple users
- [ ] Security audit of HCL parsing
- [ ] Error handling for all edge cases
- [ ] Monitoring/telemetry (optional)

Currently at: **Alpha/Beta stage** - functional but needs polish.
