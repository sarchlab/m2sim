// arithmetic_sequential_long.s - Long-running arithmetic benchmark
// 10M iterations of independent ADD operations
// Purpose: Benchmark execution time >> process startup overhead (~18ms)
//
// Each iteration: 20 independent ADDs = 200M total instructions
// Expected execution time: ~50-100ms (at 2-4B instructions/sec) + ~18ms overhead
// This allows meaningful timing measurements.
//
// Exit code: Loop counter final value modulo 256 (implementation detail)

.global _main
.align 4

_main:
    // Initialize loop counter
    mov x10, #0              // iteration counter
    // Load 10000000 (0x989680) into x11
    movz x11, #0x9680        // lower 16 bits
    movk x11, #0x0098, lsl #16  // upper bits (10000000 = 0x00989680)

    // Initialize work registers
    mov x0, #0
    mov x1, #0
    mov x2, #0
    mov x3, #0
    mov x4, #0

.loop:
    // --- Timing region: 20 independent ADDs ---
    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1

    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1

    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1

    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1
    // --- End timing region ---

    // Loop control
    add x10, x10, #1         // increment counter
    cmp x10, x11             // compare with limit
    b.lt .loop               // branch if less than

    // x0 = 4M, but exit code is 8-bit, so we use a predictable value
    // Return x0 & 0xFF for exit code (will be 0 since 4M mod 256 = 0)
    and x0, x0, #0xFF

    // Exit syscall
    mov x16, #1              // SYS_exit
    svc #0x80
