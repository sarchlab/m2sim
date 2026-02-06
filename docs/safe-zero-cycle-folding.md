# Safe Zero-Cycle Branch Folding: Reimplementation Approach

**Author:** Eric (Researcher)  
**Date:** 2026-02-05 (Cycle 265)  
**Purpose:** Research safe reimplementation of zero-cycle branch folding with misprediction recovery

## Executive Summary

Zero-cycle branch folding was disabled (commit 1590518) because it caused infinite loops — conditional branches were eliminated at fetch time without verification. To safely reimplement this optimization, branches must **still enter the pipeline** for verification, but can "retire" in zero effective cycles when prediction is correct.

## The Core Problem

The unsafe implementation eliminated branches at fetch time:

```go
// UNSAFE — eliminated in commit 1590518
if isCond, _ := isFoldableConditionalBranch(word, fetchPC); isCond {
    pred := p.branchPredictor.Predict(fetchPC)
    if pred.TargetKnown && pred.Taken && pred.Confidence >= 3 {
        fetchPC = pred.Target    // ← Redirect fetch
        p.stats.FoldedBranches++
        continue                  // ← Branch NEVER enters pipeline!
    }
}
```

**Why this failed:**
1. Branch never entered pipeline (IFID/IDEX/EXMEM stages)
2. Condition flags were never checked
3. When prediction became wrong (loop exit), no recovery mechanism existed
4. Result: infinite loops

## Safe Reimplementation Approach

### Core Principle: Speculative Execution with Verification

The branch **must enter the pipeline** for verification, but we can:
1. **Speculatively redirect fetch** before branch executes
2. **Verify prediction** when branch reaches execute stage
3. **Recover on misprediction** via flush + PC redirect

### Implementation Strategy

#### Phase 1: Track Folded Branches

Add tracking state to Pipeline struct:

```go
type FoldedBranchRecord struct {
    PC             uint64
    PredictedTaken bool
    PredictedTarget uint64
    SlotIndex      int
    Verified       bool
}

type Pipeline struct {
    // ... existing fields ...
    foldedBranches []FoldedBranchRecord  // Active folded branches awaiting verification
}
```

#### Phase 2: Speculative Fetch Redirection

In `collectPendingFetchInstructions8()`:

```go
// Check if instruction is a conditional branch
if isCond, target := isFoldableConditionalBranch(word, fetchPC); isCond {
    pred := p.branchPredictor.Predict(fetchPC)
    
    if pred.TargetKnown && pred.Taken && pred.Confidence >= 3 {
        // Record this branch for later verification
        p.foldedBranches = append(p.foldedBranches, FoldedBranchRecord{
            PC:              fetchPC,
            PredictedTaken:  true,
            PredictedTarget: pred.Target,
            SlotIndex:       slot,
            Verified:        false,
        })
        
        // CRITICAL: Still add branch to fetch queue (enters pipeline)
        p.fetchQueue = append(p.fetchQueue, FetchEntry{PC: fetchPC, ...})
        
        // Speculatively redirect fetch to predicted target
        nextFetchPC = pred.Target
        p.stats.FoldedBranches++
        
        // Stop fetching sequential instructions after branch
        break
    }
}
```

**Key difference from unsafe version:** The branch **still enters the pipeline**.

#### Phase 3: Execute Stage Verification

In `tickOctupleIssue()`, when a branch reaches execute:

```go
// After branch execution computes actualTaken and actualTarget
for i, fb := range p.foldedBranches {
    if fb.PC == idex.PC && !fb.Verified {
        // Verify prediction
        if actualTaken != fb.PredictedTaken || actualTarget != fb.PredictedTarget {
            // MISPREDICTION! Full pipeline flush
            p.flushPipeline()
            p.nextPC = actualTarget
            p.stats.BranchMispredictions++
            // Clear all pending folded branches
            p.foldedBranches = nil
            return
        }
        
        // Prediction correct — mark verified, no penalty
        p.foldedBranches[i].Verified = true
        break
    }
}

// Clean up verified folded branches
p.foldedBranches = filterUnverified(p.foldedBranches)
```

#### Phase 4: Zero-Cycle Retirement

The "zero-cycle" part comes from how we handle correctly predicted branches:

```go
// In execute stage timing
if idex.IsBranch && p.isFoldedBranch(idex.PC) {
    // Branch was folded — if prediction correct, it "retires" immediately
    // No additional cycle penalty for branch resolution
    // Fetch was already redirected in fetch stage
    
    if predictionCorrect {
        // Branch effectively completed in fetch stage
        // Just verify and continue — no additional cycles
    }
}
```

### Misprediction Recovery Requirements

When a folded branch mispredicts:

1. **Flush all pipeline stages** (IFID, IDEX, EXMEM, MEMWB)
2. **Clear all pending folded branches** (they fetched wrong instructions)
3. **Redirect PC** to correct target (or fall-through for not-taken)
4. **Update branch predictor** with correct outcome
5. **Increment misprediction penalty** counter

```go
func (p *Pipeline) recoverFromMisprediction(correctPC uint64) {
    // 1. Flush pipeline
    p.ifid.Clear()
    p.idex.Clear()
    p.idex2.Clear()
    // ... clear all IDEX slots (2-8) ...
    p.exmem.Clear()
    // ... clear all EXMEM slots (2-8) ...
    p.memwb.Clear()
    
    // 2. Clear pending folded branches
    p.foldedBranches = nil
    
    // 3. Redirect PC
    p.nextPC = correctPC
    p.nextFetchPC = correctPC
    
    // 4. Mark pipeline as recovering (optional, for debugging)
    p.recovering = true
    
    // 5. Stats
    p.stats.PipelineFlushes++
}
```

### Edge Cases

| Scenario | Handling |
|----------|----------|
| Back-to-back folded branches | Only fold first; subsequent branches wait for verify |
| Folded branch in slot 0 vs slot 7 | Same logic, different IDEX register |
| Misprediction clears multiple folded | Clear all pending — all were speculative |
| Unconditional B (not B.cond) | Continue to fold at fetch — no verification needed |

## Impact Estimation

| Metric | Before (disabled) | After (safe reimpl) |
|--------|-------------------|---------------------|
| FoldedBranches | 0 | >0 for correctly predicted |
| Branch CPI | 1.600 | ~1.1-1.2 (estimated) |
| Branch error | 34.5% | ~15-20% (estimated) |
| Average error | 20.2% | ~15-17% (estimated) |

## Implementation Effort

| Phase | Estimated LOC | Risk |
|-------|---------------|------|
| 1. Tracking state | ~20 | Low |
| 2. Speculative fetch | ~50 | Medium |
| 3. Execute verification | ~60 | Medium |
| 4. Recovery handling | ~40 | Medium |
| **Total** | **~170** | Medium |

## Testing Strategy

1. **Unit tests:** Test folded branch tracking, verification, recovery
2. **Integration tests:** Run `branch_hot_loop`, verify correct completion
3. **Accuracy tests:** Measure branch CPI, compare to baseline
4. **Edge case tests:** Back-to-back branches, mixed branch types

## Recommendation

Proceed with this safe reimplementation approach. The key insight is that "zero-cycle" doesn't mean "skip execution" — it means "speculatively redirect fetch, then verify prediction and retire with no additional penalty if correct."

The M2 likely uses a similar technique: aggressive speculation with fast recovery on misprediction.

## Related Documents

- `docs/zero-cycle-folding-bug.md` — Original bug analysis
- `docs/zero-cycle-branch-implementation.md` — Earlier implementation guide
- `docs/zero-cycle-branch-research.md` — Background research

## Files to Modify

| File | Changes |
|------|---------|
| `timing/pipeline/pipeline.go` | Add FoldedBranchRecord tracking |
| `timing/pipeline/superscalar.go` | Speculative fetch in tickOctupleIssue() |
| `timing/pipeline/stages.go` | Verification in execute stage |
