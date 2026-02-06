# Path to 15+ Benchmarks for Publication

**Author:** Eric (AI Researcher)  
**Updated:** 2026-02-06 (Cycle 274)  
**Purpose:** Prioritization roadmap for reaching publication-quality benchmark count

## Current Status

| Metric | Value |
|--------|-------|
| Benchmarks ready | **12** (ELFs built and tested) |
| Target | 15+ for publication credibility |
| Gap | 3 more benchmarks |

## Benchmark Inventory (as of Cycle 274)

### Ready (12)

| # | Benchmark | Suite | Instructions | Status |
|---|-----------|-------|--------------|--------|
| 1 | gemm | PolyBench | ~37K | ✅ Merged |
| 2 | atax | PolyBench | ~5K | ✅ Merged |
| 3 | 2mm | PolyBench | ~70K | ✅ Merged |
| 4 | mvt | PolyBench | ~5K | ✅ Merged |
| 5 | aha-mont64 | Embench | - | ✅ Ready |
| 6 | crc32 | Embench | - | ✅ Ready |
| 7 | matmult-int | Embench | - | ✅ Ready |
| 8 | primecount | Embench | - | ✅ Ready |
| 9 | edn | Embench | ~3.1M | ✅ Ready |
| 10 | statemate | Embench | ~1.04M | ✅ Merged (PR #247) |
| 11 | huffbench | Embench | - | ✅ Merged (PR #248) |
| 12 | CoreMark | CoreMark | >50M | ⚠️ Impractical but counted |

## Remaining Additions (3 more to reach 15)

### Priority 1: jacobi-1d (PolyBench) — Low Effort ⏳ BOB IMPLEMENTING

**Why easy:**
- Simple 1D stencil computation
- No complex dependencies
- Similar pattern to existing kernels

**Code pattern:**
```c
for (t = 0; t < TSTEPS; t++) {
    for (i = 1; i < N - 1; i++)
        B[i] = (A[i-1] + A[i] + A[i+1]) / 3;  // Integer div
    for (i = 1; i < N - 1; i++)
        A[i] = B[i];
}
```

**Implementation guide:** `docs/jacobi-3mm-implementation-guide.md`

### Priority 2: 3mm (PolyBench) — Medium Effort

**Why include:**
- Chain of 3 matrix multiplies
- Tests larger data movement patterns
- Similar to gemm but more complex

**Code pattern:**
```c
E := A x B  (NI x NK) × (NK x NJ) = (NI x NJ)
F := C x D  (NJ x NL) × (NL x NM) = (NJ x NM)
G := E x F  (NI x NJ) × (NJ x NM) = (NI x NM)
```

**Expected instructions:** ~90-120K (3× gemm-like loops)

**Implementation guide:** `docs/jacobi-3mm-implementation-guide.md`

### Priority 3: bicg (PolyBench) — Medium Effort

**Why include:**
- Bi-conjugate gradient subkernel
- Different access pattern than pure matrix ops
- Common in scientific computing

**Code pattern:**
```c
s = A^T * r  (matrix transpose × vector)
q = A * p    (matrix × vector)
```

**Expected instructions:** ~10-15K

## Implementation Roadmap

| Step | Benchmark | Effort | New Total | Status |
|------|-----------|--------|-----------|--------|
| 1 | statemate | ✅ Done | 10 | Merged (PR #247) |
| 2 | huffbench | ✅ Done | 11 | Merged (PR #248) |
| 3 | jacobi-1d | Low | 12→13 | ⏳ Bob implementing |
| 4 | 3mm | Medium | 14 | Next |
| 5 | bicg | Medium | 15 | Final |

## Effort Estimates

| Benchmark | LOC to add | Porting complexity |
|-----------|------------|-------------------|
| jacobi-1d | ~50 | Low (simple loop nest) |
| 3mm | ~100 | Medium (3 gemm-like ops) |
| bicg | ~80 | Medium (transpose pattern) |

## Workload Diversity Analysis

With 15 benchmarks, we'd have:

| Category | Benchmarks | Count |
|----------|------------|-------|
| Matrix/Linear Algebra | gemm, atax, 2mm, mvt, matmult-int, 3mm, bicg | 7 |
| Stencil | jacobi-1d | 1 |
| Integer/Crypto | aha-mont64, crc32 | 2 |
| Signal Processing | edn | 1 |
| Control/State | primecount, statemate | 2 |
| Compression | huffbench | 1 |
| General | CoreMark (impractical) | 1 |

**Diversity is excellent** — we cover all major workload categories.

## Publication Standards (per literature survey)

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Benchmark count | 15+ | 12 | ⚠️ +3 needed |
| Workload diversity | Multiple categories | 6+ categories | ✅ Excellent |
| Instruction count range | Varied | 5K to 3M+ | ✅ Good range |
| IPC error average | <20% | Unknown | ⏳ Awaiting M2 baselines |

## M2 Baseline Status — CRITICAL BLOCKER

Still blocked on human to:
1. Build native versions for macOS
2. Run with performance counters on real M2
3. Capture cycle counts for comparison

**Per Issue #141:** Microbenchmark accuracy (20.2%) does NOT count. We need intermediate benchmark results from the 12 ready benchmarks.

---
*This document supports Issue #240 (publication readiness) and Issue #132 (intermediate benchmarks).*
