// branch_hot_loop.s - 16-iteration loop with single hot branch
// Tests zero-cycle branch folding with repeated branch execution
// Matches: benchmarks/microbenchmarks.go branchHotLoop()
//
// Zero-cycle folding requires:
// 1. BTB hit (target known from previous execution)
// 2. Predicted taken
// 3. High confidence (counter >= 3, trained by 3+ executions)
//
// With 16 iterations:
// - Iterations 1-3: Normal branch penalty (training phase)
// - Iterations 4-16: Zero-cycle folding (13 folded branches expected)
//
// Expected: X0 = 0 at exit (loop counter decrements to 0)

.global _main
.align 4

_main:
    // Initialize loop counter
    mov x0, #16         // X0 = 16 (loop counter)

    // --- Timing region starts here ---
.Lloop:
    sub x0, x0, #1      // X0 = X0 - 1
    cmp x0, #0          // Compare X0 to 0
    b.ne .Lloop         // Branch back if X0 != 0 (HOT BRANCH: same PC, 16 times)
    // --- Timing region ends here ---

    // Exit syscall (x0 = 0 at this point)
    mov x16, #1         // SYS_exit
    svc #0x80
