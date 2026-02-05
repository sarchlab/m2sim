# Zero-Cycle Predicted-Taken Branch Implementation Guide

**Created:** 2026-02-05 (Cycle 240)
**Author:** Eric (Research Agent)
**Purpose:** Detailed implementation guide for Bob to implement zero-cycle branches

## Executive Summary

This is the **highest-impact** optimization available to reduce the 34.5% branch error. M2 likely achieves low branch CPI by "folding" correctly predicted taken branches — executing them at zero effective cost. Our simulator currently costs 1+ cycles even for correctly predicted branches.

## Current Flow (Problem)

```
Cycle N:   Fetch(PC)      → gets branch instruction
Cycle N+1: Decode(branch) → identifies as branch
Cycle N+2: Execute(branch)→ evaluates condition, computes target
Cycle N+3: Fetch(target)  → finally fetches from target
```

**Cost:** 3+ cycles per branch, even when correctly predicted!

## Target Flow (Solution)

```
Cycle N:   Fetch(PC) → BTB hit + prediction "taken"
           → Immediately redirect fetch to target (no execute needed)
Cycle N+1: Fetch(target) → continues from target
```

**Cost:** 0 effective cycles for predicted-taken branches!

## Implementation Approach

### Option 1: Fetch-Stage Branch Resolution (Recommended)

Modify the fetch logic in `timing/pipeline/pipeline.go` to check BTB before decode:

```go
// In tickFetch or collectPendingFetchInstructions
func (p *Pipeline) fetchWithBranchPrediction(pc uint64) {
    // Check BTB before fetching
    pred := p.branchPredictor.Predict(pc)
    
    if pred.TargetKnown && pred.Taken {
        // BTB hit and prediction is "taken"
        // Mark this branch as "folded" — no execute stage needed
        p.foldedBranchPC = pc
        p.foldedBranchTarget = pred.Target
        
        // Redirect fetch to target immediately
        p.nextPC = pred.Target
        
        // Still need to fetch the branch instruction for later verification
        // but don't wait for execute stage to resolve direction
    }
}
```

### Option 2: Superscalar Integration

Modify `tickOctupleIssue` in `superscalar.go` to handle folded branches:

```go
// During fetch collection for 8-wide
func (p *Pipeline) collectPendingFetchInstructions8() {
    for slot := 0; slot < 8; slot++ {
        pc := p.nextFetchPC + uint64(slot*4)
        
        // Check for branch before adding to fetch queue
        pred := p.branchPredictor.Predict(pc)
        if pred.TargetKnown && pred.Taken {
            // This is a predicted-taken branch
            // Stop fetching sequential instructions after this one
            // Redirect next fetch to target
            p.nextFetchPC = pred.Target
            break // Don't fetch beyond the branch
        }
    }
}
```

### Option 3: Execute Stage Bypass (Simpler but Less Effective)

Add a "fast path" in execute stage for branches:

```go
// In Execute method
if idex.IsBranch && p.branchWasPredictedTaken(idex.PC) {
    // Skip full branch evaluation, trust prediction
    result.BranchTaken = true
    result.BranchTarget = p.getPredictedTarget(idex.PC)
    // Saves 1 cycle of execute latency
}
```

## Key Data Structures to Add

```go
// Add to Pipeline struct
type Pipeline struct {
    // ... existing fields ...
    
    // Zero-cycle branch tracking
    foldedBranches map[uint64]uint64 // PC -> predicted target
    branchInFlight bool              // Is a predicted branch being verified?
    
    // Recovery state
    pendingBranchVerify struct {
        PC       uint64
        Target   uint64
        Taken    bool
        Verified bool
    }
}
```

## Verification and Recovery

Even with zero-cycle branches, we must **verify** predictions and recover on mispredict:

```go
// When branch reaches execute stage, verify prediction
func (p *Pipeline) verifyBranchPrediction(pc uint64, actualTaken bool, actualTarget uint64) {
    predicted := p.foldedBranches[pc]
    
    if actualTaken != true || actualTarget != predicted {
        // MISPREDICT! Need to flush and recover
        p.flushPipeline()
        p.nextPC = actualTarget
        p.branchPredictor.Update(pc, actualTaken, actualTarget)
        // Incur misprediction penalty (5+ cycles)
    } else {
        // Prediction was correct, no penalty
        delete(p.foldedBranches, pc)
        p.branchPredictor.Update(pc, actualTaken, actualTarget)
    }
}
```

## Impact Estimation

| Scenario | Current Cost | After Zero-Cycle | Improvement |
|----------|--------------|------------------|-------------|
| Predicted-taken (correct) | 2-3 cycles | 0 cycles | ~15-20% CPI |
| BTB miss | 3+ cycles | 3+ cycles | No change |
| Misprediction | 5-8 cycles | 5-8 cycles | No change |

For `branchTakenConditional` benchmark:
- 5 iterations, ~100% prediction accuracy after warmup
- Current: CPI 1.600
- Expected after zero-cycle: CPI ~1.1-1.2 (near M2's 1.190)

**Estimated branch error reduction:** 34.5% → ~15-20%

## Testing Considerations

1. **Correctness test:** Run all benchmarks, verify same instruction count and results
2. **Accuracy test:** Branch CPI should drop significantly
3. **Edge cases:** 
   - BTB miss (first iteration)
   - Back-to-back branches
   - Indirect branches (BLR, BR)

## Files to Modify

| File | Changes |
|------|---------|
| `timing/pipeline/pipeline.go` | Add folded branch tracking, fetch-stage prediction |
| `timing/pipeline/superscalar.go` | Handle folded branches in 8-wide fetch |
| `timing/pipeline/stages.go` | Add prediction verification in execute |

## Recommended Implementation Order

1. Add `foldedBranches` map to Pipeline struct
2. Modify fetch to check BTB and redirect on predicted-taken
3. Add verification in execute stage
4. Handle misprediction recovery (flush)
5. Update statistics logging

## Summary

Implementing zero-cycle predicted-taken branches will have the highest impact on branch accuracy. The key insight is to resolve branch direction in the fetch stage (using BTB + predictor) rather than waiting for the execute stage. This matches how M2 likely achieves its low branch CPI.

**Estimated effort:** Medium (100-200 lines of code)
**Estimated impact:** 34.5% → ~15-20% branch error
