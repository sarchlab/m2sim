# M2Sim Validation Baseline

This document captures the validation state of M2Sim before M3 (Timing Model) changes.

## Validation Date
2026-02-02

## Milestone Status
- **M1: Foundation (MVP)** ✅ Complete
- **M2: Memory & Control Flow** ✅ Complete
- **M3: Timing Model** 🚧 Not started (pending this validation)

## Test Suite Results

### Unit Tests Summary

| Package | Tests | Status |
|---------|-------|--------|
| emu | 188 | ✅ PASS |
| driver | 17 | ✅ PASS |
| insts | 46 | ✅ PASS |
| loader | 17 | ✅ PASS |
| timing/cache | 8 | ✅ PASS |
| timing/mem | 6 | ✅ PASS |
| timing/core | - | ⚠️ Build failed (WIP for M3) |
| timing/pipeline | - | ⚠️ Build failed (WIP for M3) |

**Total: 282 tests passing** (excluding timing WIP)

### Functional Emulator Capabilities

#### Supported Instructions
- **ALU (Immediate)**: ADD, ADDS, SUB, SUBS
- **ALU (Register)**: ADD, ADDS, SUB, SUBS, AND, ANDS, ORR, EOR
- **Branch**: B, BL, B.cond (all conditions), BR, BLR, RET
- **Load/Store**: LDR (32/64-bit), STR (32/64-bit)
- **System**: SVC (syscall)

#### Supported Syscalls
- **exit (93)**: Program termination with exit code
- **write (64)**: Write to file descriptor (stdout/stderr)

#### Known Limitations
1. N-bit not handled in logical register instructions (BIC, ORN, EON)
2. No shifted register operands for ALU instructions
3. No SIMD/floating-point support
4. Nested function calls require manual stack management (no automatic LR save)

## Ethan Validation Test Suite

The validation test suite is implemented in `emu/validation_test.go` and includes 11 test programs.

### Test Programs and Results

| Program | Description | Exit Code | Instructions | Status |
|---------|-------------|-----------|--------------|--------|
| simple_exit | exit(42) | 42 | 3 | ✅ PASS |
| arithmetic | 10 + 5 | 15 | 5 | ✅ PASS |
| subtraction | 100 - 58 | 42 | 4 | ✅ PASS |
| loop | count down 3→0 | 0 | 9 | ✅ PASS |
| loop_sum | 1+2+3+4+5 | 15 | 19 | ✅ PASS |
| hello | write "Hello\n" | 0 | 7 | ✅ PASS |
| function_call | BL/RET | 15 | 6 | ✅ PASS |
| factorial | 5! (placeholder) | 120 | 3 | ✅ PASS |
| logical_ops | AND/ORR/EOR | 240 | 8 | ✅ PASS |
| memory_ops | LDR/STR | 77 | 3 | ✅ PASS |
| chained_calls | sequential BL/RET | 35 | 9 | ✅ PASS |

### Running the Validation Suite

```bash
# Run all emu tests (includes validation suite)
go test ./emu/... -v

# Run only validation tests
go test ./emu/... -v -run "Ethan"

# Run full test suite (functional emulator only)
go test ./emu/... ./driver/... ./insts/... ./loader/... -v
```

## Regression Baseline

The following represents the expected behavior that MUST be preserved when M3 timing changes are integrated:

### Exit Codes
All test programs must produce identical exit codes.

### Output
Programs with write syscalls must produce identical output.

### Instruction Semantics
The functional behavior (register values, memory state) must remain unchanged.

### Instruction Count
While timing may change, instruction counts for a given program should remain the same
(the emulator executes the same instructions, just with timing information added).

## Notes for Timing Model Integration

When M3 timing changes are integrated:
1. All existing unit tests must continue to pass
2. All validation test programs must produce identical results
3. New timing-specific tests should be added
4. Performance metrics can differ (that's expected)
5. The timing/core and timing/pipeline packages need to be completed

## Sign-off

- [x] **Bob**: Validation infrastructure created (2026-02-02)
  - Created `emu/validation_test.go` with 11 test programs
  - All 282 tests passing in functional emulator packages
- [ ] **Ethan**: Full test validation (pending Ethan review)
- [ ] **Alice**: Approved for M3 timing work
