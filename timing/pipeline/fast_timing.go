package pipeline

import (
	"github.com/sarchlab/m2sim/emu"
	"github.com/sarchlab/m2sim/insts"
	"github.com/sarchlab/m2sim/timing/latency"
)

// FastTiming provides a simplified timing simulation optimized for calibration.
// It eliminates detailed pipeline simulation overhead while preserving basic
// timing relationships between instructions.
type FastTiming struct {
	regFile        *emu.RegFile
	memory         *emu.Memory
	decoder        *insts.Decoder
	latencyTable   *latency.Table
	syscallHandler emu.SyscallHandler

	// Simplified state
	PC              uint64
	halted          bool
	exitCode        int64
	cycleCount      uint64
	instrCount      uint64
	maxInstructions uint64 // 0 means no limit

	// Simple latency tracking - instructions that complete in future cycles
	pendingOps []DelayedOp
}

// DelayedOp represents an instruction that completes after a delay.
type DelayedOp struct {
	completeCycle uint64
	writeReg      uint8
	writeValue    uint64
}

// FastTimingOption configures fast timing simulation.
type FastTimingOption func(*FastTiming)

// WithMaxInstructions sets the maximum number of instructions to execute.
// A value of 0 means no limit.
func WithMaxInstructions(max uint64) FastTimingOption {
	return func(ft *FastTiming) {
		ft.maxInstructions = max
	}
}

// NewFastTiming creates a new fast timing simulation.
func NewFastTiming(regFile *emu.RegFile, memory *emu.Memory, latencyTable *latency.Table, syscallHandler emu.SyscallHandler, opts ...FastTimingOption) *FastTiming {
	ft := &FastTiming{
		regFile:         regFile,
		memory:          memory,
		decoder:         insts.NewDecoder(),
		latencyTable:    latencyTable,
		syscallHandler:  syscallHandler,
		pendingOps:      make([]DelayedOp, 0, 8), // Pre-allocate small buffer
		maxInstructions: 0,                       // Default: no limit
	}

	// Apply options
	for _, opt := range opts {
		opt(ft)
	}

	return ft
}

// SetPC sets the program counter.
func (ft *FastTiming) SetPC(pc uint64) {
	ft.PC = pc
}

// Run executes the fast timing simulation until halt.
func (ft *FastTiming) Run() int64 {
	for !ft.halted {
		ft.Tick()
	}
	return ft.exitCode
}

// Tick executes one fast timing cycle.
func (ft *FastTiming) Tick() {
	if ft.halted {
		return
	}

	// Check instruction limit before executing
	if ft.maxInstructions > 0 && ft.instrCount >= ft.maxInstructions {
		ft.halted = true
		ft.exitCode = 0 // Exit normally when instruction limit reached
		return
	}

	ft.cycleCount++

	// Complete any pending operations
	ft.completePendingOps()

	// Fetch and execute instruction
	word := ft.memory.Read32(ft.PC)
	inst := ft.decoder.Decode(word)

	if inst == nil || inst.Op == insts.OpUnknown {
		// Unknown instruction - halt
		ft.halted = true
		ft.exitCode = -1
		return
	}

	// Execute instruction with simplified timing
	ft.executeInstruction(inst, ft.PC)
	ft.instrCount++
}

// completePendingOps completes operations that finish in this cycle.
func (ft *FastTiming) completePendingOps() {
	// Process operations in reverse order to avoid index shifting
	for i := len(ft.pendingOps) - 1; i >= 0; i-- {
		op := ft.pendingOps[i]
		if op.completeCycle <= ft.cycleCount {
			// Complete the operation
			if op.writeReg != 31 { // Not XZR
				ft.regFile.WriteReg(op.writeReg, op.writeValue)
			}

			// Remove completed operation
			ft.pendingOps[i] = ft.pendingOps[len(ft.pendingOps)-1]
			ft.pendingOps = ft.pendingOps[:len(ft.pendingOps)-1]
		}
	}
}

// executeInstruction executes an instruction with simplified timing.
func (ft *FastTiming) executeInstruction(inst *insts.Instruction, pc uint64) {
	writeReg := uint8(31) // XZR (no write)
	var writeValue uint64
	var instLatency uint64
	needsDelay := false

	// Read operands
	rnValue := ft.regFile.ReadReg(inst.Rn)
	rmValue := ft.regFile.ReadReg(inst.Rm)

	switch inst.Op {
	case insts.OpADD:
		instLatency = ft.latencyTable.GetLatency(inst)
		writeReg = inst.Rd
		writeValue = ft.executeADD(inst, rnValue, rmValue)
		needsDelay = (instLatency > 1)

	case insts.OpSUB:
		instLatency = ft.latencyTable.GetLatency(inst)
		writeReg = inst.Rd
		writeValue = ft.executeSUB(inst, rnValue, rmValue)
		needsDelay = (instLatency > 1)

	case insts.OpLDR:
		instLatency = ft.latencyTable.GetLatency(inst)
		writeReg = inst.Rd
		var addr uint64

		// Handle different addressing modes
		switch inst.IndexMode {
		case insts.IndexPost:
			// Post-index: [Rn], #imm
			addr = rnValue
			writeValue = ft.memory.Read64(addr)
			// Update base register after load
			newAddr := rnValue + uint64(inst.SignedImm)
			ft.regFile.WriteReg(inst.Rn, newAddr)
		case insts.IndexPre:
			// Pre-index: [Rn, #imm]!
			addr = rnValue + uint64(inst.SignedImm)
			writeValue = ft.memory.Read64(addr)
			// Update base register before load
			ft.regFile.WriteReg(inst.Rn, addr)
		default:
			// Normal addressing: [Rn, #imm] or [Rn + Rm]
			addr = rnValue + uint64(int64(inst.Imm))
			writeValue = ft.memory.Read64(addr)
		}
		needsDelay = true // Always delay loads

	case insts.OpSTR:
		addr := rnValue + uint64(int64(inst.Imm))
		storeValue := ft.regFile.ReadReg(inst.Rd)
		ft.memory.Write64(addr, storeValue)
		// Stores don't write registers

	case insts.OpB:
		ft.handleBranch(inst, pc)
		return // Branch handled separately

	case insts.OpBCond:
		ft.handleConditionalBranch(inst, pc)
		return

	case insts.OpSVC:
		ft.handleSyscall()
		return

	case insts.OpADRP:
		writeReg = inst.Rd
		// ADRP: (PC & ~0xFFF) + (SignExtend(imm) << 12)
		pcPage := pc &^ 0xFFF
		pageOffset := int64(inst.Imm) << 12
		writeValue = uint64(int64(pcPage) + pageOffset)

	case insts.OpMOVZ:
		writeReg = inst.Rd
		// MOVZ: Move wide with zero (clear other bits)
		shift := uint64(inst.Shift * 16) // Shift is in units of 16 bits
		writeValue = inst.Imm << shift

	case insts.OpSTP:
		// Store pair: [base + offset] = reg1, [base + offset + 8] = reg2
		addr := rnValue + uint64(inst.SignedImm)
		value1 := ft.regFile.ReadReg(inst.Rd)
		value2 := ft.regFile.ReadReg(inst.Rt2)
		ft.memory.Write64(addr, value1)
		ft.memory.Write64(addr+8, value2)
		// Update base register if pre/post-index
		if inst.IndexMode != insts.IndexNone && inst.IndexMode != insts.IndexSigned {
			ft.regFile.WriteReg(inst.Rn, addr)
		}

	case insts.OpLDP:
		// Load pair: reg1 = [base + offset], reg2 = [base + offset + 8]
		addr := rnValue + uint64(inst.SignedImm)
		value1 := ft.memory.Read64(addr)
		value2 := ft.memory.Read64(addr + 8)
		// Write both registers
		if inst.Rd != 31 {
			ft.regFile.WriteReg(inst.Rd, value1)
		}
		if inst.Rt2 != 31 {
			ft.regFile.WriteReg(inst.Rt2, value2)
		}
		// Update base register if pre/post-index
		if inst.IndexMode != insts.IndexNone && inst.IndexMode != insts.IndexSigned {
			ft.regFile.WriteReg(inst.Rn, addr)
		}

	case insts.OpBL:
		// Branch with link: LR = PC + 4, PC = PC + offset
		ft.regFile.WriteReg(30, pc+4)  // X30 = LR
		offset := int64(inst.Imm) << 2 // Word-aligned offset
		ft.PC = uint64(int64(pc) + offset)
		return // Don't advance PC at end

	case insts.OpAND:
		instLatency = ft.latencyTable.GetLatency(inst)
		writeReg = inst.Rd
		switch inst.Format {
		case insts.FormatDPImm, insts.FormatLogicalImm:
			writeValue = rnValue & inst.Imm
		case insts.FormatDPReg:
			writeValue = rnValue & rmValue
		default:
			writeValue = rnValue & inst.Imm
		}
		needsDelay = (instLatency > 1)

	case insts.OpORR:
		instLatency = ft.latencyTable.GetLatency(inst)
		writeReg = inst.Rd
		switch inst.Format {
		case insts.FormatDPImm, insts.FormatLogicalImm:
			writeValue = rnValue | inst.Imm
		case insts.FormatDPReg:
			writeValue = rnValue | rmValue
		default:
			writeValue = rnValue | inst.Imm
		}
		needsDelay = (instLatency > 1)

	case insts.OpEOR:
		instLatency = ft.latencyTable.GetLatency(inst)
		writeReg = inst.Rd
		switch inst.Format {
		case insts.FormatDPImm, insts.FormatLogicalImm:
			writeValue = rnValue ^ inst.Imm
		case insts.FormatDPReg:
			writeValue = rnValue ^ rmValue
		default:
			writeValue = rnValue ^ inst.Imm
		}
		needsDelay = (instLatency > 1)

	case insts.OpRET:
		// RET: Return - jump to address in X30 (LR)
		targetAddr := ft.regFile.ReadReg(30) // X30 = LR
		ft.PC = targetAddr
		return // Don't advance PC at end

	default:
		// Unknown instruction - treat as 1-cycle NOP
	}

	// Handle instruction completion
	if needsDelay && writeReg != 31 {
		// Delay the register write
		ft.pendingOps = append(ft.pendingOps, DelayedOp{
			completeCycle: ft.cycleCount + instLatency,
			writeReg:      writeReg,
			writeValue:    writeValue,
		})
	} else if writeReg != 31 {
		// Immediate write
		ft.regFile.WriteReg(writeReg, writeValue)
	}

	// Advance PC
	ft.PC += 4
}

// executeADD performs ADD instruction calculation.
func (ft *FastTiming) executeADD(inst *insts.Instruction, rnValue, rmValue uint64) uint64 {
	switch inst.Format {
	case insts.FormatDPImm:
		return rnValue + uint64(int64(inst.Imm))
	case insts.FormatDPReg:
		return rnValue + rmValue
	default:
		return rnValue + uint64(int64(inst.Imm))
	}
}

// executeSUB performs SUB instruction calculation.
func (ft *FastTiming) executeSUB(inst *insts.Instruction, rnValue, rmValue uint64) uint64 {
	switch inst.Format {
	case insts.FormatDPImm:
		return rnValue - uint64(int64(inst.Imm))
	case insts.FormatDPReg:
		return rnValue - rmValue
	default:
		return rnValue - uint64(int64(inst.Imm))
	}
}

// handleBranch processes unconditional branch instructions.
func (ft *FastTiming) handleBranch(inst *insts.Instruction, pc uint64) {
	// Calculate target - immediate is already word-aligned offset
	offset := int64(inst.Imm) // Already shifted by decoder
	target := uint64(int64(pc) + offset)
	ft.PC = target
}

// handleConditionalBranch processes conditional branch instructions.
func (ft *FastTiming) handleConditionalBranch(inst *insts.Instruction, pc uint64) {
	// Simple condition evaluation (simplified - no detailed flag tracking)
	taken := ft.evaluateCondition(uint8(inst.Cond))

	if taken {
		offset := int64(inst.Imm) << 2
		ft.PC = uint64(int64(pc) + offset)
	} else {
		ft.PC += 4
	}
}

// evaluateCondition evaluates a condition code (simplified).
func (ft *FastTiming) evaluateCondition(cond uint8) bool {
	// For fast timing, use actual PSTATE flags
	pstate := &ft.regFile.PSTATE

	switch cond & 0xE {
	case 0x0: // EQ/NE
		return pstate.Z == (cond&1 == 0)
	case 0x2: // CS/CC
		return pstate.C == (cond&1 == 0)
	case 0x4: // MI/PL
		return pstate.N == (cond&1 == 0)
	case 0x6: // VS/VC
		return pstate.V == (cond&1 == 0)
	case 0x8: // HI/LS
		if cond&1 == 0 {
			return pstate.C && !pstate.Z
		}
		return !pstate.C || pstate.Z
	case 0xA: // GE/LT
		if cond&1 == 0 {
			return pstate.N == pstate.V
		}
		return pstate.N != pstate.V
	case 0xC: // GT/LE
		if cond&1 == 0 {
			return !pstate.Z && (pstate.N == pstate.V)
		}
		return pstate.Z || (pstate.N != pstate.V)
	case 0xE: // AL (always)
		return true
	default:
		return false
	}
}

// handleSyscall processes system call instructions.
func (ft *FastTiming) handleSyscall() {
	if ft.syscallHandler != nil {
		// Use simplified syscall handling
		syscallNum := ft.regFile.ReadReg(8) // X8 contains syscall number

		// For exit syscall, halt
		if syscallNum == 93 { // SYS_EXIT
			ft.halted = true
			ft.exitCode = int64(ft.regFile.ReadReg(0)) // X0 contains exit code
			return
		}

		// Other syscalls - delegate to handler
		// Note: This is simplified - full syscall handling would require
		// more context about process state
	}

	ft.PC += 4
}

// Stats returns simulation statistics.
func (ft *FastTiming) Stats() Statistics {
	return Statistics{
		Cycles:       ft.cycleCount,
		Instructions: ft.instrCount,
		// Simplified stats - no detailed hazard tracking in fast mode
		Stalls:               0,
		Flushes:              0,
		ExecStalls:           0,
		MemStalls:            0,
		DataHazards:          0,
		BranchPredictions:    0,
		BranchCorrect:        0,
		BranchMispredictions: 0,
	}
}
