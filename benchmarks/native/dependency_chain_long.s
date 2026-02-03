// dependency_chain_long.s - Long-running dependency chain benchmark
// 10M iterations of dependent operations (RAW hazards)
// Purpose: Benchmark execution time >> process startup overhead (~18ms)
//
// Each iteration: 20 dependent ADDs in a chain = high CPI
// This stresses the pipeline with data dependencies.
//
// Exit code: Final counter value modulo 256

.global _main
.align 4

_main:
    // Initialize loop counter
    mov x10, #0              // iteration counter
    // Load 10000000 (0x989680) into x11
    movz x11, #0x9680
    movk x11, #0x0098, lsl #16

    // Initialize accumulator
    mov x0, #0

.loop:
    // --- Timing region: 20 dependent ADDs (RAW chain) ---
    // Each instruction depends on the previous one
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1

    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1

    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1

    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    // --- End timing region ---

    // Loop control
    add x10, x10, #1         // increment counter
    cmp x10, x11             // compare with limit
    b.lt .loop               // branch if less than

    // x0 = 20M, exit code is 8-bit
    // 20M mod 256 = 0 (20M = 20000000 = 78125 * 256)
    and x0, x0, #0xFF

    // Exit syscall
    mov x16, #1              // SYS_exit
    svc #0x80
