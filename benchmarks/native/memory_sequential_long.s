// memory_sequential_long.s - Long-running memory access benchmark
// 10M iterations of load/store operations
// Purpose: Benchmark execution time >> process startup overhead (~18ms)
//
// Each iteration: Store + Load cycle with data dependency
// Tests memory subsystem performance.
//
// Note: Uses stack space for memory operations

.global _main
.align 4

_main:
    // Reserve stack space for our data
    sub sp, sp, #64          // allocate 64 bytes on stack

    // Initialize loop counter
    mov x10, #0              // iteration counter
    // Load 10000000 (0x989680) into x11
    movz x11, #0x9680
    movk x11, #0x0098, lsl #16

    // Initialize value to store
    mov x0, #0

.loop:
    // --- Timing region: Memory operations ---
    // Store and load cycle (creates memory dependency)
    str x0, [sp]             // store value
    ldr x1, [sp]             // load it back (dependent on store)
    add x0, x1, #1           // increment (dependent on load)

    str x0, [sp, #8]         // store to offset 8
    ldr x1, [sp, #8]         // load back
    add x0, x1, #1           // increment

    str x0, [sp, #16]        // store to offset 16
    ldr x1, [sp, #16]        // load back
    add x0, x1, #1           // increment

    str x0, [sp, #24]        // store to offset 24
    ldr x1, [sp, #24]        // load back
    add x0, x1, #1           // increment

    str x0, [sp, #32]        // store to offset 32
    ldr x1, [sp, #32]        // load back
    add x0, x1, #1           // increment (5 increments per iteration)
    // --- End timing region ---

    // Loop control
    add x10, x10, #1         // increment counter
    cmp x10, x11             // compare with limit
    b.lt .loop               // branch if less than

    // Restore stack
    add sp, sp, #64

    // x0 = 5M, exit code is 8-bit
    // 5M mod 256 = 5000000 mod 256 = 32 (5000000 = 19531 * 256 + 32)
    and x0, x0, #0xFF

    // Exit syscall
    mov x16, #1              // SYS_exit
    svc #0x80
