# Matrix Multiply Integer Benchmark for M2Sim

Port of Embench-IoT matmult-int benchmark for bare-metal ARM64 timing simulation.

## Build

```bash
./build.sh
```

Produces `matmult-int_m2sim.elf`.

## Source

From [Embench-IoT](https://github.com/embench/embench-iot) `src/matmult-int/`.

Integer matrix multiplication. Tests nested loops, array indexing, and cache locality.
