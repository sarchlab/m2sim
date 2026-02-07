# SPEC.md - M2Sim Project Specification

## Project Goal

Build a cycle-accurate Apple M2 CPU simulator using the Akita simulation framework that can execute ARM64 user-space programs and predict execution time with high accuracy.

## Success Criteria

- [x] Execute ARM64 user-space programs correctly (functional emulation)
- [ ] Predict execution time with <20% average error across benchmarks
- [ ] Modular design: functional and timing simulation are separate
- [ ] Support benchmarks in μs to ms range

## Design Philosophy

### Independence from MGPUSim

While M2Sim uses Akita (like MGPUSim) and draws inspiration from MGPUSim's architecture, **M2Sim is not bound to follow MGPUSim's structure**. Make design decisions that best fit an ARM64 CPU simulator.

**Guidelines:**

1. **Choose meaningful names**: If a different name is more appropriate, use it
2. **Adapt to CPU semantics**: GPU and CPU have different abstractions (no wavefronts, warps, or GPU-specific concepts)
3. **Keep it simple**: M2Sim targets single-core initially
4. **Diverge when it makes sense**: Document why you're doing it differently

**What to Keep from MGPUSim:**
- Akita component/port patterns (they work well)
- Separation of concerns (functional vs timing)
- Testing practices (Ginkgo/Gomega)

**When in Doubt:** Ask "What would make this clearest for a CPU simulator?" — not "What does MGPUSim do?"

## Milestones

### High-Level Milestones

| # | Milestone | Status |
|---|-----------|--------|
| H1 | Core simulator (decode, execute, timing, caches) | ✅ COMPLETE |
| H2 | SPEC benchmark enablement (syscalls, ELF loading, validation) | 🚧 IN PROGRESS |
| H3 | Accuracy calibration (<20% error on SPEC) | ⬜ NOT STARTED |
| H4 | Multi-core support | ⬜ NOT STARTED |

---

### H1: Core Simulator ✅ COMPLETE

All foundation work is done: ARM64 decode, ALU/Load/Store/Branch instructions, pipeline timing (Fetch/Decode/Execute/Memory/Writeback), cache hierarchy (L1I, L1D, L2), branch prediction, 8-wide superscalar, macro-op fusion, SIMD basics. Microbenchmark suite established with 34.2% average CPI error.

<details>
<summary>Completed sub-milestones (M1–M5, C1)</summary>

- M1: Foundation — project scaffold, decoder, register file, ALU, load/store, branches
- M2: Memory & control flow — syscall emulation (exit, write), flat memory, end-to-end C programs
- M3: Timing model — pipeline stages, instruction timing
- M4: Cache hierarchy — L1I, L1D, L2 caches with timing
- M5: Advanced features — branch prediction, 8-wide superscalar, macro-op fusion, SIMD
- C1: Baseline — microbenchmarks created, M2 data collected, initial error 39.8% → 34.2%

</details>

---

### H2: SPEC Benchmark Enablement 🚧 IN PROGRESS

**Goal:** Run SPEC CPU 2017 integer benchmarks end-to-end in M2Sim.

#### H2.1: Syscall Coverage (medium-level) 🚧 IN PROGRESS

Complete the set of Linux syscalls needed by SPEC benchmarks.

##### H2.1.1: Core file I/O syscalls ✅ COMPLETE
- [x] read (63), write (64), close (57), openat (56) — all merged
- [x] FD table infrastructure — merged
- [x] fstat (80) — merged
- [x] File I/O acceptance tests — merged (PR #283)

##### H2.1.2: Memory management syscalls ✅ COMPLETE
- [x] brk (214) — merged
- [x] mmap (222) — merged

##### H2.1.3: Remaining file syscalls 🚧 IN PROGRESS (~5-10 cycles)
- [ ] lseek (62) — PR #282 open, cathy-approved, awaiting merge
- [ ] exit_group (94) — issue #272 open
- [ ] mprotect (226) — issue #278 open, research done

##### H2.1.4: Lower-priority syscalls ⬜ NOT STARTED (~10-20 cycles)
- [ ] munmap (215) — issue #271
- [ ] clock_gettime (113) — issue #274
- [ ] getpid/getuid/gettid — issue #273
- [ ] newfstatat (79) — may be needed by some benchmarks

#### H2.2: SPEC Binary Preparation (medium-level) 🚧 BLOCKED

**Blocker (issue #285):** SPEC binaries on the host machine are Mach-O (macOS native), but M2Sim requires ARM64 Linux ELF format. **Requires human action** to cross-compile SPEC using `aarch64-linux-musl-gcc`.

##### H2.2.1: Cross-compilation setup ⬜ BLOCKED ON HUMAN
- [ ] Install ARM64 Linux cross-compiler with musl libc
- [ ] Create SPEC config for ARM64 Linux static ELF
- [ ] Rebuild SPEC benchmarks

##### H2.2.2: Benchmark validation (~10-20 cycles per benchmark)
- [ ] 548.exchange2_r — Sudoku solver, pure computation, easiest target
- [ ] 505.mcf_r — vehicle scheduling, tests file I/O path
- [ ] 541.leela_r — Go AI, minimal I/O
- [ ] 531.deepsjeng_r — chess engine, larger memory

#### H2.3: Instruction Coverage Gaps ⬜ NOT STARTED

SPEC benchmarks will likely exercise ARM64 instructions not yet implemented. Expect to discover and fix gaps during validation (H2.2.2).

---

### H3: Accuracy Calibration ⬜ NOT STARTED

**Goal:** Achieve <20% average CPI error on SPEC benchmarks vs real M2 hardware.

**Current microbenchmark baseline (cycle 230):**

| Benchmark | Sim CPI | M2 CPI | Error |
|-----------|---------|--------|-------|
| arithmetic_sequential | 0.400 | 0.268 | 49.3% |
| dependency_chain | 1.200 | 1.009 | 18.9% |
| branch_taken_conditional | 1.600 | 1.190 | 34.5% |
| **Average** | — | — | **34.2%** |

#### H3.1: Pipeline tuning (~50-100 cycles)
- [ ] Full 8-wide execution (expected: 49.3% → ~28% arithmetic error)
- [ ] Out-of-order execution modeling
- [ ] Memory latency calibration

#### H3.2: SPEC-level calibration (~100+ cycles)
- [ ] Run SPEC benchmarks with timing, compare to M2 hardware
- [ ] Tune parameters to minimize error
- [ ] All benchmarks <30% individual error, <20% average

---

### H4: Multi-Core Support ⬜ NOT STARTED

Long-term goal (issue #139). Not planned in detail yet.

## Scope

### In Scope
- ARM64 user-space instructions
- CPU core simulation (single-core MVP, multi-core later)
- Cache hierarchy
- Timing prediction

### Out of Scope
- GPU / Neural Engine
- Kernel-space execution
- Full OS simulation
- I/O devices beyond basic syscalls

## Technical Constraints

- Use Akita v4 simulation framework
- Follow MGPUSim architecture patterns
- Go programming language
- Tests use Ginkgo/Gomega

## References

- Akita: https://github.com/sarchlab/akita
- MGPUSim: https://github.com/sarchlab/mgpusim
- ARM Architecture Reference Manual
- See `docs/calibration.md` for timing parameter reference
