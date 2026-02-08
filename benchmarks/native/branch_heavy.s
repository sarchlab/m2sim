// branch_heavy.s - 10 conditional branches (alternating taken/not-taken)
// Tests branch prediction stress with alternating patterns
// Matches: benchmarks/microbenchmarks.go branchHeavy()
//
// Expected: X0 = 10 at exit

.global _main
.align 4

_main:
    // Initialize
    mov x0, #0          // X0 = 0 (result counter)
    mov x1, #5          // X1 = 5 (comparison value)
    mov x3, #0          // X3 = 0 (not-taken counter)

    // --- Timing region starts here ---
    // Pattern: CMP X0, X1; B.LT +8 (taken while X0 < 5)
    // First 5 branches taken, last 5 not taken

    // Branch 1: X0=0 < 5, taken (skip corrupt instruction)
    cmp x0, x1          // CMP X0, X1
    b.lt .L1            // B.LT +8 (skip next instruction)
    add x1, x1, #99     // skipped (would corrupt X1)
.L1:
    add x0, x0, #1      // X0 += 1

    // Branch 2: X0=1 < 5, taken
    cmp x0, x1
    b.lt .L2
    add x1, x1, #99
.L2:
    add x0, x0, #1

    // Branch 3: X0=2 < 5, taken
    cmp x0, x1
    b.lt .L3
    add x1, x1, #99
.L3:
    add x0, x0, #1

    // Branch 4: X0=3 < 5, taken
    cmp x0, x1
    b.lt .L4
    add x1, x1, #99
.L4:
    add x0, x0, #1

    // Branch 5: X0=4 < 5, taken
    cmp x0, x1
    b.lt .L5
    add x1, x1, #99
.L5:
    add x0, x0, #1

    // Branch 6: X0=5 >= 5, NOT taken (falls through)
    cmp x0, x1
    b.lt .L6
    add x3, x3, #1      // X3 += 1 (not-taken counter)
.L6:
    add x0, x0, #1

    // Branch 7: X0=6 >= 5, NOT taken
    cmp x0, x1
    b.lt .L7
    add x3, x3, #1
.L7:
    add x0, x0, #1

    // Branch 8: X0=7 >= 5, NOT taken
    cmp x0, x1
    b.lt .L8
    add x3, x3, #1
.L8:
    add x0, x0, #1

    // Branch 9: X0=8 >= 5, NOT taken
    cmp x0, x1
    b.lt .L9
    add x3, x3, #1
.L9:
    add x0, x0, #1

    // Branch 10: X0=9 >= 5, NOT taken
    cmp x0, x1
    b.lt .L10
    add x3, x3, #1
.L10:
    add x0, x0, #1
    // --- Timing region ends here ---
    // X0 should now be 10

    // Exit syscall (x0 = 10)
    mov x16, #1         // SYS_exit
    svc #0x80