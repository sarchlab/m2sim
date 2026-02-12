#!/usr/bin/env python3
"""
polybench_calibration.py - Linear Regression Calibration for PolyBench

Uses varying kernel repetition counts to separate process startup overhead
from actual per-instruction latency via linear regression — the same
methodology as linear_calibration.py for microbenchmarks.

This replaces the broken single-point methodology that produced baselines
of 7,000+ ns/inst (should be ~0.3 ns/inst).

Approach:
  1. Build each PolyBench kernel with N repetitions (calls kernel N times)
  2. Measure instruction count and wall-clock time for each N
  3. Fit linear regression: time_ms = slope * instruction_count / 1e6 + overhead
  4. slope (ns/instruction) = hardware baseline latency
"""

import json
import os
import subprocess
import sys
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import Dict, List, Optional, Tuple

POLYBENCH_BENCHMARKS = {
    "gemm": "General matrix multiply (GEMM) - C := alpha*A*B + beta*C",
    "atax": "Matrix transpose and vector multiply - y = A^T * (A * x)",
    "2mm": "Two matrix multiplications - D = A*B; E = C*D",
    "mvt": "Matrix vector product and transpose",
    "jacobi-1d": "1D Jacobi iterative stencil computation",
    "3mm": "Three matrix multiplications",
    "bicg": "BiCG sub-kernel of BiCGStab linear solver",
}

# O(n^3) benchmarks need fewer reps (each rep has ~2.4M instructions at SMALL)
# O(n^2) benchmarks need more reps (each rep has ~85K instructions at SMALL)
CUBIC_BENCHMARKS = {"gemm", "2mm", "3mm"}
REP_COUNTS_CUBIC = [10, 50, 100, 500, 1000, 5000]
REP_COUNTS_QUADRATIC = [100, 500, 1000, 5000, 10000, 50000]


def get_polybench_dir() -> Path:
    return Path(__file__).parent.parent / "polybench"


def build_benchmark(bench: str, reps: int) -> Optional[str]:
    """Build a PolyBench benchmark with given repetition count. Returns binary path."""
    polybench_dir = get_polybench_dir()
    build_script = polybench_dir / "build_native.sh"
    result = subprocess.run(
        ["bash", str(build_script), bench, str(reps)],
        capture_output=True, text=True, cwd=str(polybench_dir),
    )
    if result.returncode != 0:
        print(f"  BUILD ERROR ({bench} r{reps}): {result.stderr.strip()}")
        return None
    binary_path = str(polybench_dir / f"{bench}_native_r{reps}")
    return binary_path if os.path.exists(binary_path) else None


def count_instructions(binary_path: str) -> Optional[int]:
    """Count retired instructions using macOS /usr/bin/time -l."""
    try:
        result = subprocess.run(
            ["/usr/bin/time", "-l", binary_path],
            capture_output=True, text=True,
        )
        for line in result.stderr.split("\n"):
            if "instructions retired" in line:
                return int(line.strip().split()[0])
    except Exception:
        pass
    return None


def run_timed(binary_path: str, runs: int = 15, warmup: int = 3) -> List[float]:
    """Run binary multiple times with warmup, return times in seconds."""
    for _ in range(warmup):
        subprocess.run([binary_path], capture_output=True)
    times = []
    for _ in range(runs):
        start = time.perf_counter()
        subprocess.run([binary_path], capture_output=True)
        end = time.perf_counter()
        times.append(end - start)
    return times


def trimmed_mean(values: List[float], trim_pct: float = 0.2) -> float:
    """Trimmed mean, removing top/bottom trim_pct."""
    if len(values) < 3:
        return sum(values) / len(values)
    s = sorted(values)
    n = len(s)
    tc = int(n * trim_pct)
    trimmed = s[tc:-tc] if tc > 0 else s
    return sum(trimmed) / len(trimmed) if trimmed else sum(s) / n


def linear_regression(x: List[float], y: List[float]) -> Tuple[float, float, float]:
    """Returns (slope, intercept, r_squared)."""
    try:
        from scipy import stats
        slope, intercept, r, _, _ = stats.linregress(x, y)
        return slope, intercept, r ** 2
    except ImportError:
        pass
    n = len(x)
    sx = sum(x); sy = sum(y)
    sxy = sum(a * b for a, b in zip(x, y))
    sx2 = sum(a * a for a in x)
    d = n * sx2 - sx * sx
    if abs(d) < 1e-15:
        return 0.0, sy / n if n else 0.0, 0.0
    slope = (n * sxy - sx * sy) / d
    intercept = (sy - slope * sx) / n
    ym = sy / n
    ss_tot = sum((yi - ym) ** 2 for yi in y)
    ss_res = sum((yi - (slope * xi + intercept)) ** 2 for xi, yi in zip(x, y))
    r2 = 1 - (ss_res / ss_tot) if ss_tot > 0 else 0
    return slope, intercept, r2


@dataclass
class CalibrationResult:
    benchmark: str
    description: str
    instruction_latency_ns: float
    overhead_ms: float
    r_squared: float
    data_points: List[Dict] = field(default_factory=list)


def calibrate_benchmark(
    bench: str, rep_counts: List[int], runs: int = 15, verbose: bool = True
) -> Optional[CalibrationResult]:
    """Calibrate one benchmark using varying repetition counts."""
    desc = POLYBENCH_BENCHMARKS[bench]
    if verbose:
        print(f"\n{'='*60}")
        print(f"Calibrating: {bench}")
        print(f"Description: {desc}")
        print(f"{'='*60}")

    data_points = []
    instr_list = []
    time_list = []

    for reps in rep_counts:
        if verbose:
            print(f"  reps={reps:>6}... ", end="", flush=True)

        binary = build_benchmark(bench, reps)
        if not binary:
            if verbose:
                print("BUILD FAILED")
            continue

        insts = count_instructions(binary)
        if insts is None:
            if verbose:
                print("INSTR COUNT FAILED")
            continue

        run_times = run_timed(binary, runs=runs, warmup=3)
        run_times_ms = [t * 1000 for t in run_times]
        avg_ms = trimmed_mean(run_times_ms)

        s = sorted(run_times_ms)
        tc = int(len(s) * 0.2)
        trimmed = s[tc:-tc] if tc > 0 else s
        std_ms = (sum((t - avg_ms) ** 2 for t in trimmed) / len(trimmed)) ** 0.5

        if verbose:
            print(f"{insts:>12,} insts, {avg_ms:8.2f} ms (±{std_ms:.2f})")

        data_points.append({
            "reps": reps,
            "instructions": insts,
            "time_ms": avg_ms,
        })
        instr_list.append(insts)
        time_list.append(avg_ms)

    if len(data_points) < 3:
        if verbose:
            print(f"  FAIL: Need ≥3 data points, got {len(data_points)}")
        return None

    slope, intercept, r2 = linear_regression(instr_list, time_list)
    latency_ns = slope * 1e6  # ms/instruction -> ns/instruction

    if verbose:
        cpi = latency_ns * 3.5
        print(f"\n  Latency: {latency_ns:.4f} ns/inst (CPI={cpi:.2f} @ 3.5 GHz)")
        print(f"  Overhead: {intercept:.2f} ms")
        print(f"  R² = {r2:.6f}")

    return CalibrationResult(
        benchmark=bench,
        description=desc,
        instruction_latency_ns=latency_ns,
        overhead_ms=intercept,
        r_squared=r2,
        data_points=data_points,
    )


def update_combined_calibration(results: List[CalibrationResult]):
    """Replace PolyBench entries in combined calibration with corrected values."""
    combined_path = Path(__file__).parent / "calibration_results.json"
    if not combined_path.exists():
        return

    with open(combined_path) as f:
        combined = json.load(f)

    polybench_names = {r.benchmark for r in results}
    combined["results"] = [
        e for e in combined["results"] if e["benchmark"] not in polybench_names
    ]
    for r in results:
        combined["results"].append({
            "benchmark": r.benchmark,
            "description": r.description,
            "calibrated": True,
            "instruction_latency_ns": r.instruction_latency_ns,
            "overhead_ms": r.overhead_ms,
            "r_squared": r.r_squared,
            "data_points": [
                {"instructions": d["instructions"], "time_ms": d["time_ms"]}
                for d in r.data_points
            ],
        })

    combined["methodology"] = "combined_h5_calibration"
    combined["formula"] = (
        "All benchmarks use linear regression: "
        "time_ms = latency_ns * instruction_count / 1e6 + overhead_ms"
    )
    combined.pop("sources", None)
    combined_path.write_text(json.dumps(combined, indent=2))
    print(f"Updated combined calibration: {combined_path}")


def main():
    import argparse

    parser = argparse.ArgumentParser(
        description="PolyBench Linear Regression Calibration (Issue #466)"
    )
    parser.add_argument(
        "--benchmarks", nargs="*", default=None,
        help="Benchmarks to calibrate (default: all)",
    )
    parser.add_argument(
        "--runs", type=int, default=15,
        help="Timed runs per data point (default: 15)",
    )
    parser.add_argument(
        "--output", type=str, default=None,
        help="Output JSON path",
    )
    args = parser.parse_args()

    print("=" * 70)
    print("PolyBench Linear Regression Calibration Tool")
    print("Methodology: Varying kernel repetitions (Issue #466)")
    print("=" * 70)

    benchmarks = args.benchmarks or list(POLYBENCH_BENCHMARKS.keys())
    for name in benchmarks:
        if name not in POLYBENCH_BENCHMARKS:
            print(f"Error: unknown benchmark '{name}'")
            sys.exit(1)

    results = []
    for bench in benchmarks:
        reps = REP_COUNTS_CUBIC if bench in CUBIC_BENCHMARKS else REP_COUNTS_QUADRATIC
        r = calibrate_benchmark(bench, reps, runs=args.runs)
        if r:
            results.append(r)

    if not results:
        print("\nERROR: No benchmarks calibrated successfully.")
        sys.exit(1)

    # Summary
    print("\n" + "=" * 70)
    print("CALIBRATION RESULTS")
    print("=" * 70)
    print(f"{'Benchmark':<15} {'Latency (ns)':<14} {'CPI @3.5GHz':<12} {'R²':<10}")
    print("-" * 70)
    for r in results:
        cpi = r.instruction_latency_ns * 3.5
        print(f"{r.benchmark:<15} {r.instruction_latency_ns:>11.4f}   "
              f"{cpi:>9.2f}   {r.r_squared:>8.6f}")

    # Save
    output = {
        "methodology": "linear_regression",
        "formula": "time_ms = latency_ns * instruction_count / 1e6 + overhead_ms",
        "source": "polybench_rep_scaling",
        "rep_counts_cubic": REP_COUNTS_CUBIC,
        "rep_counts_quadratic": REP_COUNTS_QUADRATIC,
        "results": [
            {
                "benchmark": r.benchmark,
                "description": r.description,
                "calibrated": True,
                "instruction_latency_ns": r.instruction_latency_ns,
                "overhead_ms": r.overhead_ms,
                "r_squared": r.r_squared,
                "data_points": [
                    {"instructions": d["instructions"], "time_ms": d["time_ms"]}
                    for d in r.data_points
                ],
            }
            for r in results
        ],
    }

    output_path = (
        Path(args.output) if args.output
        else Path(__file__).parent / "polybench_calibration_results.json"
    )
    output_path.write_text(json.dumps(output, indent=2))
    print(f"\nResults saved to: {output_path}")

    update_combined_calibration(results)


if __name__ == "__main__":
    main()
