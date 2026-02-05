# M2Sim Progress Report

**Last updated:** 2026-02-05 07:48 EST (Cycle 233)

## Current Status

| Metric | Value |
|--------|-------|
| Total PRs Merged | 63 |
| Open PRs | 0 |
| Open Issues | 13 |
| Pipeline Coverage | 77.0% |

## Cycle 233 Updates

- **PR #220** (Bob 8-wide benchmark enable) — **MERGED ✅**
- **Issue #219** — Closed (addressed by PR #220)
- **8-wide decode now active in benchmarks** — accuracy validation ready

## Key Progress This Cycle

**PR #220 — Enable 8-wide superscalar in benchmarks (MERGED ✅)**
- Benchmark harness now uses `EnableOctupleIssue: true` by default
- Enables proper validation of 8-wide decode infrastructure (PR #215)
- Next step: Run quick-calibration.sh to measure 8-wide improvement

## Accuracy Status (Microbenchmarks)

| Benchmark | Simulator CPI | M2 Real CPI | Error | Notes |
|-----------|---------------|-------------|-------|-------|
| arithmetic | 0.400 | 0.268 | 49.3% | 8-wide now enabled |
| dependency | 1.200 | 1.009 | 18.9% | ✅ Near target |
| branch_taken_conditional | 1.600 | 1.190 | 34.5% | ↓ from 62.5% |
| **Average** | — | — | 34.2% | Target: <20% |

**Key insight (Bob):** Current benchmarks use only 5-6 registers, limiting parallelism. Issue #221 (arithmetic_8wide using X0-X7) needed for true 8-wide validation.

## Coverage Analysis

| Package | Coverage | Status |
|---------|----------|--------|
| timing/cache | 89.1% | ✅ |
| timing/pipeline | 77.0% | ✅ |
| timing/latency | 73.3% | ✅ |
| timing/core | 100% | ✅ |
| emu | 55.8% | Target: 70%+ |

## Active Work

- Issue #221: Create arithmetic_8wide benchmark using X0-X7 registers (Eric)
- Emu coverage improvements ongoing (Cathy)

## Potential Accuracy Improvements

Per Eric's analysis:
1. ~~CMP + B.cond fusion~~ — **DONE** (PR #212)
2. ~~8-wide decode~~ — **DONE** (PR #215)
3. ~~8-wide benchmark enable~~ — **DONE** (PR #220)
4. arithmetic_8wide benchmark (Issue #221) — needed for true 8-wide validation
5. Branch predictor tuning (see docs/branch-predictor-tuning.md)
6. Pipeline stall reduction

## Calibration Milestones

| Milestone | Status | Description |
|-----------|--------|-------------|
| C1 | ✅ Complete | Benchmarks execute to completion |
| C2 | 🚧 In Progress | Accuracy calibration — 34.2% avg, target <20% |
| C3 | Pending | Intermediate benchmark timing (PolyBench) |

## Stats

- 63 PRs merged total
- 205+ tests passing
- timing/core coverage: 100% ✓
- emu coverage: 55.8% (target 70%+)
