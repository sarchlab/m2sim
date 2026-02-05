# M2Sim Progress Report

**Last updated:** 2026-02-04 22:30 EST (Cycle 199)

## Current Status

| Metric | Value |
|--------|-------|
| Total PRs Merged | 44 |
| Open PRs | 1 |
| Open Issues | 15 |
| Pipeline Coverage | 75.9% |

## 🎉 Major Breakthrough: Primecount Fixed!

### Root Cause Found (Cycle 199)
Bob discovered that `executeDPReg` was ignoring `ShiftType`/`ShiftAmount` for shifted register instructions like `ADD x7, x8, x9, LSL #2`.

**Impact:** All benchmarks using shifted register operations were affected.

**Fix:** Added `applyShift64`/`applyShift32` helpers, now correctly applies shifts before operations.

### Results
| Metric | Before | After |
|--------|--------|-------|
| Instructions | 256 | 2,835,622 |
| Primes counted | 4 | **3,512** ✅ |

## Embench Phase 1 — Complete! ✅

| Benchmark | Instructions | Exit Code | Status |
|-----------|-------------|-----------|--------|
| aha-mont64 | 1.88M | 0 ✓ | ✅ Complete |
| crc32 | 1.57M | 0 ✓ | ✅ Complete |
| matmult-int | 3.85M | 0 ✓ | ✅ Complete |

## Embench Phase 2 — In Progress

| Issue | Benchmark | Status |
|-------|-----------|--------|
| #184 | primecount | PR #188 **FIXED** ✅ |
| #185 | edn | Ready for Bob |
| #186 | huffbench | Ready for Bob |
| #187 | statemate | Ready for Bob |

## Accuracy Status (Microbenchmarks)

From Eric's analysis (Cycle 199):

| Benchmark | Simulator CPI | M2 Real CPI | Error |
|-----------|---------------|-------------|-------|
| arithmetic | 0.400 | 0.268 | 49.3% |
| dependency | 1.200 | 1.009 | 18.9% |
| branch | 1.800 | 1.190 | 51.3% |
| **Average** | — | — | **39.8%** |

**Target:** <20% average error (interim), <2% (final)

## Active Work

### PR #188 — Primecount Benchmark (Bob)
- **Status:** Cathy-approved, CI pending
- **Contains:** Shifted register fix + LDRSW + SBFIZ fixes
- **Ready for merge** once CI passes

### #122 — Pipeline Refactor (Cathy)
- **Branch:** `cathy/122-pipeline-refactor-writeback`
- **Status:** WritebackSlot interface added

## Recent Progress

### Cycle 199 (Current) 🎉
- **Bob found and fixed shifted register bug** — major breakthrough!
- **Eric created accuracy report** — 39.8% average error baseline
- **Human suggested (#189):** Bob on benchmarks, Eric on accuracy

### Cycle 198
- Bob added LDRSW + SBFIZ fixes (partial)
- Primecount still broken (now fixed!)

### Prior
- PR #182 merged — exit code fix
- Phase 1 Embench complete

## Calibration Milestones

| Milestone | Status | Description |
|-----------|--------|-------------|
| C1 | 🎉 **COMPLETE** | Phase 1 Embench + CoreMark execute |
| C1.5 | **UNBLOCKED** | Phase 2 — primecount fixed! |
| C2 | In Progress | Microbenchmark Accuracy — <20% avg error |
| C3 | Pending | Intermediate Benchmark Accuracy |
| C4 | Pending | SPEC Benchmark Accuracy |

## Next Steps

1. **Merge PR #188** — once CI passes
2. **Continue Phase 2** — edn, huffbench, statemate
3. **Start accuracy tuning** — Eric's calibration work
4. **Test other benchmarks** — shifted register fix may help
