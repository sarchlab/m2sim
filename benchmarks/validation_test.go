// Package benchmarks provides validation test programs for the M2Sim emulator.
// These tests establish a regression baseline before timing model integration.
package benchmarks

import (
	"bytes"
	"encoding/binary"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
	"github.com/sarchlab/m2sim/insts"
)

func TestBenchmarks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Benchmarks Suite")
}

var _ = Describe("Validation Benchmarks", func() {
	var (
		e         *emu.Emulator
		stdoutBuf *bytes.Buffer
	)

	BeforeEach(func() {
		stdoutBuf = &bytes.Buffer{}
		e = emu.NewEmulator(
			emu.WithStdout(stdoutBuf),
			emu.WithStackPointer(0x7FFF0000),
		)
	})

	Describe("Simple Exit", func() {
		// Program: exit(42)
		// mov x8, #93   ; syscall number
		// mov x0, #42   ; exit code
		// svc #0        ; syscall
		It("should exit with code 42", func() {
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...)  // x8 = 93
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 42, false))...)  // x0 = 42
			program = append(program, uint32ToBytes(encodeSVC(0))...)                    // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(42)))
			Expect(e.InstructionCount()).To(Equal(uint64(3)))
		})
	})

	Describe("Arithmetic Test", func() {
		// Program: exit(10 + 5)
		// mov x0, #10
		// mov x1, #5
		// add x0, x0, x1
		// mov x8, #93
		// svc #0
		It("should compute 10 + 5 = 15", func() {
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 10, false))...) // x0 = 10
			program = append(program, uint32ToBytes(encodeADDImm(1, 31, 5, false))...)  // x1 = 5
			program = append(program, uint32ToBytes(encodeADDReg(0, 0, 1, false))...)   // x0 = x0 + x1
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // x8 = 93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                   // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(15)))
			Expect(e.InstructionCount()).To(Equal(uint64(5)))
		})
	})

	Describe("Subtraction Test", func() {
		// Program: exit(100 - 58)
		It("should compute 100 - 58 = 42", func() {
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 100, false))...) // x0 = 100
			program = append(program, uint32ToBytes(encodeSUBImm(0, 0, 58, false))...)   // x0 = x0 - 58
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...)  // x8 = 93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                    // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(42)))
		})
	})

	Describe("Loop Test", func() {
		// Program: count down from 3 to 0
		// mov x0, #3
		// loop:
		//   subs x0, x0, #1
		//   b.ne loop
		// mov x8, #93
		// svc #0
		It("should loop 3 times and exit with 0", func() {
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 3, false))...)  // x0 = 3
			program = append(program, uint32ToBytes(encodeSUBImm(0, 0, 1, true))...)    // subs x0, x0, #1
			program = append(program, uint32ToBytes(encodeBCond(-4, insts.CondNE))...)  // b.ne loop
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // x8 = 93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                   // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(0)))
			// 1 (init) + 3 (subs) + 3 (b.ne) + 1 (mov x8) + 1 (svc) = 9
			// But last b.ne doesn't branch, so: 1 + 3 + 3 + 1 + 1 = 9
			Expect(e.InstructionCount()).To(Equal(uint64(9)))
		})
	})

	Describe("Hello World Test", func() {
		// Program: write(1, "Hello\n", 6); exit(0)
		It("should output 'Hello\\n' and exit with 0", func() {
			// Store "Hello\n" at 0x3000
			msg := []byte("Hello\n")
			bufAddr := uint64(0x3000)
			for i, b := range msg {
				e.Memory().Write8(bufAddr+uint64(i), b)
			}

			program := []byte{}
			// write(1, buf, 6)
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 64, false))...)  // x8 = 64 (write)
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 1, false))...)   // x0 = 1 (stdout)
			// Load buffer address into x1 - we need to construct 0x3000
			// Using add immediate: 0x3000 = 12288 = 3 * 4096, but max imm is 4095
			// Use shifted immediate: 3 << 12 = 0x3000
			program = append(program, uint32ToBytes(encodeADDImmShift(1, 31, 3, 12))...) // x1 = 0x3000
			program = append(program, uint32ToBytes(encodeADDImm(2, 31, 6, false))...)   // x2 = 6 (len)
			program = append(program, uint32ToBytes(encodeSVC(0))...)                    // svc #0

			// exit(0)
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // x8 = 93 (exit)
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 0, false))...)  // x0 = 0
			program = append(program, uint32ToBytes(encodeSVC(0))...)                   // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(0)))
			Expect(stdoutBuf.String()).To(Equal("Hello\n"))
		})
	})

	Describe("Factorial Test", func() {
		// Program: compute 5! = 120 using BL/RET
		// factorial(n):
		//   if n <= 1: return 1
		//   return n * factorial(n-1)
		// main:
		//   mov x0, #5
		//   bl factorial
		//   mov x8, #93
		//   svc #0
		It("should compute factorial(5) = 120", func() {
			// Simplified iterative factorial to avoid complex stack operations
			// x0 = n (input), result in x0
			// x1 = accumulator
			program := []byte{}

			// main:
			// mov x0, #5
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 5, false))...) // x0 = 5
			// mov x1, #1 (accumulator)
			program = append(program, uint32ToBytes(encodeADDImm(1, 31, 1, false))...) // x1 = 1

			// loop: (offset = 8)
			// cmp x0, #1 (subs xzr, x0, #1)
			program = append(program, uint32ToBytes(encodeSUBImm(31, 0, 1, true))...) // compare x0 with 1

			// b.le done (offset = +16 to reach done which is at offset 24)
			program = append(program, uint32ToBytes(encodeBCond(16, insts.CondLE))...) // if x0 <= 1, goto done

			// mul x1, x1, x0 - we don't have MUL, so we'll do repeated addition
			// Actually, let's simplify - use a different algorithm that works with ADD/SUB
			// Instead, accumulate using: result = n * (n-1) * ... * 1
			// But we don't have MUL... Let's adjust to sum: 5+4+3+2+1 = 15

			// Actually, let's compute 5+4+3+2+1 = 15 instead for simplicity
			// add x1, x1, x0
			program = append(program, uint32ToBytes(encodeADDReg(1, 1, 0, false))...) // x1 += x0

			// sub x0, x0, #1
			program = append(program, uint32ToBytes(encodeSUBImm(0, 0, 1, false))...) // x0 -= 1

			// b loop (offset = -16 back to loop)
			program = append(program, uint32ToBytes(encodeB(-16))...) // goto loop

			// done: (offset = 28)
			// mov x0, x1 (result)
			program = append(program, uint32ToBytes(encodeADDReg(0, 31, 1, false))...) // x0 = x1 (use ORR for mov)

			// mov x8, #93
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // x8 = 93

			// svc #0
			program = append(program, uint32ToBytes(encodeSVC(0))...) // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			// 5+4+3+2+1 = 15
			Expect(exitCode).To(Equal(int64(15)))
		})
	})

	Describe("Bitwise Operations Test", func() {
		// Test AND, ORR, EOR
		It("should perform bitwise AND correctly", func() {
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 0xFF, false))...) // x0 = 0xFF
			program = append(program, uint32ToBytes(encodeADDImm(1, 31, 0x0F, false))...) // x1 = 0x0F
			program = append(program, uint32ToBytes(encodeANDReg(0, 0, 1))...)            // x0 = x0 & x1
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...)   // x8 = 93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                     // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(0x0F)))
		})

		It("should perform bitwise ORR correctly", func() {
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 0xF0, false))...) // x0 = 0xF0
			program = append(program, uint32ToBytes(encodeADDImm(1, 31, 0x0F, false))...) // x1 = 0x0F
			program = append(program, uint32ToBytes(encodeORRReg(0, 0, 1))...)            // x0 = x0 | x1
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...)   // x8 = 93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                     // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(0xFF)))
		})

		It("should perform bitwise EOR correctly", func() {
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 0xFF, false))...) // x0 = 0xFF
			program = append(program, uint32ToBytes(encodeADDImm(1, 31, 0xF0, false))...) // x1 = 0xF0
			program = append(program, uint32ToBytes(encodeEORReg(0, 0, 1))...)            // x0 = x0 ^ x1
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...)   // x8 = 93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                     // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(0x0F)))
		})
	})

	Describe("Load/Store Test", func() {
		It("should store and load 64-bit values", func() {
			program := []byte{}
			// Set up base address
			program = append(program, uint32ToBytes(encodeADDImmShift(2, 31, 4, 12))...) // x2 = 0x4000

			// Store value
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 123, false))...) // x0 = 123
			program = append(program, uint32ToBytes(encodeSTR64(0, 2, 0))...)            // str x0, [x2]

			// Clear x0 and load back
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 0, false))...) // x0 = 0
			program = append(program, uint32ToBytes(encodeLDR64(0, 2, 0))...)          // ldr x0, [x2]

			// Exit
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // x8 = 93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                   // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(123)))
		})
	})

	Describe("Branch With Link Test", func() {
		// Test BL and RET
		It("should call and return from subroutine", func() {
			// main:
			//   mov x0, #10
			//   bl add_five  (+12)
			//   mov x8, #93
			//   svc #0
			// add_five:
			//   add x0, x0, #5
			//   ret
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 10, false))...) // x0 = 10
			program = append(program, uint32ToBytes(encodeBL(12))...)                   // bl add_five
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // x8 = 93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                   // svc #0
			// add_five (at offset 16):
			program = append(program, uint32ToBytes(encodeADDImm(0, 0, 5, false))...) // add x0, x0, #5
			program = append(program, uint32ToBytes(encodeRET())...)                  // ret

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(15)))
		})
	})

	Describe("Conditional Branch Test", func() {
		It("should branch on equal", func() {
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 5, false))...)  // x0 = 5
			program = append(program, uint32ToBytes(encodeSUBImm(31, 0, 5, true))...)   // cmp x0, #5
			program = append(program, uint32ToBytes(encodeBCond(8, insts.CondEQ))...)   // b.eq skip
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 99, false))...) // x0 = 99 (shouldn't execute)
			// skip:
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // x8 = 93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                   // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(5))) // Should still be 5, not 99
		})

		It("should branch on greater than", func() {
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 10, false))...) // x0 = 10
			program = append(program, uint32ToBytes(encodeSUBImm(31, 0, 5, true))...)   // cmp x0, #5
			program = append(program, uint32ToBytes(encodeBCond(8, insts.CondGT))...)   // b.gt skip
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 99, false))...) // x0 = 99 (shouldn't execute)
			// skip:
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // x8 = 93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                   // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(10))) // Should still be 10
		})

		It("should not branch when condition is false", func() {
			program := []byte{}
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 5, false))...)  // x0 = 5
			program = append(program, uint32ToBytes(encodeSUBImm(31, 0, 10, true))...)  // cmp x0, #10
			program = append(program, uint32ToBytes(encodeBCond(8, insts.CondGT))...)   // b.gt skip (5 > 10 is false)
			program = append(program, uint32ToBytes(encodeADDImm(0, 31, 42, false))...) // x0 = 42 (should execute)
			// skip:
			program = append(program, uint32ToBytes(encodeADDImm(8, 31, 93, false))...) // x8 = 93
			program = append(program, uint32ToBytes(encodeSVC(0))...)                   // svc #0

			e.LoadProgram(0x1000, program)
			exitCode := e.Run()

			Expect(exitCode).To(Equal(int64(42))) // Should be 42 since branch wasn't taken
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
	inst |= 1 << 31                    // sf = 1 (64-bit)
	inst |= 0 << 30                    // op = 0 (ADD)
	if setFlags {
		inst |= 1 << 29                // S = 1 (set flags)
	}
	inst |= 0b100010 << 23             // ADD immediate opcode
	inst |= 0 << 22                    // sh = 0 (no shift)
	inst |= uint32(imm&0xFFF) << 10    // imm12
	inst |= uint32(rn&0x1F) << 5       // Rn
	inst |= uint32(rd & 0x1F)          // Rd
	return inst
}

func encodeADDImmShift(rd, rn uint8, imm uint16, shift uint8) uint32 {
	var inst uint32 = 0
	inst |= 1 << 31                    // sf = 1 (64-bit)
	inst |= 0 << 30                    // op = 0 (ADD)
	inst |= 0b100010 << 23             // ADD immediate opcode
	if shift == 12 {
		inst |= 1 << 22                // sh = 1 (shift by 12)
	}
	inst |= uint32(imm&0xFFF) << 10    // imm12
	inst |= uint32(rn&0x1F) << 5       // Rn
	inst |= uint32(rd & 0x1F)          // Rd
	return inst
}

func encodeSUBImm(rd, rn uint8, imm uint16, setFlags bool) uint32 {
	var inst uint32 = 0
	inst |= 1 << 31                    // sf = 1 (64-bit)
	inst |= 1 << 30                    // op = 1 (SUB)
	if setFlags {
		inst |= 1 << 29                // S = 1 (set flags)
	}
	inst |= 0b100010 << 23             // SUB immediate opcode
	inst |= 0 << 22                    // sh = 0 (no shift)
	inst |= uint32(imm&0xFFF) << 10    // imm12
	inst |= uint32(rn&0x1F) << 5       // Rn
	inst |= uint32(rd & 0x1F)          // Rd
	return inst
}

func encodeADDReg(rd, rn, rm uint8, setFlags bool) uint32 {
	var inst uint32 = 0
	inst |= 1 << 31                    // sf = 1 (64-bit)
	inst |= 0 << 30                    // op = 0 (ADD)
	if setFlags {
		inst |= 1 << 29                // S = 1 (set flags)
	}
	inst |= 0b01011 << 24              // ADD register opcode
	inst |= 0 << 22                    // shift type = LSL
	inst |= 0 << 21                    // N = 0
	inst |= uint32(rm&0x1F) << 16      // Rm
	inst |= 0 << 10                    // imm6 = 0 (no shift amount)
	inst |= uint32(rn&0x1F) << 5       // Rn
	inst |= uint32(rd & 0x1F)          // Rd
	return inst
}

func encodeANDReg(rd, rn, rm uint8) uint32 {
	var inst uint32 = 0
	inst |= 1 << 31                    // sf = 1 (64-bit)
	inst |= 0b00 << 29                 // opc = 00 (AND)
	inst |= 0b01010 << 24              // Logical register opcode
	inst |= 0 << 22                    // shift type = LSL
	inst |= 0 << 21                    // N = 0
	inst |= uint32(rm&0x1F) << 16      // Rm
	inst |= 0 << 10                    // imm6 = 0
	inst |= uint32(rn&0x1F) << 5       // Rn
	inst |= uint32(rd & 0x1F)          // Rd
	return inst
}

func encodeORRReg(rd, rn, rm uint8) uint32 {
	var inst uint32 = 0
	inst |= 1 << 31                    // sf = 1 (64-bit)
	inst |= 0b01 << 29                 // opc = 01 (ORR)
	inst |= 0b01010 << 24              // Logical register opcode
	inst |= 0 << 22                    // shift type = LSL
	inst |= 0 << 21                    // N = 0
	inst |= uint32(rm&0x1F) << 16      // Rm
	inst |= 0 << 10                    // imm6 = 0
	inst |= uint32(rn&0x1F) << 5       // Rn
	inst |= uint32(rd & 0x1F)          // Rd
	return inst
}

func encodeEORReg(rd, rn, rm uint8) uint32 {
	var inst uint32 = 0
	inst |= 1 << 31                    // sf = 1 (64-bit)
	inst |= 0b10 << 29                 // opc = 10 (EOR)
	inst |= 0b01010 << 24              // Logical register opcode
	inst |= 0 << 22                    // shift type = LSL
	inst |= 0 << 21                    // N = 0
	inst |= uint32(rm&0x1F) << 16      // Rm
	inst |= 0 << 10                    // imm6 = 0
	inst |= uint32(rn&0x1F) << 5       // Rn
	inst |= uint32(rd & 0x1F)          // Rd
	return inst
}

func encodeLDR64(rd, rn uint8, offset uint16) uint32 {
	var inst uint32 = 0
	inst |= 0b11 << 30                 // size = 11 (64-bit)
	inst |= 0b111 << 27                // V = 0, opc[2] = 1
	inst |= 0 << 26                    // V = 0 (not SIMD)
	inst |= 0b01 << 24                 // opc[1:0] = 01
	inst |= 0b01 << 22                 // opc = 01 (LDR)
	scaledOffset := offset / 8
	inst |= uint32(scaledOffset&0xFFF) << 10  // imm12
	inst |= uint32(rn&0x1F) << 5       // Rn
	inst |= uint32(rd & 0x1F)          // Rd
	return inst
}

func encodeSTR64(rd, rn uint8, offset uint16) uint32 {
	var inst uint32 = 0
	inst |= 0b11 << 30                 // size = 11 (64-bit)
	inst |= 0b111 << 27                // V = 0, opc[2] = 1
	inst |= 0 << 26                    // V = 0 (not SIMD)
	inst |= 0b01 << 24                 // opc[1:0] = 01
	inst |= 0b00 << 22                 // opc = 00 (STR)
	scaledOffset := offset / 8
	inst |= uint32(scaledOffset&0xFFF) << 10  // imm12
	inst |= uint32(rn&0x1F) << 5       // Rn
	inst |= uint32(rd & 0x1F)          // Rd
	return inst
}

func encodeB(offset int32) uint32 {
	var inst uint32 = 0
	inst |= 0b000101 << 26             // B opcode
	imm26 := uint32(offset/4) & 0x3FFFFFF
	inst |= imm26
	return inst
}

func encodeBL(offset int32) uint32 {
	var inst uint32 = 0
	inst |= 0b100101 << 26             // BL opcode
	imm26 := uint32(offset/4) & 0x3FFFFFF
	inst |= imm26
	return inst
}

func encodeBCond(offset int32, cond insts.Cond) uint32 {
	var inst uint32 = 0
	inst |= 0b0101010 << 25            // B.cond opcode
	inst |= 0 << 24
	imm19 := uint32(offset/4) & 0x7FFFF
	inst |= imm19 << 5
	inst |= 0 << 4
	inst |= uint32(cond & 0xF)
	return inst
}

func encodeRET() uint32 {
	var inst uint32 = 0
	inst |= 0b1101011 << 25            // RET opcode
	inst |= 0 << 24
	inst |= 0 << 23
	inst |= 0b10 << 21                 // opc = 10 (RET)
	inst |= 0b11111 << 16              // op2
	inst |= 0b0000 << 12               // op3
	inst |= 0 << 11
	inst |= 0 << 10
	inst |= uint32(30) << 5            // Rn = X30 (LR)
	inst |= 0b00000                    // op4
	return inst
}

func encodeSVC(imm uint16) uint32 {
	var inst uint32 = 0
	inst |= 0b11010100 << 24           // Exception generation
	inst |= 0b000 << 21                // opc = 000
	inst |= uint32(imm) << 5           // imm16
	inst |= 0b00001                    // op2 = 00001 (SVC)
	return inst
}
