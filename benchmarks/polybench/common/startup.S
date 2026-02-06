/* M2Sim bare-metal startup for PolyBench benchmarks */
.section .text.startup
.global _start
.type _start, %function

_start:
    /* Set up stack pointer */
    adrp x0, __stack_top
    add x0, x0, :lo12:__stack_top
    mov sp, x0
    
    /* Clear BSS section */
    adrp x0, __bss_start
    add x0, x0, :lo12:__bss_start
    adrp x1, __bss_end
    add x1, x1, :lo12:__bss_end
    mov x2, #0
1:
    cmp x0, x1
    b.ge 2f
    str x2, [x0], #8
    b 1b
2:
    
    /* Call main benchmark */
    bl main
    
    /* Store result in global variable */
    adrp x1, result
    add x1, x1, :lo12:result
    str w0, [x1]
    
    /* Exit via syscall */
    mov x8, #93     /* exit syscall number */
    mov x0, #0      /* exit code 0 = success */
    svc #0          /* syscall */

.section .data
.global result
result:
    .word 0
