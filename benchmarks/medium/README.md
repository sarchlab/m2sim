# Medium-Sized Benchmarks for M2Sim

Medium-complexity benchmarks (100-1000 lines) serving as stepping stones between microbenchmarks and full SPEC benchmarks.

## Purpose

- Bridge the gap between simple microbenchmarks and complex SPEC workloads
- Enable H3 calibration with predictable compute + cache patterns
- Test instruction mix and memory access patterns at medium scale
- Validate emulator functionality with manageable verification complexity

## Requirements

All medium benchmarks follow these conventions:
- **Size**: 100-1000 lines of source code
- **Syscalls**: Only `exit` and `write` (minimal system dependencies)
- **Build**: Static ARM64 Linux ELF via `aarch64-linux-musl-gcc -static -O2`
- **Output**: Print deterministic checksum to stdout for verification
- **Verification**: Include basic correctness checks

## Available Benchmarks

### matmul - Integer Matrix Multiplication

- **Description**: 100x100 integer matrix multiply with deterministic patterns
- **Target**: Integer ALU throughput + cache locality testing
- **Characteristics**: Triple-nested loops, predictable memory access patterns
- **Expected checksum**: Deterministic based on input pattern
- **Build**: `make matmul`

## Build Instructions

```bash
# Build all benchmarks
make

# Build specific benchmark
make matmul

# Clean
make clean
```

## Integration with M2Sim

Built binaries are ARM64 Linux ELF files ready for M2Sim emulation:

```bash
# Run in M2Sim emulator
go run ./cmd/m2sim/main.go benchmarks/medium/matmul
```

Benchmarks are registered in the timing harness for automated execution and validation.

## Current Limitations

**Note**: The compiled benchmark currently uses SIMD instructions (`dup v0.16b`) in the libc memset function that are not yet implemented in M2Sim's SIMD instruction decoder. This will be resolved as part of ongoing SIMD instruction support expansion.

For immediate testing, benchmarks should target instruction sets listed in `SUPPORTED.md` or use manual memory operations instead of libc functions that generate unsupported SIMD code.