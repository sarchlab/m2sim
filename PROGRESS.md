# M2Sim Progress Report

**Last updated:** 2026-02-05 02:21 EST (Cycle 213)

## Current Status

| Metric | Value |
|--------|-------|
| Total PRs Merged | 50 |
| Open PRs | 0 |
| Open Issues | 13 |
| Pipeline Coverage | 76.7% |

## Cycle 213 Updates

- **Alice:** Incremented action count to 213, assigned new tasks
- **Eric:** Created docs/intermediate-benchmarks.md and docs/m2-microarchitecture.md
- **Bob:** Created #199 (branch prediction accuracy investigation)
- **Cathy:** Coverage analysis — identified dead code and interface coverage gaps

## New Research Documents

1. **docs/intermediate-benchmarks.md** — Evaluated 5 benchmark suites for validation
2. **docs/m2-microarchitecture.md** — M2 branch prediction research for accuracy tuning

## Embench Phase 1 — Complete! ✅

| Benchmark | Instructions | Exit Code | Status |
|-----------|-------------|-----------|--------|
| aha-mont64 | 1.88M | 0 ✓ | ✅ Complete |
| crc32 | 1.57M | 0 ✓ | ✅ Complete |
| matmult-int | 3.85M | 0 ✓ | ✅ Complete |

## Embench Phase 2 — Complete

| Issue | Benchmark | Status |
|-------|-----------|--------|
| #184 | primecount | ✅ Merged (2.84M instructions) |
| #185 | edn | ✅ Merged |
| #186 | huffbench | ❌ Closed (needs libc stubs) |
| #187 | statemate | ❌ Closed (needs libc stubs) |

**5 Embench benchmarks working** — sufficient for accuracy calibration

## Accuracy Status (Microbenchmarks)

| Benchmark | Simulator CPI | M2 Real CPI | Error |
|-----------|---------------|-------------|-------|
| arithmetic | 0.400 | 0.268 | 49.3% |
| dependency | 1.200 | 1.009 | 18.9% |
| branch | 1.800 | 1.190 | 51.3% |
| **Average** | — | — | **39.8%** |

**Target:** <20% average error (#141)

## Pipeline Refactor Progress (#122) — COMPLETE! ✅

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1 | ✅ Complete | WritebackSlot interface + implementations |
| Phase 2 | ✅ Complete | Replace inline writeback with helper calls |
| Phase 3 | ✅ Complete | Primary slot unified with WritebackSlot |
| Phase 4 | ✅ Complete | MemorySlot interface (PR #196 merged) |
| Tests | ✅ Complete | MemorySlot interface tests (PR #198 merged) |

All pipeline refactoring done! Foundation ready for accuracy tuning.

## Active Investigations

- **#199** — Branch prediction accuracy gap (51.3% error)
  - Initial finding: M2 defaults to "not-taken" for cold branches, we default to "weakly taken"
  - M2 Avalanche cores are 6-wide; our 4-wide may explain arithmetic gap

## Calibration Milestones

| Milestone | Status | Description |
|-----------|--------|-------------|
| C1 | ✅ Complete | Benchmarks execute to completion |
| C2 | 🚧 Blocked | Accuracy calibration — needs Embench timing data |
| C3 | Pending | Intermediate benchmark timing |
| C4 | Pending | SPEC benchmark accuracy |

## Current Blockers

1. **Embench timing simulation** — takes 5-10+ min per benchmark, needs human-triggered overnight run
   - Issue #197 created with instructions
   - Workaround: quick-calibration.sh for microbenchmarks

## Next Steps

1. Human triggers overnight timing batch job (see #197)
2. Once timing data available, proceed with C2 calibration milestone
3. Investigate branch accuracy (#199) and arithmetic accuracy gaps
4. Continue accuracy improvements (39.8% → <20% target)
