// memory_strided.s - 10 store/load pairs with stride-4 access
// Tests strided memory access pattern (stride = 32 bytes)
// Matches: benchmarks/microbenchmarks.go memoryStrided()
//
// Expected: X0 = 7 at exit (value stored and loaded back)

.global _main
.align 4

_main:
    // Allocate stack space for buffer (288 bytes to accommodate offset 36*8=288)
    sub sp, sp, #320

    // Initialize
    mov x0, #7          // Value to store/load
    mov x1, sp          // Base address (stack buffer)

    // --- Timing region starts here ---
    // 10 store/load pairs at stride-4 offsets (32-byte stride)
    // Offsets: 0, 4, 8, 12, 16, 20, 24, 28, 32, 36 (each unit = 8 bytes)
    str x0, [x1, #0]      // offset 0 * 8 = 0
    ldr x0, [x1, #0]

    str x0, [x1, #32]     // offset 4 * 8 = 32
    ldr x0, [x1, #32]

    str x0, [x1, #64]     // offset 8 * 8 = 64
    ldr x0, [x1, #64]

    str x0, [x1, #96]     // offset 12 * 8 = 96
    ldr x0, [x1, #96]

    str x0, [x1, #128]    // offset 16 * 8 = 128
    ldr x0, [x1, #128]

    str x0, [x1, #160]    // offset 20 * 8 = 160
    ldr x0, [x1, #160]

    str x0, [x1, #192]    // offset 24 * 8 = 192
    ldr x0, [x1, #192]

    str x0, [x1, #224]    // offset 28 * 8 = 224
    ldr x0, [x1, #224]

    str x0, [x1, #256]    // offset 32 * 8 = 256
    ldr x0, [x1, #256]

    str x0, [x1, #288]    // offset 36 * 8 = 288
    ldr x0, [x1, #288]
    // --- Timing region ends here ---

    // Cleanup stack
    add sp, sp, #320

    // Exit syscall (x0 = 7)
    mov x16, #1         // SYS_exit
    svc #0x80