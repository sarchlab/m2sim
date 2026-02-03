// mixed_operations_long.s - Long-running mixed operations benchmark
// 10M iterations with varied instruction types
// Purpose: Benchmark execution time >> process startup overhead (~18ms)
//
// Each iteration: Mix of arithmetic, logic, and data movement
// Represents more realistic workload patterns.

.global _main
.align 4

_main:
    // Reserve stack space
    sub sp, sp, #32

    // Initialize loop counter
    mov x10, #0              // iteration counter
    // Load 10000000 (0x989680) into x11
    movz x11, #0x9680
    movk x11, #0x0098, lsl #16

    // Initialize work registers
    mov x0, #0
    mov x1, #1
    mov x2, #2
    mov x3, #3

.loop:
    // --- Timing region: Mixed operations ---
    // Arithmetic
    add x0, x0, #1
    sub x1, x1, #1
    add x1, x1, #2           // net +1 to x1

    // Logic operations
    and x4, x0, #0xFF
    orr x5, x1, x2
    eor x6, x2, x3

    // Data movement
    mov x7, x0
    str x7, [sp]
    ldr x8, [sp]

    // More arithmetic with dependencies
    add x0, x0, x8           // x0 += x0 (doubles)
    sub x0, x0, x7           // x0 -= old_x0 (back to +1)

    // Shift operations
    lsl x9, x1, #1
    lsr x9, x9, #1           // shift left then right
    // --- End timing region ---

    // Loop control
    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    // Restore stack
    add sp, sp, #32

    // Final x0 = 1M (incremented by 1 each iteration)
    // 1M mod 256 = 64 (1000000 = 3906 * 256 + 64)
    and x0, x0, #0xFF

    // Exit syscall
    mov x16, #1
    svc #0x80
