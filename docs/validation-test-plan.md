# Validation Test Plan for M2Sim Accuracy

**Author:** Cathy (QA Agent)  
**Date:** 2026-02-03  
**Status:** DRAFT  

## Objective

Create automated validation tests to verify M2Sim achieves <2% error against real Apple M2 hardware measurements.

## Dependencies

- ✅ Issue #95 (Bob's benchmark harness) - provides simulator timing data
- ✅ Issue #96 (Cathy's M2 baseline) - provides real hardware measurements
- ⏳ Issue #97 (accuracy analysis) - this plan implements

## Test Categories

### 1. Core Accuracy Tests (Go tests in `benchmarks/`)

#### Test: `TestAccuracy_ArithmeticThroughput`
**Purpose:** Verify simulator CPI matches real M2 for independent ALU operations.

```go
// Compare simulator arithmetic_sequential vs baseline arithmetic
// Baseline: CPI = 0.268 (M2 can issue ~4 ADDs per cycle)
// Target: |sim_cpi - 0.268| / 0.268 < 2%
```

**Current gap:** Simulator shows CPI ~1.2 for arithmetic_sequential vs M2's 0.268. This is expected - simulator doesn't model superscalar issue width.

**Action:** Document as known limitation or update timing model.

#### Test: `TestAccuracy_DependencyChain`
**Purpose:** Verify simulator CPI matches real M2 for RAW hazard chains.

```go
// Compare simulator dependency_chain vs baseline dependency
// Baseline: CPI = 1.009 (one add per cycle due to data dependency)
// Target: |sim_cpi - 1.009| / 1.009 < 2%
```

**Current status:** Simulator shows CPI ~2.2 vs M2's 1.009. May indicate forwarding path not modeled accurately.

#### Test: `TestAccuracy_BranchPrediction`
**Purpose:** Verify simulator branch overhead matches real M2.

```go
// Compare simulator branch_taken vs baseline branch
// Baseline: CPI = 1.19 (well-predicted branches minimal penalty)
// Target: |sim_cpi - 1.19| / 1.19 < 2%
```

**Current status:** Simulator shows CPI ~2.9 vs M2's 1.19. Branch prediction model needs calibration.

### 2. Automated Comparison Tool

Create `benchmarks/accuracy_test.go`:

```go
func TestAccuracyAgainstBaseline(t *testing.T) {
    // 1. Load baseline from benchmarks/native/m2_baseline.json
    baseline := loadBaseline("native/m2_baseline.json")
    
    // 2. Run simulator benchmarks (without caches to isolate core timing)
    config := DefaultConfig()
    config.EnableICache = false
    config.EnableDCache = false
    harness := NewHarness(config)
    harness.AddBenchmarks(GetMicrobenchmarks())
    results := harness.RunAll()
    
    // 3. Map simulator benchmarks to baselines
    mapping := map[string]string{
        "arithmetic_sequential": "arithmetic",
        "dependency_chain":      "dependency",
        "branch_taken":          "branch",
    }
    
    // 4. Calculate and report errors
    for simName, baselineName := range mapping {
        simResult := findResult(results, simName)
        baselineData := findBaseline(baseline, baselineName)
        
        error := math.Abs(simResult.CPI - baselineData.CPI) / 
                 math.Min(simResult.CPI, baselineData.CPI)
        
        t.Logf("%s: sim=%.3f, real=%.3f, error=%.1f%%",
            simName, simResult.CPI, baselineData.CPI, error*100)
        
        if error > 0.02 {
            t.Errorf("%s exceeds 2%% error threshold", simName)
        }
    }
}
```

### 3. CI Integration

Add to GitHub Actions workflow:

```yaml
- name: Run accuracy validation
  run: go test ./benchmarks/... -run TestAccuracy -v
```

### 4. Accuracy Report Generation

Create command to generate detailed accuracy report:

```bash
go run ./cmd/benchmark -format=json -core > sim_results.json
go run ./cmd/accuracy-report baseline=native/m2_baseline.json sim=sim_results.json
```

Output: `accuracy_report.md` with:
- Per-benchmark comparison table
- Error percentages
- Pass/fail status for <2% target
- Recommendations for calibration

## Benchmark Mapping

| Simulator Benchmark | M2 Baseline | What It Tests |
|---------------------|-------------|---------------|
| `arithmetic_sequential` | `arithmetic` | ALU throughput (ILP) |
| `dependency_chain` | `dependency` | Pipeline forwarding |
| `branch_taken` | `branch` | Branch prediction |
| `memory_sequential` | (needed) | Memory latency |
| `function_calls` | (needed) | Call/return overhead |
| `matrix_operations` | (needed) | Load/compute/store |
| `loop_simulation` | (needed) | Loop pattern |
| `mixed_operations` | (needed) | Realistic workload |

## Current Gap Analysis (Preliminary)

| Benchmark | Simulator CPI | Real M2 CPI | Error | Status |
|-----------|---------------|-------------|-------|--------|
| arithmetic | 1.200 | 0.268 | 348% | ❌ Need superscalar model |
| dependency | 2.200 | 1.009 | 118% | ❌ Forwarding needs work |
| branch | 2.900 | 1.190 | 144% | ❌ Branch prediction calibration |

## Recommendations

1. **Priority 1:** Fix forwarding path - dependency chain should be ~1 CPI
2. **Priority 2:** Add superscalar issue model - arithmetic should be <1 CPI
3. **Priority 3:** Calibrate branch predictor - branch should be ~1.2 CPI
4. **Future:** Collect more baseline data for memory, function calls, etc.

## Implementation Plan

1. **Phase 1:** Create `accuracy_test.go` with basic comparison
2. **Phase 2:** Add `accuracy-report` command for detailed analysis
3. **Phase 3:** Integrate into CI as regression check
4. **Phase 4:** Expand baselines for full benchmark coverage

## Success Criteria

- [ ] Automated tests compare simulator vs baseline
- [ ] Tests report error percentage per benchmark
- [ ] CI fails if any benchmark exceeds 2% error
- [ ] Accuracy report documents gaps and recommendations

## Notes

- Current simulator is significantly slower than real M2 (higher CPI)
- This is expected for initial implementation
- Goal is to identify gaps and guide timing model improvements
- <2% accuracy is ambitious - may need iterative calibration
