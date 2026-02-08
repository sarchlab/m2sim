// load_heavy.s - 20 loads from sequential addresses
// Tests load unit throughput and memory subsystem pressure
// Matches: benchmarks/microbenchmarks.go loadHeavy()
//
// Expected: X0 = 20 at exit (value from last load)

.global _main
.align 4

_main:
    // Allocate stack space for buffer (160 bytes = 20 * 8 bytes)
    sub sp, sp, #160

    // Initialize buffer with known values (1, 2, 3, ... 20)
    mov x1, sp          // Base address
    mov x2, #1          // Counter

    // Pre-fill memory loop
fill_loop:
    cmp x2, #20
    b.gt fill_done

    // Calculate offset: (x2-1) * 8 = (x2-1) << 3
    sub x3, x2, #1      // x3 = counter - 1
    lsl x3, x3, #3      // x3 = (counter - 1) * 8
    str x2, [x1, x3]    // store counter at [base + offset]

    add x2, x2, #1      // increment counter
    b fill_loop

fill_done:

    // --- Timing region starts here ---
    // 20 loads to independent registers (no RAW hazards)
    ldr x0, [x1, #0]    // x0 = 1
    ldr x2, [x1, #8]    // x2 = 2
    ldr x3, [x1, #16]   // x3 = 3
    ldr x4, [x1, #24]   // x4 = 4
    ldr x5, [x1, #32]   // x5 = 5
    ldr x6, [x1, #40]   // x6 = 6
    ldr x7, [x1, #48]   // x7 = 7
    ldr x9, [x1, #56]   // x9 = 8 (skip x8, used for syscall)
    ldr x10, [x1, #64]  // x10 = 9
    ldr x11, [x1, #72]  // x11 = 10
    ldr x12, [x1, #80]  // x12 = 11
    ldr x13, [x1, #88]  // x13 = 12
    ldr x14, [x1, #96]  // x14 = 13
    ldr x15, [x1, #104] // x15 = 14
    ldr x16, [x1, #112] // x16 = 15
    ldr x17, [x1, #120] // x17 = 16
    ldr x18, [x1, #128] // x18 = 17
    ldr x19, [x1, #136] // x19 = 18
    ldr x20, [x1, #144] // x20 = 19
    ldr x0, [x1, #152]  // x0 = 20 (final value for exit)
    // --- Timing region ends here ---

    // Cleanup stack
    add sp, sp, #160

    // Exit syscall (x0 = 20)
    mov x16, #1         // SYS_exit
    svc #0x80