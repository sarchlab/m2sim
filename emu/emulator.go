// Package emu provides functional ARM64 emulation.
package emu

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sarchlab/m2sim/insts"
)

// Errors returned by the emulator.
var (
	// ErrProgramNotLoaded is returned when trying to execute without loading a program.
	ErrProgramNotLoaded = errors.New("program not loaded")

	// ErrMaxInstructions is returned when the instruction limit is reached.
	ErrMaxInstructions = errors.New("maximum instruction count reached")
)

// StepResult represents the result of executing a single instruction.
type StepResult struct {
	// Exited is true if the program terminated (via exit syscall).
	Exited bool

	// ExitCode is the exit status if Exited is true.
	ExitCode int64

	// InstructionCount is the total number of instructions executed so far.
	InstructionCount uint64
}

// SyscallHandlerFactory creates a SyscallHandler with the given RegFile, Memory, and I/O writers.
type SyscallHandlerFactory func(regFile *RegFile, memory *Memory, stdout, stderr io.Writer) SyscallHandler

// Emulator executes ARM64 instructions functionally.
type Emulator struct {
	regFile        *RegFile
	memory         *Memory
	decoder        *insts.Decoder
	syscallHandler SyscallHandler

	// Execution units
	alu        *ALU
	lsu        *LoadStoreUnit
	branchUnit *BranchUnit

	// I/O
	stdout io.Writer
	stderr io.Writer

	// Factory for creating syscall handler
	syscallFactory SyscallHandlerFactory

	// Execution state
	programLoaded    bool
	instructionCount uint64
	maxInstructions  uint64 // 0 means no limit
}

// EmulatorOption is a functional option for configuring the Emulator.
type EmulatorOption func(*Emulator)

// WithStdout sets a custom stdout writer.
func WithStdout(w io.Writer) EmulatorOption {
	return func(e *Emulator) {
		e.stdout = w
	}
}

// WithStderr sets a custom stderr writer.
func WithStderr(w io.Writer) EmulatorOption {
	return func(e *Emulator) {
		e.stderr = w
	}
}

// WithStackPointer sets the initial stack pointer value.
func WithStackPointer(sp uint64) EmulatorOption {
	return func(e *Emulator) {
		e.regFile.SP = sp
	}
}

// WithSyscallHandler sets a custom syscall handler.
func WithSyscallHandler(handler SyscallHandler) EmulatorOption {
	return func(e *Emulator) {
		e.syscallHandler = handler
	}
}

// WithSyscallHandlerFactory sets a factory for creating the syscall handler.
func WithSyscallHandlerFactory(factory SyscallHandlerFactory) EmulatorOption {
	return func(e *Emulator) {
		e.syscallFactory = factory
	}
}

// WithMaxInstructions sets the maximum number of instructions to execute.
// A value of 0 means no limit.
func WithMaxInstructions(max uint64) EmulatorOption {
	return func(e *Emulator) {
		e.maxInstructions = max
	}
}

// NewEmulator creates a new ARM64 emulator.
func NewEmulator(opts ...EmulatorOption) *Emulator {
	regFile := &RegFile{}
	memory := NewMemory()

	e := &Emulator{
		regFile:          regFile,
		memory:           memory,
		decoder:          insts.NewDecoder(),
		stdout:           os.Stdout,
		stderr:           os.Stderr,
		programLoaded:    false,
		instructionCount: 0,
		maxInstructions:  0,
	}

	// Apply options
	for _, opt := range opts {
		opt(e)
	}

	// Create execution units
	e.alu = NewALU(regFile)
	e.lsu = NewLoadStoreUnit(regFile, memory)
	e.branchUnit = NewBranchUnit(regFile)

	// Create syscall handler if factory provided and handler not set
	if e.syscallHandler == nil && e.syscallFactory != nil {
		e.syscallHandler = e.syscallFactory(regFile, memory, e.stdout, e.stderr)
	}

	// If still no handler, create default handler
	if e.syscallHandler == nil {
		e.syscallHandler = NewDefaultSyscallHandler(regFile, memory, e.stdout, e.stderr)
	}

	return e
}

// RegFile returns the emulator's register file.
func (e *Emulator) RegFile() *RegFile {
	return e.regFile
}

// Memory returns the emulator's memory.
func (e *Emulator) Memory() *Memory {
	return e.memory
}

// SetSyscallHandler sets the syscall handler.
func (e *Emulator) SetSyscallHandler(handler SyscallHandler) {
	e.syscallHandler = handler
}

// InstructionCount returns the total number of instructions executed.
func (e *Emulator) InstructionCount() uint64 {
	return e.instructionCount
}

// LoadProgram loads a program into memory and sets the entry point.
// The program can be provided as either a []byte slice or a *Memory.
// If a *Memory is provided, it replaces the emulator's memory.
func (e *Emulator) LoadProgram(entry uint64, program interface{}) {
	switch p := program.(type) {
	case []byte:
		e.memory.LoadProgram(entry, p)
	case *Memory:
		// Use the provided memory directly
		e.memory = p
		// Update execution units to use new memory
		e.lsu = NewLoadStoreUnit(e.regFile, e.memory)
	}

	e.regFile.PC = entry
	e.programLoaded = true
}

// Reset resets the emulator state.
func (e *Emulator) Reset() {
	e.regFile = &RegFile{}
	e.memory = NewMemory()
	e.alu = NewALU(e.regFile)
	e.lsu = NewLoadStoreUnit(e.regFile, e.memory)
	e.branchUnit = NewBranchUnit(e.regFile)
	e.programLoaded = false
	e.instructionCount = 0

	// Recreate syscall handler with new regFile and memory
	if e.syscallFactory != nil {
		e.syscallHandler = e.syscallFactory(e.regFile, e.memory, e.stdout, e.stderr)
	} else if e.syscallHandler != nil {
		// If we had a handler but no factory, create a new default handler
		e.syscallHandler = NewDefaultSyscallHandler(e.regFile, e.memory, e.stdout, e.stderr)
	}
}

// Step executes a single instruction.
// Returns a StepResult indicating whether execution should continue,
// and an error if one occurred.
func (e *Emulator) Step() (StepResult, error) {
	if !e.programLoaded {
		return StepResult{}, ErrProgramNotLoaded
	}

	// Check instruction limit
	if e.maxInstructions > 0 && e.instructionCount >= e.maxInstructions {
		return StepResult{InstructionCount: e.instructionCount}, ErrMaxInstructions
	}

	// 1. Fetch: Read 4 bytes at PC
	word := e.memory.Read32(e.regFile.PC)

	// 2. Decode
	inst := e.decoder.Decode(word)

	// 3. Execute
	result := e.execute(inst)

	// Increment instruction count
	e.instructionCount++
	result.InstructionCount = e.instructionCount

	if result.Exited {
		return result, nil
	}

	// Check for execution errors
	if inst.Op == insts.OpUnknown {
		return result, fmt.Errorf("unknown instruction 0x%08X at PC=0x%X", word, e.regFile.PC-4)
	}

	return result, nil
}

// Run executes instructions until the program exits or an error occurs.
// Returns the final result and any error.
func (e *Emulator) Run() (StepResult, error) {
	if !e.programLoaded {
		return StepResult{}, ErrProgramNotLoaded
	}

	for {
		result, err := e.Step()
		if err != nil {
			return result, err
		}
		if result.Exited {
			return result, nil
		}
	}
}

// execute dispatches and executes a decoded instruction.
func (e *Emulator) execute(inst *insts.Instruction) StepResult {
	// Check for unknown instruction
	if inst.Op == insts.OpUnknown {
		// PC will be advanced, but error will be returned by Step()
		e.regFile.PC += 4
		return StepResult{}
	}

	// Handle SVC (syscall) separately
	if inst.Op == insts.OpSVC {
		return e.executeSVC()
	}

	// Execute based on instruction type
	switch inst.Format {
	case insts.FormatDPImm:
		e.executeDPImm(inst)
	case insts.FormatDPReg:
		e.executeDPReg(inst)
	case insts.FormatBranch:
		e.executeBranch(inst)
		return StepResult{} // PC already updated by branch
	case insts.FormatBranchCond:
		e.executeBranchCond(inst)
		return StepResult{} // PC already updated
	case insts.FormatBranchReg:
		e.executeBranchReg(inst)
		return StepResult{} // PC already updated
	case insts.FormatLoadStore:
		e.executeLoadStore(inst)
	default:
		// Unimplemented format - advance PC anyway
		e.regFile.PC += 4
		return StepResult{}
	}

	// Advance PC by 4 (for non-branch instructions)
	e.regFile.PC += 4

	return StepResult{}
}

// executeSVC handles the SVC (supervisor call) instruction.
func (e *Emulator) executeSVC() StepResult {
	// Advance PC first (syscall return address is next instruction)
	e.regFile.PC += 4

	// Invoke syscall handler
	syscallResult := e.syscallHandler.Handle()

	return StepResult{
		Exited:   syscallResult.Exited,
		ExitCode: syscallResult.ExitCode,
	}
}

// executeDPImm executes Data Processing Immediate instructions.
func (e *Emulator) executeDPImm(inst *insts.Instruction) {
	imm := inst.Imm
	if inst.Shift > 0 {
		imm <<= inst.Shift
	}

	switch inst.Op {
	case insts.OpADD:
		if inst.Is64Bit {
			e.alu.ADD64Imm(inst.Rd, inst.Rn, imm, inst.SetFlags)
		} else {
			e.alu.ADD32Imm(inst.Rd, inst.Rn, uint32(imm), inst.SetFlags)
		}
	case insts.OpSUB:
		if inst.Is64Bit {
			e.alu.SUB64Imm(inst.Rd, inst.Rn, imm, inst.SetFlags)
		} else {
			e.alu.SUB32Imm(inst.Rd, inst.Rn, uint32(imm), inst.SetFlags)
		}
	}
}

// executeDPReg executes Data Processing Register instructions.
func (e *Emulator) executeDPReg(inst *insts.Instruction) {
	switch inst.Op {
	case insts.OpADD:
		if inst.Is64Bit {
			e.alu.ADD64(inst.Rd, inst.Rn, inst.Rm, inst.SetFlags)
		} else {
			e.alu.ADD32(inst.Rd, inst.Rn, inst.Rm, inst.SetFlags)
		}
	case insts.OpSUB:
		if inst.Is64Bit {
			e.alu.SUB64(inst.Rd, inst.Rn, inst.Rm, inst.SetFlags)
		} else {
			e.alu.SUB32(inst.Rd, inst.Rn, inst.Rm, inst.SetFlags)
		}
	case insts.OpAND:
		if inst.Is64Bit {
			e.alu.AND64(inst.Rd, inst.Rn, inst.Rm, inst.SetFlags)
		} else {
			e.alu.AND32(inst.Rd, inst.Rn, inst.Rm, inst.SetFlags)
		}
	case insts.OpORR:
		if inst.Is64Bit {
			e.alu.ORR64(inst.Rd, inst.Rn, inst.Rm)
		} else {
			e.alu.ORR32(inst.Rd, inst.Rn, inst.Rm)
		}
	case insts.OpEOR:
		if inst.Is64Bit {
			e.alu.EOR64(inst.Rd, inst.Rn, inst.Rm)
		} else {
			e.alu.EOR32(inst.Rd, inst.Rn, inst.Rm)
		}
	}
}

// executeBranch executes unconditional branch instructions (B, BL).
func (e *Emulator) executeBranch(inst *insts.Instruction) {
	switch inst.Op {
	case insts.OpB:
		e.branchUnit.B(inst.BranchOffset)
	case insts.OpBL:
		e.branchUnit.BL(inst.BranchOffset)
	}
}

// executeBranchCond executes conditional branch instructions.
func (e *Emulator) executeBranchCond(inst *insts.Instruction) {
	// Convert insts.Cond to emu.Cond
	cond := Cond(inst.Cond)

	if e.branchUnit.CheckCondition(cond) {
		e.regFile.PC = uint64(int64(e.regFile.PC) + inst.BranchOffset)
	} else {
		// Condition not met, advance to next instruction
		e.regFile.PC += 4
	}
}

// executeBranchReg executes branch to register instructions (BR, BLR, RET).
func (e *Emulator) executeBranchReg(inst *insts.Instruction) {
	switch inst.Op {
	case insts.OpBR:
		e.branchUnit.BR(inst.Rn)
	case insts.OpBLR:
		e.branchUnit.BLR(inst.Rn)
	case insts.OpRET:
		e.branchUnit.RET(inst.Rn)
	}
}

// executeLoadStore executes load and store instructions.
func (e *Emulator) executeLoadStore(inst *insts.Instruction) {
	// Check if base register is SP (register 31 in load/store context means SP)
	useSP := inst.Rn == 31

	switch inst.Op {
	case insts.OpLDR:
		if inst.Is64Bit {
			if useSP {
				e.lsu.LDR64SP(inst.Rd, inst.Imm)
			} else {
				e.lsu.LDR64(inst.Rd, inst.Rn, inst.Imm)
			}
		} else {
			if useSP {
				e.lsu.LDR32SP(inst.Rd, inst.Imm)
			} else {
				e.lsu.LDR32(inst.Rd, inst.Rn, inst.Imm)
			}
		}
	case insts.OpSTR:
		if inst.Is64Bit {
			if useSP {
				e.lsu.STR64SP(inst.Rd, inst.Imm)
			} else {
				e.lsu.STR64(inst.Rd, inst.Rn, inst.Imm)
			}
		} else {
			if useSP {
				e.lsu.STR32SP(inst.Rd, inst.Imm)
			} else {
				e.lsu.STR32(inst.Rd, inst.Rn, inst.Imm)
			}
		}
	}
}
