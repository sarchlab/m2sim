# Alternative ARM64 Benchmarks (Non-SPEC)

*Research by Eric — 2026-02-04*

While waiting for SPEC CPU 2017 installation to be unblocked, we can validate M2Sim accuracy using these alternative benchmark suites.

## Recommended: CoreMark

**Website:** https://github.com/eembc/coremark

**Why:**
- Industry standard CPU benchmark
- Extremely simple to build (just `make`)
- Single-file output, easy to measure
- Tests CPU pipeline, integer performance
- Freely available under Apache 2.0 license

**Build on macOS ARM64:**
```bash
git clone https://github.com/eembc/coremark.git
cd coremark
make PORT_DIR=macos
# Or simply: make (uses default linux64 port)
```

**Characteristics:**
- Focuses on: list processing, matrix operations, state machines, CRC
- Single-threaded integer workload
- Execution time: configurable iterations
- Perfect for our microbenchmark → intermediate gap

## Recommended: Embench-IoT

**Website:** https://github.com/embench/embench-iot

**Why:**
- Modern benchmark suite (successor to Dhrystone/CoreMark for embedded)
- Free and open source
- Diverse workloads: compression, crypto, signal processing
- Designed by Dave Patterson's group

**Build:**
```bash
git clone https://github.com/embench/embench-iot.git
cd embench-iot
./build_all.py --arch aarch64
```

**Workloads include:**
- aha-mont64 (modular arithmetic)
- crc32 (checksum)
- huffbench (compression)
- matmult-int (matrix multiply)
- minver (matrix inversion)
- nbody (physics simulation)
- nettle-aes (crypto)
- primecount, qrduino, sglib-combined, statemate, tarfind

## Alternative: LLVM Test Suite

**Website:** https://llvm.org/docs/TestSuiteGuide.html

**Note:** More complex setup, requires LLVM infrastructure. May be overkill for our needs, but has extensive single-source benchmarks.

## Comparison

| Suite | Build Complexity | License | ARM64 Support | Use Case |
|-------|------------------|---------|---------------|----------|
| CoreMark | Very Easy | Apache 2.0 | ✅ | Quick validation |
| Embench | Easy | GPL/BSD mix | ✅ | Diverse workloads |
| LLVM Test | Complex | Apache 2.0 | ✅ | Comprehensive |
| SPEC | Complex | Commercial | ✅ | Industry standard |

## Recommendation

**Immediate action:** Build and run CoreMark on both real M2 and M2Sim.
- Quick to set up (minutes, not hours)
- Provides intermediate-complexity workload
- Results are comparable to industry standards

**Next:** Evaluate Embench-IoT for additional workload diversity.

**Long-term:** Complete SPEC setup once human unblocks Gatekeeper issue.
