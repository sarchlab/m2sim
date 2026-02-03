// branch_taken_long.s - Long-running branch prediction benchmark
// 10M iterations with predictable branches (always taken)
// Purpose: Benchmark execution time >> process startup overhead (~18ms)
//
// Each iteration: Multiple conditional branches that are always taken
// Tests branch prediction with predictable pattern.

.global _main
.align 4

_main:
    // Initialize loop counter
    mov x10, #0              // iteration counter
    // Load 10000000 (0x989680) into x11
    movz x11, #0x9680
    movk x11, #0x0098, lsl #16

    // Initialize work register
    mov x0, #0

.loop:
    // --- Timing region: Predictable branches ---
    // All these branches are always taken (x0 >= 0)
    cmp x0, #0
    b.ge .taken1             // always taken (x0 >= 0)
    add x0, x0, #100         // never executed
.taken1:
    add x0, x0, #1

    cmp x0, #0
    b.ge .taken2             // always taken
    add x0, x0, #100
.taken2:
    add x0, x0, #1

    cmp x0, #0
    b.ge .taken3             // always taken
    add x0, x0, #100
.taken3:
    add x0, x0, #1

    cmp x0, #0
    b.ge .taken4             // always taken
    add x0, x0, #100
.taken4:
    add x0, x0, #1

    cmp x0, #0
    b.ge .taken5             // always taken
    add x0, x0, #100
.taken5:
    add x0, x0, #1
    // --- End timing region: 5 increments per iteration ---

    // Loop control
    add x10, x10, #1         // increment counter
    cmp x10, x11             // compare with limit
    b.lt .loop               // branch if less than

    // x0 = 5M, exit code is 8-bit
    // 5M mod 256 = 32
    and x0, x0, #0xFF

    // Exit syscall
    mov x16, #1              // SYS_exit
    svc #0x80
