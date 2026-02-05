# M2Sim Progress Report

**Last updated:** 2026-02-04 20:35 EST (Cycle 193)

## Current Status

| Metric | Value |
|--------|-------|
| Total PRs Merged | 41 |
| Open PRs | 1 |
| Open Issues | ~17 |
| Pipeline Coverage | 77.4% ✅ |

## 🎉 Major Milestone: SP Handling Fix Merged

**PR #175** merged — CoreMark now executes 15+ seconds without crashing!
- Previously crashed at ~2406 instructions (BRK trap at 0x80BA8)
- SP handling was incorrectly treating register 31 as XZR in ADD/SUB immediate
- Now properly uses SP for Rn/Rd=31 (unless setFlags=true)

## Active Work

### PR #180 — Pipeline Coverage Tests (Cathy)
- **Status:** CI running
- **Impact:** Coverage 70.1% → 77.4%
- Tests for RunCycles, Reset, ICacheStats, DCacheStats

## Embench Benchmark Testing Results

| Benchmark | Instructions | Status |
|-----------|-------------|--------|
| aha-mont64 | 62 | ❌ Missing EXTR instruction |
| crc32 | 1,569,645 | ⚠️ Exit code -1 |
| matmult-int | 3,849,380 | ⚠️ Exit code -1 |

**Issue #179** created for EXTR instruction (needed for aha-mont64).

## Recent Progress

### This Cycle (193)
- **CoreMark verified:** Runs 15+ seconds without crash (SP fix working!)
- **Eric tested Embench:** 2 of 3 benchmarks execute millions of instructions
- **Cathy's PR #180:** Pipeline coverage up to 77.4%
- **Issue #179** created: Missing EXTR instruction

### Previous Cycle (192)
- **PR #175 merged** (Bob): ADD/SUB SP handling + NOP
- **PR #178 merged** (Cathy): Pipeline stats coverage tests
- **Issue #177 resolved**: Unit test hang fixed

## Calibration Milestones

| Milestone | Status | Description |
|-----------|--------|-------------|
| C1 | 🚧 Active | Execution Completeness — CoreMark appears to be running! |
| C2 | Pending | Microbenchmark Accuracy — <20% avg error |
| C3 | Pending | Intermediate Benchmark Accuracy |
| C4 | Pending | SPEC Benchmark Accuracy |

## Next Steps

1. Verify CoreMark completes execution (or find next blocker)
2. Implement EXTR instruction (#179) for aha-mont64
3. Investigate crc32 and matmult-int exit code -1
4. Merge PR #180 when bob-approved + CI passes
