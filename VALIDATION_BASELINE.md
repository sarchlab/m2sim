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

## Validation Test Programs (benchmarks/validation_test.go)

14 programmatic test cases covering all supported functionality:

| Test | Description | Expected |
|------|-------------|----------|
| Simple Exit | Basic exit(42) | Exit 42 |
| Arithmetic | 10 + 5 | Exit 15 |
| Subtraction | 100 - 58 | Exit 42 |
| Loop | Count down 3→0 | Exit 0, 9 instructions |
| Hello World | write(1, "Hello\n", 6) | Exit 0, stdout="Hello\n" |
| Iterative Sum | 5+4+3+2+1 | Exit 15 |
| Bitwise AND | 0xFF & 0x0F | Exit 0x0F |
| Bitwise ORR | 0xF0 \| 0x0F | Exit 0xFF |
| Bitwise EOR | 0xFF ^ 0xF0 | Exit 0x0F |
| Load/Store | STR then LDR 64-bit | Exit 123 |
| Branch with Link | BL/RET subroutine | Exit 15 |
| Conditional EQ | Branch when equal | Exit 5 (not 99) |
| Conditional GT | Branch when greater | Exit 10 (not 99) |
| Conditional Not Taken | No branch on false | Exit 42 |

## Regression Baseline

Run validation tests:
```bash
go test ./benchmarks/... -v
```

Expected: **14 specs, all passing**

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
