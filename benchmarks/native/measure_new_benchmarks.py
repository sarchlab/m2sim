#!/usr/bin/env python3
"""
Measure M2 hardware timing for the 4 new microbenchmarks.
These are fixed-size benchmarks (not parameterized like the calibration template).
"""

import json
import statistics
import subprocess
import time
from pathlib import Path

# Benchmark configurations
BENCHMARKS = {
    "memory_strided": {
        "description": "10 store/load pairs with stride-4 access (strided memory pattern)",
        "instructions_per_iteration": 20,  # 10 store/load pairs = 20 instructions
        "executable": "./memory_strided",
        "expected_exit_code": 7
    },
    "load_heavy": {
        "description": "20 load instructions per iteration (load-heavy workload)",
        "instructions_per_iteration": 20,
        "executable": "./load_heavy",
        "expected_exit_code": 20
    },
    "store_heavy": {
        "description": "20 store instructions per iteration (store-heavy workload)",
        "instructions_per_iteration": 20,
        "executable": "./store_heavy",
        "expected_exit_code": 3
    },
    "branch_heavy": {
        "description": "20 branch instructions per iteration (branch-heavy workload)",
        "instructions_per_iteration": 20,
        "executable": "./branch_heavy",
        "expected_exit_code": 10
    }
}

def run_benchmark(executable: str, expected_exit_code: int, runs: int = 15, warmup: int = 3) -> list:
    """Run benchmark multiple times and return execution times in seconds."""
    times = []

    # Warmup runs (discarded)
    for _ in range(warmup):
        start = time.perf_counter()
        result = subprocess.run([executable], capture_output=True)
        end = time.perf_counter()
        # Don't check exit code during warmup

    # Actual measurement runs
    for _ in range(runs):
        start = time.perf_counter()
        result = subprocess.run([executable], capture_output=True)
        end = time.perf_counter()

        if result.returncode != expected_exit_code:
            raise Exception(f"Benchmark {executable} failed with exit code {result.returncode}, expected {expected_exit_code}")

        times.append(end - start)

    return times

def measure_benchmark(name: str, config: dict) -> dict:
    """Measure a single benchmark and return timing data."""
    print(f"\n{'='*60}")
    print(f"Measuring: {name}")
    print(f"Description: {config['description']}")
    print(f"{'='*60}")

    times = run_benchmark(config["executable"], config["expected_exit_code"])

    # Remove outliers (20% trimmed mean)
    times_sorted = sorted(times)
    trim_count = int(len(times_sorted) * 0.1)  # Remove 10% from each end
    if trim_count > 0:
        trimmed_times = times_sorted[trim_count:-trim_count]
    else:
        trimmed_times = times_sorted

    # Calculate statistics
    mean_time_s = statistics.mean(trimmed_times)
    mean_time_ms = mean_time_s * 1000
    std_time_ms = statistics.stdev(trimmed_times) * 1000

    # Calculate per-instruction latency
    instructions = config["instructions_per_iteration"]
    latency_ns_per_instruction = (mean_time_s * 1e9) / instructions

    # Calculate CPI at 3.5 GHz
    frequency_ghz = 3.5
    cpi = latency_ns_per_instruction * frequency_ghz / 1000
    ipc = 1.0 / cpi if cpi > 0 else float('inf')

    print(f"  Runs: {len(trimmed_times)} (after outlier removal)")
    print(f"  Mean time: {mean_time_ms:.2f} ms (±{std_time_ms:.2f})")
    print(f"  Per-instruction latency: {latency_ns_per_instruction:.4f} ns")
    print(f"  CPI @ 3.5 GHz: {cpi:.3f}")
    print(f"  IPC @ 3.5 GHz: {ipc:.2f}")

    return {
        "name": name,
        "description": config["description"],
        "instructions_per_iteration": instructions,
        "latency_ns_per_instruction": latency_ns_per_instruction,
        "cpi_at_3_5_ghz": cpi,
        "ipc_at_3_5_ghz": ipc,
        "mean_time_ms": mean_time_ms,
        "std_time_ms": std_time_ms,
        "runs_used": len(trimmed_times)
    }

def main():
    print("M2 Hardware Baseline Measurement for New Microbenchmarks")
    print("=" * 70)

    results = []
    for name, config in BENCHMARKS.items():
        try:
            result = measure_benchmark(name, config)
            results.append(result)
        except Exception as e:
            print(f"Error measuring {name}: {e}")
            continue

    # Summary
    print(f"\n{'='*70}")
    print("MEASUREMENT RESULTS SUMMARY")
    print(f"{'='*70}")
    print(f"{'Benchmark':<15} {'Latency (ns)':<15} {'CPI':<8} {'IPC':<8}")
    print("-" * 60)

    for result in results:
        print(f"{result['name']:<15} {result['latency_ns_per_instruction']:<15.4f} "
              f"{result['cpi_at_3_5_ghz']:<8.3f} {result['ipc_at_3_5_ghz']:<8.2f}")

    # Save results
    output_file = "new_benchmarks_m2_baseline.json"
    with open(output_file, 'w') as f:
        json.dump({
            "metadata": {
                "version": "1.0",
                "date": "2026-02-07",
                "hardware": "Apple M2 (P-core @ 3.5 GHz)",
                "methodology": "Direct timing with outlier removal",
                "benchmarks_measured": len(results)
            },
            "baselines": results
        }, f, indent=2)

    print(f"\nResults saved to: {output_file}")

if __name__ == "__main__":
    main()