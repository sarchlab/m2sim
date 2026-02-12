# M2Sim: Cycle-Accurate Apple M2 CPU Simulator

[![Build Status](https://github.com/sarchlab/m2sim/workflows/CI/badge.svg)](https://github.com/sarchlab/m2sim/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/sarchlab/m2sim)](https://goreportcard.com/report/github.com/sarchlab/m2sim)
[![License](https://img.shields.io/github/license/sarchlab/m2sim.svg)](LICENSE)

**M2Sim** is a cycle-accurate simulator for the Apple M2 CPU, built on the [Akita simulation framework](https://github.com/sarchlab/akita). M2Sim enables detailed performance analysis of ARM64 workloads on Apple Silicon architectures.

## Project Status: In Progress

The simulator is functional with emulation and timing simulation modes. Accuracy validation is ongoing via CI benchmarks.

| Component | Status |
|-----------|--------|
| **Functional Emulation** | ARM64 user-space execution working |
| **Timing Model** | Configurable pipeline with cache hierarchy |
| **Modular Design** | Separate functional/timing layers |
| **Benchmark Suite** | 18 benchmarks (accuracy under verification) |

## Quick Start

### Prerequisites
- Go 1.21 or later
- ARM64 cross-compiler (`aarch64-linux-musl-gcc`)
- Python 3.8+ (for analysis tools)

### Installation
```bash
# Clone the repository
git clone https://github.com/sarchlab/m2sim.git
cd m2sim

# Build the simulator
go build ./...

# Run tests
ginkgo -r

# Build main binary
go build -o m2sim ./cmd/m2sim
```

### Basic Usage
```bash
# Functional emulation only
./m2sim -elf benchmarks/arithmetic.elf

# Cycle-accurate timing simulation
./m2sim -elf benchmarks/arithmetic.elf -timing

# Fast timing approximation
./m2sim -elf benchmarks/arithmetic.elf -fasttiming
```

### Reproduce Paper Results
```bash
# Run complete experimental validation
python3 reproduce_experiments.py

# Generate figures for paper
python3 paper/generate_figures.py

# Compile LaTeX paper
cd paper && pdflatex m2sim_micro2026.tex
```

## Performance Results

Accuracy validation is in progress. Results will be published once CI-based benchmark runs are verified end-to-end. See `.github/workflows/polybench-segmented.yml` for the benchmark CI configuration.

## Architecture Overview

### Simulator Components

```
M2Sim Architecture
├── Functional Emulator (emu/)     # ARM64 instruction execution
│   ├── Decoder                    # 200+ ARM64 instructions
│   ├── Register File              # ARM64 register state
│   └── Syscall Interface          # Linux syscall emulation
├── Timing Model (timing/)         # Cycle-accurate performance
│   ├── Pipeline                   # Configurable superscalar, 5-stage
│   ├── Cache Hierarchy            # L1I (192KB), L1D (128KB), L2 (24MB)
│   └── Branch Prediction          # Tournament predictor (bimodal + gshare)
└── Integration Layer              # ELF loading, measurement framework
```

### Pipeline Configuration (Defaults)
- **Architecture:** Configurable superscalar (default 1-wide, up to 8-wide), in-order execution
- **Stages:** Fetch → Decode → Execute → Memory → Writeback
- **Branch Predictor:** Tournament (bimodal + gshare), 12-cycle misprediction penalty
- **Cache Hierarchy:** L1I (192KB, 6-way, 1-cycle hit), L1D (128KB, 8-way, 4-cycle hit), L2 (24MB, 16-way, 12-cycle hit)
- **Execution Constraints:** Up to 6 ALU ports, 3 load ports, 2 store ports, 4 register write ports (M2 Avalanche modeling)

## Project Structure

```
m2sim/
├── cmd/m2sim/                 # Main simulator binary
├── emu/                       # Functional ARM64 emulator
├── timing/                    # Cycle-accurate timing model
│   ├── core/                  # CPU core timing
│   ├── cache/                 # Cache hierarchy
│   ├── pipeline/              # Pipeline implementation
│   └── latency/               # Instruction latencies
├── benchmarks/                # Validation benchmark suite
│   ├── microbenchmarks/       # Targeted stress tests
│   └── polybench/            # Linear algebra kernels
├── docs/                      # Documentation
│   ├── reference/             # Core technical references
│   ├── development/           # Historical development docs
│   └── archive/               # Archived analysis
├── results/                   # Experimental results
│   ├── final/                 # Completion reports
│   └── baselines/             # Hardware measurement data
├── paper/                     # Research paper and figures
└── reproduce_experiments.py   # Complete reproducibility script
```

## Research Usage

### Adding New Benchmarks

1. **Compile to ARM64 ELF:**
   ```bash
   aarch64-linux-musl-gcc -static -O2 -o benchmark.elf benchmark.c
   ```

2. **Collect Hardware Baseline:**
   ```python
   # Use multi-scale regression methodology
   # Measure at multiple input sizes: 100, 500, 1K, 5K, 10K instructions
   # Apply linear regression: y = mx + b (m = per-instruction latency)
   ```

3. **Run Simulation:**
   ```bash
   ./m2sim -elf benchmark.elf -timing -limit 100000
   ```

4. **Calculate Error:**
   ```
   error = |t_sim - t_real| / min(t_sim, t_real)
   ```

### Extending the Simulator

**Multi-Core Support:** Framework ready for cache coherence and shared memory
**SIMD Enhancement:** Detailed vector pipeline for improved accuracy
**Out-of-Order:** Register renaming for arithmetic co-issue
**Power Modeling:** Leverage M2's efficiency characteristics

## Validation Methodology

### Hardware Baseline Collection
- **Platform:** Apple M2 MacBook Air (2022)
- **Measurement:** 15 runs per data point, trimmed mean
- **Regression:** Multi-scale linear fitting (R² > 0.999 required)
- **Validation:** Statistical confidence intervals

### Benchmark Suite Design
- **Microbenchmarks:** Target individual architectural features
- **PolyBench:** Intermediate-complexity linear algebra kernels
- **Coverage:** Arithmetic, memory, branches, SIMD, dependencies

### Error Analysis
- **Formula:** Symmetric relative error measurement
- **Target:** <20% average error across benchmark suite
- **Categories:** Excellent (<10%), Good (10-20%), Acceptable (20-30%)

## Documentation

### Core References
- **[Architecture Guide](docs/reference/architecture.md)** - M2 microarchitecture research
- **[Timing Guide](docs/reference/timing-guide.md)** - Performance modeling details
- **[Build Setup](docs/reference/build-setup.md)** - Cross-compilation and environment
- **[Calibration Reference](docs/reference/calibration.md)** - Parameter tuning guide

### Research Papers
- **[MICRO 2026 Paper](paper/m2sim_micro2026.pdf)** - Complete technical description
- **[Project Report](results/final/project_report.md)** - Comprehensive completion analysis
- **[Accuracy Validation](results/final/accuracy_validation.md)** - Detailed experimental results

### Development History
- **[Development Docs](docs/development/)** - Research and analysis from development
- **[Historical Reports](results/archive/)** - Evolution of accuracy and methodology

## Milestones

- **H1:** Core simulator with pipeline timing and cache hierarchy
- **H2:** SPEC benchmark enablement with syscall coverage
- **H3:** Microbenchmark calibration
- **H4:** Multi-core analysis framework
- **H5:** Intermediate benchmarks (PolyBench suite)

## Development

### Building from Source
```bash
# Development build with all checks
go build ./...
golangci-lint run ./...
ginkgo -r

# Performance profiling
go build -o profile ./cmd/profile
./profile -elf benchmark.elf -cpuprofile cpu.prof
```

### Contributing
1. **Read:** [CLAUDE.md](CLAUDE.md) for development guidelines
2. **Test:** Ensure all tests pass and lint checks succeed
3. **Document:** Update relevant documentation for changes
4. **Validate:** Verify accuracy on affected benchmarks

## Citation

If you use M2Sim in your research, please cite:

```bibtex
@inproceedings{m2sim2026,
  title={M2Sim: Cycle-Accurate Apple M2 CPU Simulation},
  author={M2Sim Team},
  booktitle={Proceedings of the 59th IEEE/ACM International Symposium on Microarchitecture},
  year={2026},
  organization={IEEE/ACM}
}
```

## Related Projects

- **[Akita](https://github.com/sarchlab/akita)** - Underlying simulation framework
- **[MGPUSim](https://github.com/sarchlab/mgpusim)** - GPU simulator using Akita
- **[SARCH Lab](https://github.com/sarchlab)** - Computer architecture research

## Support

- **Issues:** [GitHub Issues](https://github.com/sarchlab/m2sim/issues)
- **Documentation:** [Project Wiki](https://github.com/sarchlab/m2sim/wiki)
- **Research:** Contact [SARCH Lab](https://github.com/sarchlab)

## License

This project is developed by the [SARCH Lab](https://github.com/sarchlab) at [University/Institution].

---

**M2Sim** - Enabling Apple Silicon research through cycle-accurate simulation.

*Last updated: February 2026*