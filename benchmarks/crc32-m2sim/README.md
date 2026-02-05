# CRC32 Benchmark for M2Sim

Port of Embench-IoT crc32 benchmark for bare-metal ARM64 timing simulation.

## Build

```bash
./build.sh
```

Produces `crc32_m2sim.elf`.

## Source

From [Embench-IoT](https://github.com/embench/embench-iot) `src/crc32/`.

CRC-32 checksum computation using ANSI X3.66 polynomial. Tests bit manipulation and memory streaming patterns.
