# PolyBench Phase 1 Implementation Approach

**Author:** Eric (Researcher)
**Created:** 2026-02-05 (Cycle 263)
**Related Issues:** #237, #191, #132

## Overview

This document provides a detailed implementation approach for adding PolyBench Phase 1 benchmarks (gemm and atax) to M2Sim.

## Phase 1 Targets

| Benchmark | Description | Why First |
|-----------|-------------|-----------|
| gemm | Matrix multiply C=α.A.B+β.C | Most common BLAS kernel, tests compute density |
| atax | Matrix transpose + vector multiply | Integer indices, tests memory access patterns |

## Implementation Steps

### Step 1: Clone PolyBench/C

```bash
cd benchmarks
git clone https://github.com/MatthiasJReisinger/PolyBenchC-4.2.1.git polybench-src
```

### Step 2: Create Bare-Metal Build Infrastructure

Create `benchmarks/polybench/` directory structure:

```
benchmarks/polybench/
├── Makefile          # Cross-compilation rules
├── common/
│   ├── polybench.h   # Modified (no libc deps)
│   └── startup.S     # Bare-metal entry point
├── gemm/
│   └── gemm.c        # Adapted kernel
└── atax/
    └── atax.c        # Adapted kernel
```

### Step 3: Makefile Template

Based on existing native benchmark patterns:

```makefile
# Cross-compiler settings
CC = aarch64-elf-gcc
CFLAGS = -O2 -static -nostdlib -ffreestanding
CFLAGS += -DPOLYBENCH_USE_RESTRICT -DMINI_DATASET

# Linker settings (bare-metal)
LDFLAGS = -static -nostdlib -e _start

# Benchmarks
BENCHMARKS = gemm atax

.PHONY: all clean

all: $(BENCHMARKS)

gemm: gemm/gemm.c common/startup.S
	$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $^

atax: atax/atax.c common/startup.S
	$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $^

clean:
	rm -f $(BENCHMARKS) *.o
```

### Step 4: Modify polybench.h for Bare-Metal

Key changes needed:
- Remove `#include <stdio.h>` (no printf)
- Remove `#include <stdlib.h>` (no malloc)
- Use static arrays instead of dynamic allocation
- Use MINI_DATASET for small problem sizes

```c
// Simplified polybench.h for bare-metal
#ifndef _POLYBENCH_H
#define _POLYBENCH_H

// Dataset size (MINI for validation)
#ifndef NI
  #define NI 16
  #define NJ 16
  #define NK 16
#endif

// Timing macros (stub for bare-metal)
#define polybench_start_instruments
#define polybench_stop_instruments
#define polybench_print_instruments

#endif // _POLYBENCH_H
```

### Step 5: Create startup.S Entry Point

```asm
// startup.S - Bare-metal entry for PolyBench
.section .text
.global _start

_start:
    // Set up stack pointer
    ldr x0, =_stack_top
    mov sp, x0
    
    // Call main
    bl main
    
    // Exit with return value
    mov x8, #93          // exit syscall
    svc #0
    
.section .bss
.align 4
_stack:
    .space 4096
_stack_top:
```

### Step 6: Adapt gemm Kernel

```c
// gemm/gemm.c - Matrix multiply for M2Sim
#include "../common/polybench.h"

#define DATA_TYPE int
#define ALPHA 1
#define BETA 1

// Static arrays (MINI dataset: 16x16)
DATA_TYPE A[NI][NK];
DATA_TYPE B[NK][NJ];
DATA_TYPE C[NI][NJ];

void init_array(void) {
    for (int i = 0; i < NI; i++)
        for (int k = 0; k < NK; k++)
            A[i][k] = i * NK + k;
    
    for (int k = 0; k < NK; k++)
        for (int j = 0; j < NJ; j++)
            B[k][j] = k * NJ + j;
    
    for (int i = 0; i < NI; i++)
        for (int j = 0; j < NJ; j++)
            C[i][j] = 0;
}

void kernel_gemm(void) {
    // C := alpha*A*B + beta*C
    for (int i = 0; i < NI; i++) {
        for (int j = 0; j < NJ; j++) {
            C[i][j] *= BETA;
        }
        for (int k = 0; k < NK; k++) {
            for (int j = 0; j < NJ; j++) {
                C[i][j] += ALPHA * A[i][k] * B[k][j];
            }
        }
    }
}

int main(void) {
    init_array();
    kernel_gemm();
    
    // Return checksum (lower 8 bits)
    int sum = 0;
    for (int i = 0; i < NI; i++)
        for (int j = 0; j < NJ; j++)
            sum += C[i][j];
    
    return sum & 0xFF;
}
```

### Step 7: Integrate into Accuracy Test Harness

Add to `benchmarks/timing_harness.go`:

```go
func (h *TimingHarness) RegisterPolyBenchBenchmarks() {
    h.RegisterBenchmark("polybench_gemm", "polybench/gemm", 
        BenchmarkConfig{ExpectedCycles: 0}) // TBD after M2 baseline
    
    h.RegisterBenchmark("polybench_atax", "polybench/atax",
        BenchmarkConfig{ExpectedCycles: 0})
}
```

### Step 8: Capture M2 Baseline Timing

```bash
cd benchmarks/polybench
./gemm
# Capture with Instruments/xctrace or perf
```

## Expected Outcomes

| Benchmark | Problem Size | Iterations | Expected Cycle Range |
|-----------|--------------|------------|----------------------|
| gemm | 16x16x16 | 4,096 MACs | ~10K-50K cycles |
| atax | 16x16 | 512 ops | ~5K-20K cycles |

## Dependencies

- [x] Cross-compiler: aarch64-elf-gcc 15.2.0 installed
- [x] Build pattern: Native Makefile as template
- [ ] M2 hardware: For baseline timing capture

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Floating-point ops | Use integer versions (DATA_TYPE = int) |
| Large matrices | Use MINI dataset (16x16) |
| Build complexity | Follow existing Embench pattern |
| Syscall dependencies | Remove printf, use static allocation |

## Timeline Estimate

| Task | Effort |
|------|--------|
| Directory setup | 1 cycle |
| gemm adaptation | 1-2 cycles |
| atax adaptation | 1 cycle |
| Test harness integration | 1 cycle |
| M2 baseline capture | 1 cycle |
| **Total** | **5-6 cycles** |

## Next Steps

1. →Bob: Create `benchmarks/polybench/` directory structure
2. →Bob: Implement gemm bare-metal build
3. →Cathy: Review build configuration
4. →Bob: Capture M2 baseline timing
5. →Eric: Analyze accuracy results
