# M2Sim Progress Report

**Last updated:** 2026-02-05 21:06 EST (Cycle 269)

## Current Status

| Metric | Value |
|--------|-------|
| Total PRs Merged | **77** 🎉 |
| Open PRs | 0 |
| Open Issues | 17 (excl. tracker) |
| Pipeline Coverage | 69.6% |
| Emu Coverage | 79.9% ✅ |

## Cycle 269 Updates

### 📊 Pipeline Coverage Progress

Cathy improved pipeline coverage to 69.6% (+0.2pp):
- Added tests for IsCMP, isUnconditionalBranch sign extension
- Added tests for BranchPredictor stats (Accuracy, MispredictionRate, BTBHitRate)
- Added tests for NewBranchPredictor default config, WritebackSlot
- 8 helper functions now at 100% coverage
- **Only ~0.4% remaining to reach 70% target!**

### 🔬 CoreMark Research Finding (Bob)

Bob investigated CoreMark execution and discovered:
- **All required instructions already work!** (ADRP, LDR literal, etc.)
- **Real blocker:** CoreMark takes **>50M instructions per iteration**
- This makes CoreMark impractical for validation timing
- Issue #241 closed — not a missing instruction issue

### 📊 Benchmark Inventory (Eric)

Eric created comprehensive benchmark inventory:
| Suite | Ready | Status |
|-------|-------|--------|
| PolyBench | 2 (gemm, atax) | ✅ |
| Embench | 4 (aha-mont64, crc32, matmult-int, primecount) | ✅ |
| CoreMark | 1 | ⚠️ Impractical (>50M instr) |
| **Total** | **7 ready** | Need 15+ for publication |

### 📋 New Issues Created

Eric created issues for benchmark expansion:
- #243: Build edn benchmark ELF (low effort)
- #244: Add PolyBench 2mm and mvt kernels
- #245: Add huffbench and statemate from Embench-IoT

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
| pipeline | 69.6% | ⚠️ Needs ~0.4% more for 70% |

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
3. **Pipeline coverage** — 69.6% → 70%+ target (~0.4% remaining)
4. **Benchmark expansion** — Build edn, add 2mm/mvt, huffbench/statemate (#243, #244, #245)
