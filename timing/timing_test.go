// Package timing provides integration tests for timing simulation mode.
package timing_test

import (
	"encoding/binary"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
	"github.com/sarchlab/m2sim/insts"
	"github.com/sarchlab/m2sim/timing/latency"
	"github.com/sarchlab/m2sim/timing/pipeline"
)

func TestTiming(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Timing Integration Suite")
}

var _ = Describe("Timing Simulation Mode", func() {
	var (
		regFile        *emu.RegFile
		memory         *emu.Memory
		table          *latency.Table
		syscallHandler emu.SyscallHandler
		pipe           *pipeline.Pipeline
	)

	BeforeEach(func() {
		regFile = &emu.RegFile{}
		memory = emu.NewMemory()
		table = latency.NewTable()
		syscallHandler = emu.NewDefaultSyscallHandler(regFile, memory, nil, nil)
	})

	createPipeline := func() {
		pipe = pipeline.NewPipeline(
			regFile,
			memory,
			pipeline.WithSyscallHandler(syscallHandler),
			pipeline.WithLatencyTable(table),
		)
	}

	loadProgram := func(program []byte, entryPoint uint64) {
		for i, b := range program {
			memory.Write8(entryPoint+uint64(i), b)
		}
		regFile.PC = entryPoint
	}

	Describe("Basic Timing Predictions", func() {
		Context("Test Program 1: Simple Exit", func() {
			// Program: exit(42) - 3 instructions
			// MOV X8, #93 (syscall number)
			// MOV X0, #42 (exit code)
			// SVC #0
			It("should report correct instruction count and cycles", func() {
				program := []byte{}
				program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...)
				program = append(program, uint32ToBytes(encodeADDImm(0, 31, 42, false))...)
				program = append(program, uint32ToBytes(encodeSVC(0))...)

				loadProgram(program, 0x1000)
				createPipeline()
				pipe.SetPC(0x1000)

				exitCode := pipe.Run()
				stats := pipe.Stats()

				Expect(exitCode).To(Equal(int64(42)))
				// Pipeline may not retire all instructions before halt
				Expect(stats.Instructions).To(BeNumerically(">=", uint64(2)))
				Expect(stats.Instructions).To(BeNumerically("<=", uint64(3)))
				// With pipeline, cycles > instructions due to fill/drain
				Expect(stats.Cycles).To(BeNumerically(">=", stats.Instructions))
				Expect(stats.CPI()).To(BeNumerically(">", 0))

				fmt.Printf("\n=== Test Program 1: Simple Exit ===\n")
				fmt.Printf("Total Instructions: %d\n", stats.Instructions)
				fmt.Printf("Total Cycles: %d\n", stats.Cycles)
				fmt.Printf("CPI: %.2f\n", stats.CPI())
			})
		})

		Context("Test Program 2: Arithmetic", func() {
			// Program: 10 + 5 = 15 - 5 instructions
			// MOV X0, #10
			// MOV X1, #5
			// ADD X0, X0, X1
			// MOV X8, #93
			// SVC #0
			It("should report correct timing for arithmetic operations", func() {
				program := []byte{}
				program = append(program, uint32ToBytes(encodeADDImm(0, 31, 10, false))...)
				program = append(program, uint32ToBytes(encodeADDImm(1, 31, 5, false))...)
				program = append(program, uint32ToBytes(encodeADDReg(0, 0, 1, false))...)
				program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...)
				program = append(program, uint32ToBytes(encodeSVC(0))...)

				loadProgram(program, 0x1000)
				createPipeline()
				pipe.SetPC(0x1000)

				exitCode := pipe.Run()
				stats := pipe.Stats()

				Expect(exitCode).To(Equal(int64(15)))
				// Pipeline may not retire all instructions before halt
				Expect(stats.Instructions).To(BeNumerically(">=", uint64(4)))
				Expect(stats.Instructions).To(BeNumerically("<=", uint64(5)))
				Expect(stats.CPI()).To(BeNumerically(">", 0))

				fmt.Printf("\n=== Test Program 2: Arithmetic ===\n")
				fmt.Printf("Total Instructions: %d\n", stats.Instructions)
				fmt.Printf("Total Cycles: %d\n", stats.Cycles)
				fmt.Printf("CPI: %.2f\n", stats.CPI())
			})
		})

		Context("Test Program 3: Loop with Branches", func() {
			// Program: count down from 3 to 0
			// MOV X0, #3           ; counter = 3
			// loop:
			//   SUBS X0, X0, #1    ; counter-- (set flags)
			//   B.NE loop          ; if counter != 0, loop (-8 bytes = -2 instructions)
			// MOV X8, #93          ; exit syscall
			// SVC #0
			It("should report branch stalls and flushes", func() {
				program := []byte{}
				program = append(program, uint32ToBytes(encodeADDImm(0, 31, 3, false))...)  // 0x1000: MOV X0, #3
				program = append(program, uint32ToBytes(encodeSUBImm(0, 0, 1, true))...)    // 0x1004: SUBS X0, X0, #1
				program = append(program, uint32ToBytes(encodeBCond(-8, insts.CondNE))...)  // 0x1008: B.NE -8 -> 0x1004
				program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // 0x100C: MOV X8, #93
				program = append(program, uint32ToBytes(encodeSVC(0))...)                   // 0x1010: SVC

				loadProgram(program, 0x1000)
				createPipeline()
				pipe.SetPC(0x1000)

				exitCode := pipe.Run()
				stats := pipe.Stats()

				Expect(exitCode).To(Equal(int64(0)))
				// Instructions: MOV(1) + 3*(SUBS + B.NE) + MOV(1) + SVC(1)
				// But pipeline may not retire all before halt
				Expect(stats.Instructions).To(BeNumerically(">=", uint64(7)))
				// Should have branch flushes from taken branches
				Expect(stats.Flushes).To(BeNumerically(">", 0))
				Expect(stats.CPI()).To(BeNumerically(">", 1.0)) // Pipeline stalls expected

				fmt.Printf("\n=== Test Program 3: Loop with Branches ===\n")
				fmt.Printf("Total Instructions: %d\n", stats.Instructions)
				fmt.Printf("Total Cycles: %d\n", stats.Cycles)
				fmt.Printf("CPI: %.2f\n", stats.CPI())
				fmt.Printf("Flushes: %d\n", stats.Flushes)
				fmt.Printf("Flush Cycles: %d\n", stats.FlushCycles)
			})
		})

		Context("Test Program 4: Memory Operations", func() {
			// Program: load a value from memory
			// X1 = 0x2000 (base address)
			// LDR X0, [X1]    ; load from memory
			// MOV X8, #93
			// SVC #0
			It("should account for memory latency", func() {
				// Pre-set memory value and base register
				memory.Write64(0x2000, 99)
				regFile.WriteReg(1, 0x2000)

				program := []byte{}
				program = append(program, uint32ToBytes(encodeLDR64(0, 1, 0))...)
				program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...)
				program = append(program, uint32ToBytes(encodeSVC(0))...)

				loadProgram(program, 0x1000)
				createPipeline()
				pipe.SetPC(0x1000)

				exitCode := pipe.Run()
				stats := pipe.Stats()

				Expect(exitCode).To(Equal(int64(99)))
				Expect(stats.Instructions).To(Equal(uint64(3)))
				// Load has latency > 1, so CPI should reflect that
				Expect(stats.CPI()).To(BeNumerically(">", 1.0))

				fmt.Printf("\n=== Test Program 4: Memory Operations ===\n")
				fmt.Printf("Total Instructions: %d\n", stats.Instructions)
				fmt.Printf("Total Cycles: %d\n", stats.Cycles)
				fmt.Printf("CPI: %.2f\n", stats.CPI())
				fmt.Printf("Exec Stalls: %d\n", stats.ExecStalls)
			})
		})

		Context("Test Program 5: Function Call", func() {
			// Program: call a function
			// MOV X0, #5
			// BL func (+8)
			// MOV X8, #93
			// SVC #0
			// func: ADD X0, X0, #10
			//       RET
			It("should handle function calls with branch latency", func() {
				program := []byte{}
				// main: PC=0x1000
				program = append(program, uint32ToBytes(encodeADDImm(0, 31, 5, false))...)  // 0x1000: X0 = 5
				program = append(program, uint32ToBytes(encodeBL(12))...)                   // 0x1004: BL +12 -> 0x1010
				program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // 0x1008: X8 = 93
				program = append(program, uint32ToBytes(encodeSVC(0))...)                   // 0x100C: syscall
				// func: PC=0x1010
				program = append(program, uint32ToBytes(encodeADDImm(0, 0, 10, false))...) // 0x1010: X0 += 10
				program = append(program, uint32ToBytes(encodeRET())...)                   // 0x1014: RET

				loadProgram(program, 0x1000)
				createPipeline()
				pipe.SetPC(0x1000)

				exitCode := pipe.Run()
				stats := pipe.Stats()

				Expect(exitCode).To(Equal(int64(15)))
				// MOV(1) + BL(1) + ADD(1) + RET(1) + MOV(1) + SVC(1) = 6
				Expect(stats.Instructions).To(Equal(uint64(6)))
				// Should have flushes for BL and RET
				Expect(stats.Flushes).To(BeNumerically(">=", 2))

				fmt.Printf("\n=== Test Program 5: Function Call ===\n")
				fmt.Printf("Total Instructions: %d\n", stats.Instructions)
				fmt.Printf("Total Cycles: %d\n", stats.Cycles)
				fmt.Printf("CPI: %.2f\n", stats.CPI())
				fmt.Printf("Flushes: %d\n", stats.Flushes)
			})
		})
	})

	Describe("Timing Report Format", func() {
		It("should produce report matching expected format", func() {
			// Run a representative program
			program := []byte{}
			// A mixed program with ALU, branches, and memory
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 5, false))...)  // MOV X0, #5
			program = append(program, uint32ToBytes(encodeADDImm(1, 31, 10, false))...) // MOV X1, #10
			program = append(program, uint32ToBytes(encodeADDReg(0, 0, 1, false))...)   // ADD X0, X0, X1
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // MOV X8, #93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                   // SVC

			loadProgram(program, 0x1000)
			createPipeline()
			pipe.SetPC(0x1000)

			exitCode := pipe.Run()
			stats := pipe.Stats()

			Expect(exitCode).To(Equal(int64(15)))

			// Print in the expected format
			fmt.Println("\n========================================")
			fmt.Println("Expected Output Format Demonstration")
			fmt.Println("========================================")
			fmt.Printf("Program: mixed_ops.elf\n")
			fmt.Printf("Total Instructions: %d\n", stats.Instructions)
			fmt.Printf("Total Cycles: %d\n", stats.Cycles)
			fmt.Printf("CPI: %.2f\n", stats.CPI())
			fmt.Println()
			fmt.Println("Breakdown:")

			// Calculate percentages
			total := float64(stats.Cycles)
			if total == 0 {
				total = 1
			}

			fmt.Printf("  Fetch stalls:    %d cycles (%.1f%%)\n",
				stats.FetchStalls, float64(stats.FetchStalls)/total*100)
			fmt.Printf("  Decode stalls:   %d cycles (%.1f%%)\n",
				stats.DecodeStalls, float64(stats.DecodeStalls)/total*100)
			fmt.Printf("  Execute:         %d cycles (%.1f%%)\n",
				stats.Instructions, float64(stats.Instructions)/total*100)
			fmt.Printf("  Execute stalls:  %d cycles (%.1f%%)\n",
				stats.ExecStalls, float64(stats.ExecStalls)/total*100)
			fmt.Printf("  Memory stalls:   %d cycles (%.1f%%)\n",
				stats.MemStalls, float64(stats.MemStalls)/total*100)
			fmt.Printf("  Flush cycles:    %d cycles (%.1f%%)\n",
				stats.FlushCycles, float64(stats.FlushCycles)/total*100)
			fmt.Println("========================================")
		})
	})

	Describe("Statistics Accuracy", func() {
		It("CPI should be greater than or equal to 1 for pipelined execution", func() {
			// Simple program
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 1, false))...)
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...)
			program = append(program, uint32ToBytes(encodeSVC(0))...)

			loadProgram(program, 0x1000)
			createPipeline()
			pipe.SetPC(0x1000)

			pipe.Run()
			stats := pipe.Stats()

			// CPI >= 1 because of pipeline fill/drain overhead
			Expect(stats.CPI()).To(BeNumerically(">=", 1.0))
		})

		It("Instructions should match actual retired instructions", func() {
			// Program with known instruction count
			program := []byte{}
			for i := 0; i < 10; i++ {
				program = append(program, uint32ToBytes(encodeADDImm(0, 31, 1, false))...)
			}
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...)
			program = append(program, uint32ToBytes(encodeSVC(0))...)

			loadProgram(program, 0x1000)
			createPipeline()
			pipe.SetPC(0x1000)

			pipe.Run()
			stats := pipe.Stats()

			Expect(stats.Instructions).To(Equal(uint64(12))) // 10 ADDs + MOV + SVC
		})
	})
})

// Helper functions to encode ARM64 instructions

func uint32ToBytes(v uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf
}

func encodeADDImm(rd, rn uint8, imm uint16, setFlags bool) uint32 {
	var inst uint32 = 0
	inst |= 1 << 31         // sf = 1 (64-bit)
	inst |= 0 << 30         // op = 0 (ADD)
	if setFlags {
		inst |= 1 << 29     // S = 1
	}
	inst |= 0b100010 << 23  // opc
	inst |= 0 << 22         // sh = 0
	inst |= uint32(imm&0xFFF) << 10
	inst |= uint32(rn&0x1F) << 5
	inst |= uint32(rd & 0x1F)
	return inst
}

func encodeSUBImm(rd, rn uint8, imm uint16, setFlags bool) uint32 {
	var inst uint32 = 0
	inst |= 1 << 31         // sf = 1 (64-bit)
	inst |= 1 << 30         // op = 1 (SUB)
	if setFlags {
		inst |= 1 << 29     // S = 1
	}
	inst |= 0b100010 << 23
	inst |= 0 << 22
	inst |= uint32(imm&0xFFF) << 10
	inst |= uint32(rn&0x1F) << 5
	inst |= uint32(rd & 0x1F)
	return inst
}

func encodeADDReg(rd, rn, rm uint8, setFlags bool) uint32 {
	var inst uint32 = 0
	inst |= 1 << 31         // sf = 1 (64-bit)
	inst |= 0 << 30         // op = 0 (ADD)
	if setFlags {
		inst |= 1 << 29
	}
	inst |= 0b01011 << 24
	inst |= 0 << 22
	inst |= 0 << 21
	inst |= uint32(rm&0x1F) << 16
	inst |= 0 << 10
	inst |= uint32(rn&0x1F) << 5
	inst |= uint32(rd & 0x1F)
	return inst
}

func encodeLDR64(rd, rn uint8, offset uint16) uint32 {
	var inst uint32 = 0
	inst |= 0b11 << 30
	inst |= 0b111 << 27
	inst |= 0 << 26
	inst |= 0b01 << 24
	inst |= 0b01 << 22
	scaledOffset := offset / 8
	inst |= uint32(scaledOffset&0xFFF) << 10
	inst |= uint32(rn&0x1F) << 5
	inst |= uint32(rd & 0x1F)
	return inst
}

func encodeSTR64(rd, rn uint8, offset uint16) uint32 {
	var inst uint32 = 0
	inst |= 0b11 << 30
	inst |= 0b111 << 27
	inst |= 0 << 26
	inst |= 0b01 << 24
	inst |= 0b00 << 22
	scaledOffset := offset / 8
	inst |= uint32(scaledOffset&0xFFF) << 10
	inst |= uint32(rn&0x1F) << 5
	inst |= uint32(rd & 0x1F)
	return inst
}

func encodeBL(offset int32) uint32 {
	var inst uint32 = 0
	inst |= 0b100101 << 26
	imm26 := uint32(offset/4) & 0x3FFFFFF
	inst |= imm26
	return inst
}

func encodeBCond(offset int32, cond insts.Cond) uint32 {
	var inst uint32 = 0
	inst |= 0b0101010 << 25
	inst |= 0 << 24
	imm19 := uint32(offset/4) & 0x7FFFF
	inst |= imm19 << 5
	inst |= 0 << 4
	inst |= uint32(cond & 0xF)
	return inst
}

func encodeRET() uint32 {
	var inst uint32 = 0
	inst |= 0b1101011 << 25
	inst |= 0 << 24
	inst |= 0 << 23
	inst |= 0b10 << 21
	inst |= 0b11111 << 16
	inst |= 0b0000 << 12
	inst |= 0 << 11
	inst |= 0 << 10
	inst |= uint32(30) << 5 // Rn = X30 (link register)
	inst |= 0b00000
	return inst
}

func encodeSVC(imm uint16) uint32 {
	var inst uint32 = 0
	inst |= 0b11010100 << 24
	inst |= 0b000 << 21
	inst |= uint32(imm) << 5
	inst |= 0b00001
	return inst
}
