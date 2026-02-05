#!/bin/bash
# [Eric] Quick calibration script using microbenchmarks
# Runs in seconds, use for rapid accuracy iteration
# Created: 2026-02-05

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "=== M2Sim Quick Calibration ==="
echo "Date: $(date)"
echo ""
echo "Running microbenchmark accuracy tests..."
echo ""

# Run accuracy tests (fast - <1 second total)
go test -v -run "TestAccuracyAgainstBaseline" ./benchmarks/ 2>&1

echo ""
echo "=== Quick calibration complete ==="
echo ""
echo "NOTE: For full Embench timing, run overnight:"
echo "  nohup ./scripts/batch-timing.sh > reports/timing.log 2>&1 &"
