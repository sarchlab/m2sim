# Same-Cycle Forwarding Impact Analysis — February 8, 2026

## Critical Finding: Zero Accuracy Impact

**PR #370/381 (same-cycle forwarding) has been successfully merged but shows ZERO impact on timing accuracy.** This contradicts expectations and requires immediate investigation.

## Comparison Results

### Before Same-Cycle Forwarding (Baseline - Feb 8)
| Benchmark | M2 CPI | Full CPI | Error % |
|-----------|--------|----------|---------|
| Arithmetic| 0.296  | 0.400    | 35.2%   |
| Dependency| 1.088  | 1.200    | 10.3%   |
| Branch    | 1.304  | 1.600    | 22.7%   |

### After Same-Cycle Forwarding (Post-Merge - Feb 8)
| Benchmark | M2 CPI | Full CPI | Error % | Change |
|-----------|--------|----------|---------|--------|
| Arithmetic| 0.296  | 0.400    | 35.2%   | **0.0%** |
| Dependency| 1.088  | 1.200    | 10.3%   | **0.0%** |
| Branch    | 1.304  | 1.600    | 22.7%   | **0.0%** |

**Average Error: 22.8% (unchanged)**

## Technical Analysis

### What Was Implemented
Leo's PR correctly implemented same-cycle ALU→ALU forwarding by:
1. **Relaxing RAW hazard blocking** in `canDualIssue()` and `canIssueWith()`
2. **Allowing ALU→ALU co-issue** when producer is non-memory ALU instruction
3. **Preserving load→ALU blocking** since load data unavailable until MEM stage
4. **Blocking store value forwarding** which lacks same-cycle forwarding paths

### Code Changes Verification
```go
// NEW: Only block RAW if producer is a memory load
if hasRAW && prev.MemRead {
    return false  // Block load→ALU dependencies
}
// ALU→ALU dependencies now allowed to co-issue
```

## Root Cause Analysis: Why Zero Impact?

### Hypothesis 1: Benchmark Composition Mismatch
**The current calibration benchmarks may not contain the specific ALU→ALU dependency patterns that this fix targets.**

Our three core benchmarks:
- **Arithmetic**: `arithmetic_sequential` - May be pure independent ALU ops
- **Dependency**: `dependency_chain` - Likely load→ALU or control dependencies
- **Branch**: `branch_taken_conditional` - Branch prediction focused

### Hypothesis 2: Benchmark Microarchitecture
**The bottleneck may not be ALU→ALU co-issue blocking but other pipeline constraints:**
- Memory bandwidth limitations
- Branch prediction accuracy
- Load-store queue capacity
- Other hazard types (WAW, WAR)

### Hypothesis 3: Superscalar Modeling Gaps
**The fix may be correct but insufficient without other superscalar improvements:**
- Register renaming modeling
- Out-of-order execution depth
- Issue queue size limitations
- Retirement width constraints

## Immediate Action Plan

### 1. Benchmark Analysis (Priority 1)
**Investigate whether current benchmarks actually stress ALU→ALU forwarding:**
- Examine assembly output of `arithmetic_sequential` benchmark
- Check for ALU→ALU dependency chains in the critical path
- Validate that the forwarding scenarios exist in our test cases

### 2. Targeted Benchmark Development (Priority 2)
**Create a specific ALU→ALU forwarding stress test:**
```asm
# Example: ALU forwarding chain
add x1, x0, x0    # Independent
add x2, x1, x1    # Depends on x1 (should benefit from forwarding)
add x3, x2, x2    # Depends on x2 (should benefit from forwarding)
```

### 3. Pipeline Profiling (Priority 3)
**Instrument the pipeline to understand actual bottlenecks:**
- Track co-issue rates before/after the fix
- Identify which hazard types are actually blocking dual-issue
- Measure forwarding path utilization

## Strategic Implications

### For H3 Calibration
**This finding suggests our accuracy bottlenecks may be different than initially assumed:**
- **35.2% arithmetic error** may NOT be primarily ALU→ALU forwarding
- **Other microarchitectural modeling gaps** likely dominate
- **Need deeper pipeline analysis** to identify real bottlenecks

### For Issue #370 Resolution
**While the implementation is correct, issue #370 cannot be closed until we verify:**
- The fix addresses the intended scenario
- We have benchmarks that actually stress ALU→ALU forwarding
- Alternative bottlenecks are identified and prioritized

## Next Steps

1. **Immediate (This Cycle)**: Report findings on issue #370 with analysis plan
2. **Short-term**: Create targeted ALU forwarding benchmark and validate fix works
3. **Medium-term**: Expand pipeline instrumentation to identify real bottlenecks
4. **Long-term**: Develop comprehensive microarchitecture accuracy improvement plan

## Recommendation

**DO NOT close issue #370 yet.** While Leo's implementation is technically sound, the zero accuracy impact indicates either:
1. Our benchmarks don't stress the fixed scenario, OR
2. Other bottlenecks dominate the accuracy error

We need targeted validation before claiming this issue is resolved.

---
*Analysis by Alex (Data Analysis Specialist) - February 8, 2026*
*Based on CPI comparison workflow runs 21807316995 (post-fix) vs PR #376 baseline*