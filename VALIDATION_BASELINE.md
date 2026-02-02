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
| emu | 176 | ✅ PASS |
| driver | 17 | ✅ PASS |
| insts | 46 | ✅ PASS |
| loader | 17 | ✅ PASS |
| timing/cache | - | ✅ PASS |
| timing/mem | - | ✅ PASS |
| timing/core | - | ⚠️ Build failed (WIP) |
| timing/pipeline | - | ⚠️ WIP |

**Total: 256+ tests passing** (excluding timing WIP)

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

## Validation Test Programs

### 1. Simple Exit (benchmarks/simple_exit.s)
- **Purpose**: Verify basic program flow and exit syscall
- **Expected**: Exit code 42

### 2. Arithmetic Test (benchmarks/arithmetic.s)
- **Purpose**: Verify ALU operations
- **Expected**: Exit code 15 (10 + 5)

### 3. Loop Test (benchmarks/loop.s)
- **Purpose**: Verify conditional branches and loops
- **Expected**: Exit code 0 (count down from 3)

### 4. Hello World (benchmarks/hello.s)
- **Purpose**: Verify write syscall and output
- **Expected**: Output "Hello\n", exit code 0

### 5. Factorial (benchmarks/factorial.s)
- **Purpose**: Verify complex control flow and BL/RET
- **Expected**: Exit code 120 (5!)

## Regression Baseline

The following test programs establish the regression baseline:

```
Program          | Instructions | Exit Code | Output
-----------------|--------------|-----------|--------
simple_exit      | 3            | 42        | -
arithmetic       | 5            | 15        | -
loop             | 11           | 0         | -
hello            | 7            | 0         | "Hello\n"
factorial        | ~30          | 120       | -
```

## Ethan Validation Checklist

- [ ] Run all unit tests: `go test ./emu/... ./driver/... ./insts/... ./loader/...`
- [ ] Run benchmark programs through emulator
- [ ] Verify exit codes match expected values
- [ ] Verify stdout output where applicable
- [ ] Compare instruction counts with real ARM64 execution (where possible)
- [ ] Document any discrepancies

## Notes for Timing Model Integration

When M3 timing changes are integrated:
1. All existing unit tests must continue to pass
2. All benchmark programs must produce identical results
3. New timing-specific tests should be added
4. Performance metrics can differ (that's expected)

## Sign-off

- [ ] **Bob**: Validation infrastructure created
- [ ] **Ethan**: Test programs validated
- [ ] **Alice**: Approved for M3 timing work
