# Test Data

This directory contains test programs and documentation for M2Sim's end-to-end integration tests.

## Test Programs

The integration tests use synthetic ARM64 ELF binaries created programmatically (no cross-compiler required). The test programs include:

### exit_zero
Exits immediately with code 0.
```asm
ADD X8, XZR, #93    ; syscall number for exit
ADD X0, XZR, #0     ; exit code 0
SVC #0              ; invoke syscall
```

### exit_42
Exits with code 42.
```asm
ADD X8, XZR, #93    ; syscall number for exit
ADD X0, XZR, #42    ; exit code 42
SVC #0              ; invoke syscall
```

### hello
Writes "Hello, World!\n" to stdout and exits with code 0.
```asm
ADD X8, XZR, #64    ; syscall number for write
ADD X0, XZR, #1     ; fd = stdout
ADD X1, XZR, #0x600, LSL #12  ; buffer address (0x600000)
ADD X2, XZR, #14    ; length
SVC #0              ; write syscall
ADD X8, XZR, #93    ; syscall number for exit
ADD X0, XZR, #0     ; exit code 0
SVC #0              ; exit syscall
```

### compute
Computes 10 + 5 and exits with the result (15).
```asm
ADD X0, XZR, #10    ; X0 = 10
ADD X1, XZR, #5     ; X1 = 5
ADD X0, X0, X1      ; X0 = X0 + X1
ADD X8, XZR, #93    ; syscall number for exit
SVC #0              ; exit with result in X0
```

### loop
Counts down from 5 to 0 using a conditional branch loop.
```asm
ADD X0, XZR, #5     ; counter = 5
loop:
SUBS X0, X0, #1     ; counter--; set flags
B.NE loop           ; if not zero, continue
ADD X8, XZR, #93    ; syscall number for exit
SVC #0              ; exit with 0
```

### funcall
Tests function calls using BL (branch with link) and RET.
```asm
BL func             ; call function
ADD X8, XZR, #93    ; syscall number for exit
SVC #0              ; exit with return value in X0
func:
ADD X0, XZR, #100   ; return 100
RET                 ; return to caller
```

### multidata
Tests multi-segment ELF loading by reading a value from a data segment.
```asm
ADD X1, XZR, #0x600, LSL #12  ; data address (0x600000)
LDR X0, [X1]        ; load value
ADD X8, XZR, #93    ; syscall number for exit
SVC #0              ; exit with loaded value
```

## Implementation Note

The integration tests generate these ELF binaries dynamically using Go's binary encoding. This approach:

1. Avoids the need for cross-compilation toolchains
2. Tests the full pipeline: ELF loading → emulation → syscall handling
3. Verifies correct encoding and decoding of ARM64 instructions

See `integration/integration_test.go` for the implementation details.
