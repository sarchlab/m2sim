// store_heavy.s - 20 stores to sequential addresses
// Tests store unit throughput and write buffer behavior
// Matches: benchmarks/microbenchmarks.go storeHeavy()
//
// Expected: X0 = 3 at exit

.global _main
.align 4

_main:
    // Allocate stack space for buffer (160 bytes = 20 * 8 bytes)
    sub sp, sp, #160

    // Initialize
    mov x0, #3          // Exit code (preserved throughout)
    mov x1, sp          // Base address (stack buffer)
    mov x2, #99         // Value to store

    // --- Timing region starts here ---
    // 20 stores to sequential addresses (no data dependencies)
    str x2, [x1, #0]    // offset 0 * 8 = 0
    str x2, [x1, #8]    // offset 1 * 8 = 8
    str x2, [x1, #16]   // offset 2 * 8 = 16
    str x2, [x1, #24]   // offset 3 * 8 = 24
    str x2, [x1, #32]   // offset 4 * 8 = 32
    str x2, [x1, #40]   // offset 5 * 8 = 40
    str x2, [x1, #48]   // offset 6 * 8 = 48
    str x2, [x1, #56]   // offset 7 * 8 = 56
    str x2, [x1, #64]   // offset 8 * 8 = 64
    str x2, [x1, #72]   // offset 9 * 8 = 72
    str x2, [x1, #80]   // offset 10 * 8 = 80
    str x2, [x1, #88]   // offset 11 * 8 = 88
    str x2, [x1, #96]   // offset 12 * 8 = 96
    str x2, [x1, #104]  // offset 13 * 8 = 104
    str x2, [x1, #112]  // offset 14 * 8 = 112
    str x2, [x1, #120]  // offset 15 * 8 = 120
    str x2, [x1, #128]  // offset 16 * 8 = 128
    str x2, [x1, #136]  // offset 17 * 8 = 136
    str x2, [x1, #144]  // offset 18 * 8 = 144
    str x2, [x1, #152]  // offset 19 * 8 = 152
    // --- Timing region ends here ---

    // Cleanup stack
    add sp, sp, #160

    // Exit syscall (x0 = 3)
    mov x16, #1         // SYS_exit
    svc #0x80