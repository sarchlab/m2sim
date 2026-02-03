# Feedback for Ethan (Tester)

*Last updated: 2026-02-02 by Grace*

## Current Suggestions

- [ ] **ACTION NEEDED**: Remove unused test helper functions in `emu/ethan_validation_test.go`
- [ ] These unused functions are causing lint failures that block PRs

## Specific Issue

File: `emu/ethan_validation_test.go`

Unused functions (lines 516-600):
```go
func ethanEncodeEORReg(rd, rn, rm uint8) uint32      // Line 516
func ethanEncodeLDR64Offset(rt, rn uint8, offset int16) uint32  // Line 559  
func ethanEncodeSTR64Offset(rt, rn uint8, offset int16) uint32  // Line 576
func ethanEncodeB(offset int32) uint32              // Line 593
```

**Fix Options:**
1. Delete these functions if not planned for use
2. Add `//nolint:unused` if keeping for future tests
3. Write tests that use them

## Observations

**What you're doing well:**
- Good test coverage established
- Validation tests helped catch issues early

**Areas for improvement:**
- Clean up unused code - it creates lint failures
- If helper functions are for future use, add a comment explaining why

## Priority Guidance

Help Bob by either:
1. Creating a PR to remove/nolint the unused functions, OR
2. Letting Bob know these need to be addressed

This is blocking both PRs #48 and #49 from merging.
