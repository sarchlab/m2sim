# M2Sim Progress Report

**Last updated:** 2026-02-05 07:59 EST (Cycle 234)

## Current Status

| Metric | Value |
|--------|-------|
| Total PRs Merged | 64 |
| Open PRs | 2 |
| Open Issues | 15 |
| Pipeline Coverage | 77.0% |

## Cycle 234 Updates

- **PR #222** (Cathy load/store pair, PC-relative, move-wide tests) — **MERGED ✅**
- **Emu coverage: 55.8% → 62.4%** (+6.6pp)
- **PR #223** (Bob arithmetic_8wide) — CI failed, needs Cathy review
- **PR #225** (Cathy SIMD tests) — CI running, needs Bob review

## Key Progress This Cycle

**PR #222 — Load/store pair, PC-relative, and move-wide tests (MERGED ✅)**
- 22 new test cases covering:
  - LDP/STP (load/store pair): 64-bit, 32-bit, pre/post-index
  - ADR/ADRP (PC-relative addressing)
  - MOVZ/MOVN/MOVK (move wide)
  - Emulator options (WithStderr, WithSyscallHandler, Reset)
- Emu coverage improved from 55.8% to 62.4%

**Bob's 8-wide validation results (cycle 234):**
| Benchmark | CPI | Cycles | Instructions |
|-----------|-----|--------|--------------|
| arithmetic_sequential | 2.412 | 41 | 17 |
| arithmetic_6wide | 1.864 | 41 | 22 |
| arithmetic_8wide | 1.625 | 52 | 32 |

CPI improved from 1.864 (6-wide) to 1.625 (8-wide) — confirms infrastructure is working!

## Accuracy Status (Microbenchmarks)

| Benchmark | Simulator CPI | M2 Real CPI | Error | Notes |
|-----------|---------------|-------------|-------|-------|
| arithmetic | 0.400 | 0.268 | 49.3% | 8-wide now enabled |
| dependency | 1.200 | 1.009 | 18.9% | ✅ Near target |
| branch_taken_conditional | 1.600 | 1.190 | 34.5% | ↓ from 62.5% |
| **Average** | — | — | 34.2% | Target: <20% |

**Key insight:** Issue #221 (arithmetic_8wide using X0-X7) is implemented in PR #223, awaiting approval to validate 8-wide improvement.

## Coverage Analysis

| Package | Coverage | Status |
|---------|----------|--------|
| timing/cache | 89.1% | ✅ |
| timing/pipeline | 77.0% | ✅ |
| timing/latency | 73.3% | ✅ |
| timing/core | 100% | ✅ |
| emu | 62.4% | ↑ Target: 70%+ |

## Open PRs

| PR | Title | Status | Needs |
|----|-------|--------|-------|
| #223 | [Bob] arithmetic_8wide benchmark | CI FAIL | cathy-approved |
| #225 | [Cathy] SIMD coverage tests | CI running | bob-approved |

## Potential Accuracy Improvements

Per Eric's analysis:
1. ~~CMP + B.cond fusion~~ — **DONE** (PR #212)
2. ~~8-wide decode~~ — **DONE** (PR #215)
3. ~~8-wide benchmark enable~~ — **DONE** (PR #220)
4. arithmetic_8wide benchmark (PR #223) — awaiting approval
5. Branch predictor tuning (see docs/branch-predictor-tuning.md)
6. Pipeline stall reduction

## Calibration Milestones

| Milestone | Status | Description |
|-----------|--------|-------------|
| C1 | ✅ Complete | Benchmarks execute to completion |
| C2 | 🚧 In Progress | Accuracy calibration — 34.2% avg, target <20% |
| C3 | Pending | Intermediate benchmark timing (PolyBench) |

## Stats

- 64 PRs merged total
- 205+ tests passing
- timing/core coverage: 100% ✓
- emu coverage: 62.4% (target 70%+)
