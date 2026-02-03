# Native M2 Calibration Benchmarks

Native ARM64 assembly programs for calibrating M2Sim's timing model against real Apple M2 hardware.

## Overview

These benchmarks are 1:1 translations of the microbenchmarks in `benchmarks/microbenchmarks.go`. They allow direct comparison between simulator and real hardware performance.

### Short Benchmarks (~20 instructions)

| Benchmark | Description | Expected Exit Code |
|-----------|-------------|-------------------|
| `arithmetic_sequential` | 20 independent ADDs (ALU throughput) | 4 |
| `dependency_chain` | 20 dependent ADDs (forwarding latency) | 20 |
| `memory_sequential` | 10 store/load pairs (cache performance) | 42 |
| `function_calls` | 5 BL/RET pairs (call overhead) | 5 |
| `branch_taken` | 5 unconditional branches | 5 |
| `mixed_operations` | Mix of ALU, memory, calls | 100 |

### Long-Running Benchmarks (10M iterations)

These benchmarks run 10 million iterations to produce execution times that significantly exceed process startup overhead (~18ms). This enables meaningful timing comparisons.

| Benchmark | Description | Expected Exit Code |
|-----------|-------------|-------------------|
| `arithmetic_sequential_long` | 10M × 20 independent ADDs | 0 |
| `dependency_chain_long` | 10M × 20 dependent ADDs (RAW chain) | 0 |
| `memory_sequential_long` | 10M × store/load cycles | 128 |
| `branch_taken_long` | 10M × predictable branches | 128 |
| `mixed_operations_long` | 10M × mixed instruction types | 128 |

**Why long benchmarks?** The calibration report (docs/calibration-report.md) identified that short benchmarks are dominated by ~18ms process startup overhead. The long benchmarks produce measurable execution times (30-80ms) allowing timing-based calibration.

## Requirements

- Apple Silicon Mac (M1/M2/M3)
- Xcode Command Line Tools (`xcode-select --install`)

## Building

```bash
cd benchmarks/native
make
```

## Running

```bash
# Run all short benchmarks (shows exit codes)
make run

# Verify short benchmark exit codes match expectations
make verify

# Build long-running benchmarks
make long

# Run long benchmarks with timing
make run-long

# Verify long benchmark exit codes
make verify-long
```

## Benchmark Runner Scripts

Three scripts are provided for automated performance measurement:

### run_benchmarks.sh - Basic Timing Collection

Runs benchmarks multiple times and estimates CPU cycles from execution time.

```bash
# Human-readable output
./run_benchmarks.sh

# JSON output for automation
./run_benchmarks.sh --json > native_results.json

# Options
./run_benchmarks.sh --iterations 200      # More iterations for accuracy
./run_benchmarks.sh --benchmark dependency_chain  # Single benchmark
```

Output includes:
- Execution time (avg, min, max, stddev)
- Estimated CPU cycles (based on 3.5 GHz M2 P-core)
- CPI (Cycles Per Instruction)
- Exit code validation

### run_benchmarks_xctrace.sh - Accurate Cycle Counts

Uses Apple Instruments (xctrace) for hardware performance counter access.

```bash
# Collect actual CPU cycle counts
./run_benchmarks_xctrace.sh

# JSON output
./run_benchmarks_xctrace.sh --json > native_results.json
```

Note: Requires Xcode Command Line Tools and may need Terminal to have
Full Disk Access in System Settings.

### compare_with_simulator.sh - Accuracy Analysis

Compares native M2 results with M2Sim simulator output.

```bash
# Run comparison (collects both native and simulator data)
./compare_with_simulator.sh
```

Output:
- Side-by-side CPI comparison
- Error percentage for each benchmark
- Overall calibration assessment

## Manual Performance Data Collection

### Method 1: /usr/bin/time (Basic)

```bash
/usr/bin/time -l ./dependency_chain
```

Shows wall clock time, user/system time, and memory stats.

### Method 2: Instruments (Recommended)

Apple's Instruments provides access to CPU performance counters including cycle counts.

#### Using Xcode Instruments GUI:

1. Open Instruments: `open -a Instruments`
2. Choose "Time Profiler" or "CPU Counters" template
3. File → Record Options → Set target to your benchmark binary
4. Add PMC events: Cycles, Instructions, Branch Mispredictions, L1D Cache Misses
5. Record and analyze

#### Using xctrace CLI:

```bash
# Record CPU counters for a benchmark
xctrace record --template 'CPU Counters' --output trace.trace --launch -- ./dependency_chain

# Export to readable format
xctrace export --input trace.trace --output trace.xml
```

### Method 3: powermetrics (System-wide)

```bash
# Must run as root - shows system-wide CPU counters
sudo powermetrics --samplers cpu_power -i 100
```

### Method 4: Sample Script for Batch Collection

```bash
#!/bin/bash
# collect_stats.sh - Run benchmark N times and collect stats

BENCHMARK=$1
ITERATIONS=${2:-100}

echo "Benchmark: $BENCHMARK"
echo "Iterations: $ITERATIONS"
echo ""

total_ns=0
for i in $(seq 1 $ITERATIONS); do
    # Use gdate for nanosecond precision (install: brew install coreutils)
    start=$(gdate +%s%N)
    ./$BENCHMARK
    end=$(gdate +%s%N)
    elapsed=$((end - start))
    total_ns=$((total_ns + elapsed))
done

avg_ns=$((total_ns / ITERATIONS))
avg_us=$((avg_ns / 1000))

echo "Average time: ${avg_us} microseconds"
```

## Interpreting Results

### What to Compare

1. **CPI (Cycles Per Instruction)**: Key metric for timing model accuracy
   - Run simulator: `go run ./cmd/m2sim benchmark --json`
   - Run native: Collect cycles via Instruments
   - Compare CPI for each benchmark

2. **Relative Performance**: Which benchmarks are faster/slower?
   - If simulator shows dependency_chain 2x slower than arithmetic_sequential
   - Real hardware should show similar ratio

### Expected M2 Characteristics

Based on published data and testing:

| Characteristic | Expected Value |
|---------------|----------------|
| P-core frequency | 3.5 GHz |
| ALU latency (independent) | ~1 cycle |
| ALU latency (dependent) | ~1 cycle with forwarding |
| L1D hit latency | ~4 cycles |
| Branch misprediction penalty | ~12-14 cycles |
| BL/RET overhead | ~1-2 cycles (predicted) |

## Comparing with Simulator

Run the simulator benchmarks:

```bash
cd ../..
go run ./cmd/m2sim benchmark --json > sim_results.json
```

Then compare CPI values:

```bash
# Example comparison workflow (pseudo-code)
# sim_cpi = sim_results[benchmark]["cycles"] / sim_results[benchmark]["instructions"]  
# hw_cpi = hardware_cycles / hardware_instructions
# error = abs(sim_cpi - hw_cpi) / hw_cpi * 100
```

## Troubleshooting

### "This Makefile requires ARM64"
You're on an Intel Mac. These benchmarks require Apple Silicon.

### Permission denied for Instruments
Grant Terminal.app "Full Disk Access" in System Preferences → Privacy & Security.

### xctrace not found
Install Xcode Command Line Tools: `xcode-select --install`

## Files

```
benchmarks/native/
├── Makefile                      # Build system
├── README.md                     # This file
├── run_benchmarks.sh             # Timing-based benchmark runner
├── run_benchmarks_xctrace.sh     # xctrace-based cycle counter
├── compare_with_simulator.sh     # Native vs simulator comparison
│
│   # Short benchmarks (~20 instructions)
├── arithmetic_sequential.s       # ALU throughput test
├── dependency_chain.s            # RAW hazard test  
├── memory_sequential.s           # Cache/memory test
├── function_calls.s              # BL/RET overhead test
├── branch_taken.s                # Branch overhead test
├── mixed_operations.s            # Realistic workload test
│
│   # Long-running benchmarks (10M iterations)
├── arithmetic_sequential_long.s  # 10M × independent ADDs
├── dependency_chain_long.s       # 10M × dependent ADDs (RAW chain)
├── memory_sequential_long.s      # 10M × memory operations
├── branch_taken_long.s           # 10M × predictable branches
└── mixed_operations_long.s       # 10M × mixed instruction types
```
