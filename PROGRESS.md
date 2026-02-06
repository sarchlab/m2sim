# M2Sim Progress Report

**Last updated:** 2026-02-05 20:45 EST (Cycle 268)

## Current Status

| Metric | Value |
|--------|-------|
| Total PRs Merged | **77** 🎉 |
| Open PRs | 0 |
| Open Issues | 15 (excl. tracker) |
| Pipeline Coverage | 69.4% |
| Emu Coverage | 79.9% ✅ |

## Cycle 268 Updates

### 📊 Pipeline Coverage Progress

Cathy improved pipeline coverage to 69.4% (+0.5pp):
- Added 32-bit SUBS/CMP/AND/ORR/EOR tests
- Multiple utility functions now at 100% coverage
- **Only ~0.6% remaining to reach 70% target!**

### 📚 Literature Survey & CoreMark Research (Eric)

Eric completed publication standards research and CoreMark status analysis:
- Created `docs/literature-survey-simulator-validation.md`
- Publication standard: 15+ benchmarks needed, 11-25% IPC error typical (gem5)
- **CoreMark is blocked** — missing ADRP, LDR literal instructions
- Created issue #241 for missing instruction implementation

### ⚠️ M2 Baseline Capture — Requires Human

Bob confirmed M2 baseline capture for PolyBench benchmarks requires human involvement:
- Current ELFs are bare-metal for simulator, not native executables
- Need native macOS builds for performance counter integration
- Must run on actual M2 hardware with cycle measurements

### 🎯 Critical Validation Finding

Per issue #141, the 20.2% microbenchmark accuracy **doesn't count** — Human explicitly requires intermediate-size benchmarks:
> "Microbenchmarks should NOT be included in the accuracy measurement"

---

## PolyBench Phase 1 — COMPLETE! 🎉

| Benchmark | Status | Instructions |
|-----------|--------|--------------|
| gemm | ✅ Merged (PR #238) | ~37K |
| atax | ✅ Merged (PR #239) | ~5K |

Both benchmarks ready for M2 baseline capture and timing validation.

---

## Open PRs

None! 🎉 Clean slate.

## Key Achievements

**77 PRs Merged!**

**Emu Coverage Target Exceeded!**
| Package | Coverage | Status |
|---------|----------|--------|
| emu | 79.9% | ✅ Above 70% target! |
| pipeline | 68.9% | ⚠️ Needs ~1% more for 70% |

**All Timing Simulator Fixes Applied:**
| Fix | Commit | Status |
|-----|--------|--------|
| PSTATE forwarding | 9d7c2e6 | ✅ |
| Same-cycle flag forwarding | 48851e7 | ✅ |
| Branch handling slots 2-8 | d159a73 | ✅ |
| Disable unsafe branch folding | 1590518 | ✅ |
| Test count fix (11→12) | eb70656 | ✅ |

## Accuracy Status (Microbenchmarks)

| Benchmark | Sim CPI | M2 CPI | Error | Target |
|-----------|---------|--------|-------|--------|
| arithmetic_8wide | 0.250 | 0.268 | **7.2%** | ✅ <20% |
| dependency_chain | 1.200 | 1.009 | 18.9% | ✅ <20% |
| branch_conditional | 1.600 | 1.190 | **34.5%** | ❌ <20% |
| **Average** | — | — | **20.2%** | ⚠️ ~20% |

⚠️ **Note:** Per #141, microbenchmark accuracy doesn't count for M6 validation.

## Next Steps

1. **M2 baseline capture (requires human)** — Run gemm/atax on real M2 with performance counters
2. **Intermediate benchmark accuracy** — Measure PolyBench results against M2 baselines
3. **Pipeline coverage** — 68.9% → 70%+ target (~1.1% remaining)
4. **Safe zero-cycle folding** — Documented in docs/safe-zero-cycle-folding.md if needed
