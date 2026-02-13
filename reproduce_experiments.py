#!/usr/bin/env python3
"""
M2Sim Reproducible Experiments Script

This script reproduces all experiments from the M2Sim paper, including:
1. Building the simulator
2. Running accuracy validation experiments (via accuracy_report.py)
3. Generating figures and analysis
4. Creating the final paper

The accuracy experiments delegate to benchmarks/native/accuracy_report.py,
which runs actual Go test-based timing simulations and compares CPI values
against real M2 hardware calibration baselines.

Usage:
    python3 reproduce_experiments.py [--skip-build] [--skip-experiments] [--skip-figures] [--skip-paper]

Requirements:
    - Go 1.21 or later
    - Python 3.8+ with matplotlib, seaborn, pandas, numpy
    - LaTeX distribution (for paper compilation)
"""

import os
import sys
import subprocess
import time
import json
import argparse
from pathlib import Path
from typing import Dict


# Resolve repo root (directory containing this script)
REPO_ROOT = Path(__file__).resolve().parent


class Colors:
    """ANSI color codes for terminal output"""
    GREEN = '\033[92m'
    BLUE = '\033[94m'
    YELLOW = '\033[93m'
    RED = '\033[91m'
    BOLD = '\033[1m'
    END = '\033[0m'


def log(message: str, level: str = "INFO"):
    """Print colored log message"""
    color_map = {
        "INFO": Colors.BLUE,
        "SUCCESS": Colors.GREEN,
        "WARNING": Colors.YELLOW,
        "ERROR": Colors.RED,
        "HEADER": Colors.BOLD
    }
    color = color_map.get(level, Colors.END)
    timestamp = time.strftime("%H:%M:%S")
    print(f"{color}[{timestamp}] {level}: {message}{Colors.END}")


def run_command(cmd, cwd=None, check=True, timeout=600):
    """Run a command with logging.

    Args:
        cmd: Either a list of arguments or a string (run via shell).
        cwd: Working directory.
        check: Raise on non-zero exit.
        timeout: Seconds before killing the process.
    """
    if isinstance(cmd, str):
        display = cmd
    else:
        display = " ".join(str(c) for c in cmd)
    log(f"Running: {display}")
    if cwd:
        log(f"  cwd: {cwd}")

    use_shell = isinstance(cmd, str)
    try:
        result = subprocess.run(
            cmd,
            cwd=cwd,
            capture_output=True,
            text=True,
            check=check,
            shell=use_shell,
            timeout=timeout,
        )
        if result.stdout.strip():
            for line in result.stdout.strip().split('\n')[:50]:
                log(f"  {line}")
            total_lines = result.stdout.strip().count('\n') + 1
            if total_lines > 50:
                log(f"  ... ({total_lines - 50} more lines)")
        return result
    except subprocess.CalledProcessError as e:
        log(f"Command failed with exit code {e.returncode}", "ERROR")
        if e.stderr:
            for line in e.stderr.strip().split('\n')[:20]:
                log(f"  {line}", "ERROR")
        raise


def check_dependencies():
    """Check required dependencies"""
    log("Checking dependencies...", "HEADER")

    deps = [
        ("go", ["go", "version"]),
        ("python3", ["python3", "--version"]),
    ]

    missing = []
    for dep, cmd in deps:
        try:
            subprocess.run(cmd, capture_output=True, check=True)
            log(f"  {dep} found", "SUCCESS")
        except (subprocess.CalledProcessError, FileNotFoundError):
            log(f"  {dep} not found", "ERROR")
            missing.append(dep)

    if missing:
        log(f"Missing dependencies: {', '.join(missing)}", "ERROR")
        return False
    return True


def build_simulator():
    """Build M2Sim and run short tests to verify."""
    log("Building M2Sim simulator...", "HEADER")

    run_command(["go", "build", "./..."], cwd=REPO_ROOT)
    log("All packages built", "SUCCESS")

    log("Running short tests to verify build...")
    try:
        run_command(["go", "test", "./...", "-short", "-count=1"],
                    cwd=REPO_ROOT, timeout=300)
        log("Tests passed", "SUCCESS")
    except subprocess.CalledProcessError:
        log("Some tests failed - continuing anyway", "WARNING")


def run_accuracy_experiments() -> Dict:
    """Run accuracy experiments by delegating to accuracy_report.py.

    accuracy_report.py:
      - Runs Go test-based timing simulations for all benchmarks
      - Compares simulated CPI against real M2 hardware calibration baselines
      - Generates accuracy_results.json, accuracy_report.md, accuracy_figure.png

    Returns a results dict with 'summary' and 'benchmarks' keys.
    """
    log("Running accuracy validation experiments...", "HEADER")
    log("Delegating to benchmarks/native/accuracy_report.py (runs real simulations)")

    accuracy_script = REPO_ROOT / "benchmarks" / "native" / "accuracy_report.py"
    if not accuracy_script.exists():
        log(f"accuracy_report.py not found at {accuracy_script}", "ERROR")
        sys.exit(1)

    # Run accuracy_report.py — it produces accuracy_results.json
    # Allow long timeout since each benchmark may take minutes
    try:
        run_command(
            [sys.executable, str(accuracy_script)],
            cwd=REPO_ROOT,
            check=False,
            timeout=3600,  # 1 hour for full benchmark suite
        )
    except subprocess.TimeoutExpired:
        log("Accuracy experiments timed out after 1 hour", "ERROR")
        sys.exit(1)

    # Read the JSON results produced by accuracy_report.py
    json_path = REPO_ROOT / "benchmarks" / "native" / "accuracy_results.json"
    if not json_path.exists():
        log(f"accuracy_results.json not found at {json_path}", "ERROR")
        log("accuracy_report.py may have failed to produce results", "ERROR")
        sys.exit(1)

    with open(json_path) as f:
        accuracy_data = json.load(f)

    # Convert to the results format used by this script
    benchmarks = []
    for bench in accuracy_data.get("benchmarks", []):
        benchmarks.append({
            "name": bench["name"],
            "error": bench["error"],
            "sim_cpi": bench.get("sim_cpi", 0),
            "sim_latency_ns": bench.get("sim_latency_ns", 0),
            "real_latency_ns": bench.get("real_latency_ns", 0),
            "calibrated": bench.get("calibrated", True),
            "status": "completed",
        })

    summary = accuracy_data.get("summary", {})
    errors = [b["error"] for b in benchmarks]

    results = {
        "benchmarks": benchmarks,
        "summary": {
            "total_benchmarks": summary.get("benchmark_count", len(benchmarks)),
            "calibrated_benchmarks": summary.get("calibrated_count", len(benchmarks)),
            "average_error": summary.get("average_error", sum(errors) / len(errors) if errors else 0),
            "max_error": summary.get("max_error", max(errors) if errors else 0),
            "min_error": min(errors) if errors else 0,
        }
    }

    log(f"Accuracy validation complete: {len(benchmarks)} benchmarks, "
        f"{results['summary']['average_error'] * 100:.1f}% average error", "SUCCESS")

    # Copy results to repo root for convenience
    with open(REPO_ROOT / "accuracy_results.json", "w") as f:
        json.dump(results, f, indent=2)

    return results


def generate_figures():
    """Generate paper figures using paper/generate_figures.py."""
    log("Generating paper figures...", "HEADER")

    figure_script = REPO_ROOT / "paper" / "generate_figures.py"
    if figure_script.exists():
        try:
            run_command([sys.executable, str(figure_script)],
                        cwd=REPO_ROOT / "paper", timeout=120)
            log("Paper figures generated", "SUCCESS")
        except subprocess.CalledProcessError:
            log("Figure generation failed", "ERROR")
            raise
    else:
        log("Figure generation script not found at paper/generate_figures.py", "WARNING")


def compile_paper():
    """Compile LaTeX paper with bibtex."""
    log("Compiling LaTeX paper...", "HEADER")

    paper_dir = REPO_ROOT / "paper"
    paper_tex = paper_dir / "m2sim_micro2026.tex"
    if not paper_tex.exists():
        log("LaTeX source not found at paper/m2sim_micro2026.tex", "WARNING")
        return

    try:
        # pdflatex → bibtex → pdflatex × 2 (standard LaTeX build)
        run_command(["pdflatex", "-interaction=nonstopmode", "m2sim_micro2026.tex"],
                    cwd=paper_dir, check=False)
        run_command(["bibtex", "m2sim_micro2026"],
                    cwd=paper_dir, check=False)
        run_command(["pdflatex", "-interaction=nonstopmode", "m2sim_micro2026.tex"],
                    cwd=paper_dir, check=False)
        run_command(["pdflatex", "-interaction=nonstopmode", "m2sim_micro2026.tex"],
                    cwd=paper_dir, check=False)

        pdf_path = paper_dir / "m2sim_micro2026.pdf"
        if pdf_path.exists():
            log(f"Paper compiled: {pdf_path}", "SUCCESS")
        else:
            log("PDF not produced — check LaTeX logs", "ERROR")
    except subprocess.CalledProcessError:
        log("LaTeX compilation failed — is a TeX distribution installed?", "WARNING")


def generate_experiment_report(results: Dict):
    """Generate a human-readable experiment report from real results."""
    log("Generating experiment report...", "HEADER")

    summary = results["summary"]
    benchmarks = results["benchmarks"]
    avg_error = summary["average_error"]

    # Determine target achievement
    target_met = avg_error < 0.2
    target_line = (f"{'PASS' if target_met else 'FAIL'}: "
                   f"Average error {avg_error * 100:.1f}% {'<' if target_met else '>='} 20% target")

    report_lines = [
        "# M2Sim Experiment Report",
        "",
        f"**Generated:** {time.strftime('%Y-%m-%d %H:%M:%S')}",
        "",
        "## Summary",
        "",
        f"- **Total Benchmarks:** {summary['total_benchmarks']}",
        f"- **Average Error:** {avg_error:.3f} ({avg_error * 100:.1f}%)",
        f"- **Maximum Error:** {summary['max_error']:.3f} ({summary['max_error'] * 100:.1f}%)",
        f"- **Minimum Error:** {summary['min_error']:.3f} ({summary['min_error'] * 100:.1f}%)",
        "",
        f"## Target: {target_line}",
        "",
        "## Detailed Results",
        "",
        "| Benchmark | Sim CPI | Sim (ns/inst) | Real (ns/inst) | Error | Calibrated |",
        "|-----------|---------|---------------|----------------|-------|------------|",
    ]

    for bench in sorted(benchmarks, key=lambda b: b["name"]):
        cal = "yes" if bench.get("calibrated", True) else "no"
        report_lines.append(
            f"| {bench['name']} | {bench.get('sim_cpi', 0):.3f} | "
            f"{bench.get('sim_latency_ns', 0):.4f} | "
            f"{bench.get('real_latency_ns', 0):.4f} | "
            f"{bench['error'] * 100:.1f}% | {cal} |"
        )

    report_lines.extend([
        "",
        "## Reproduction Environment",
        "",
        f"- **OS:** {os.uname().sysname} {os.uname().release}",
        f"- **Arch:** {os.uname().machine}",
        f"- **Directory:** {REPO_ROOT}",
        "",
        "## How Results Were Obtained",
        "",
        "Accuracy experiments were run by `benchmarks/native/accuracy_report.py`, which:",
        "1. Runs `go test` timing simulations for each benchmark",
        "2. Extracts simulated CPI from test output",
        "3. Compares against real M2 hardware calibration baselines",
        "4. Computes error = abs(t_sim - t_real) / min(t_sim, t_real)",
        "",
    ])

    report_path = REPO_ROOT / "experiment_report.md"
    report_path.write_text('\n'.join(report_lines))
    log(f"Experiment report: {report_path}", "SUCCESS")


def main():
    """Main experiment reproduction workflow"""
    parser = argparse.ArgumentParser(description="Reproduce M2Sim experiments")
    parser.add_argument("--skip-build", action="store_true", help="Skip build phase")
    parser.add_argument("--skip-experiments", action="store_true", help="Skip experiment execution")
    parser.add_argument("--skip-figures", action="store_true", help="Skip figure generation")
    parser.add_argument("--skip-paper", action="store_true", help="Skip paper compilation")
    args = parser.parse_args()

    log("M2Sim Reproducible Experiments", "HEADER")
    log("=" * 40, "HEADER")

    start_time = time.time()

    try:
        if not check_dependencies():
            sys.exit(1)

        # Build
        if not args.skip_build:
            build_simulator()
        else:
            log("Skipping build phase", "WARNING")

        # Experiments
        if not args.skip_experiments:
            results = run_accuracy_experiments()
        else:
            log("Skipping experiments", "WARNING")
            # Try to load previously generated results
            cached = REPO_ROOT / "accuracy_results.json"
            if cached.exists():
                with open(cached) as f:
                    results = json.load(f)
                log(f"Loaded cached results from {cached}")
            else:
                log("No cached results found — run without --skip-experiments first", "ERROR")
                sys.exit(1)

        # Figures
        if not args.skip_figures:
            generate_figures()
        else:
            log("Skipping figure generation", "WARNING")

        # Paper
        if not args.skip_paper:
            compile_paper()
        else:
            log("Skipping paper compilation", "WARNING")

        # Report
        generate_experiment_report(results)

        # Done
        duration = time.time() - start_time
        log("=" * 40, "HEADER")
        log(f"Completed in {duration:.1f}s", "SUCCESS")
        log(f"Average accuracy error: {results['summary']['average_error'] * 100:.1f}%", "SUCCESS")

    except Exception as e:
        log(f"Experiment reproduction failed: {e}", "ERROR")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
