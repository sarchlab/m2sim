package integration_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
	"github.com/sarchlab/m2sim/loader"
)

var _ = Describe("End-to-End Integration Tests", func() {
	var (
		tempDir   string
		stdoutBuf *bytes.Buffer
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "m2sim-integration-test")
		Expect(err).NotTo(HaveOccurred())

		stdoutBuf = &bytes.Buffer{}
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("exit_zero program", func() {
		It("should exit with code 0", func() {
			// Create a program that:
			// MOV X8, #93      ; syscall number for exit
			// MOV X0, #0       ; exit code 0
			// SVC #0           ; invoke syscall
			elfPath := filepath.Join(tempDir, "exit_zero.elf")
			code := buildExitProgram(0)
			createTestELF(elfPath, 0x400000, 0x400000, code)

			// Load the ELF
			prog, err := loader.Load(elfPath)
			Expect(err).NotTo(HaveOccurred())

			// Create emulator and load program
			e := emu.NewEmulator(
				emu.WithStdout(stdoutBuf),
				emu.WithStackPointer(prog.InitialSP),
			)
			loadProgramIntoEmulator(e, prog)

			// Run and verify
			exitCode := e.Run()
			Expect(exitCode).To(Equal(int64(0)))
		})
	})

	Describe("exit_42 program", func() {
		It("should exit with code 42", func() {
			// Create a program that exits with code 42
			elfPath := filepath.Join(tempDir, "exit_42.elf")
			code := buildExitProgram(42)
			createTestELF(elfPath, 0x400000, 0x400000, code)

			// Load the ELF
			prog, err := loader.Load(elfPath)
			Expect(err).NotTo(HaveOccurred())

			// Create emulator and load program
			e := emu.NewEmulator(
				emu.WithStdout(stdoutBuf),
				emu.WithStackPointer(prog.InitialSP),
			)
			loadProgramIntoEmulator(e, prog)

			// Run and verify
			exitCode := e.Run()
			Expect(exitCode).To(Equal(int64(42)))
		})
	})

	Describe("hello program", func() {
		It("should print 'Hello, World!' and exit with code 0", func() {
			// Create a program that:
			// 1. Writes "Hello, World!\n" to stdout
			// 2. Exits with code 0
			elfPath := filepath.Join(tempDir, "hello.elf")
			message := "Hello, World!\n"
			code, dataAddr := buildHelloProgram(0x400000, message)
			createHelloELF(elfPath, 0x400000, 0x400000, code, dataAddr, []byte(message))

			// Load the ELF
			prog, err := loader.Load(elfPath)
			Expect(err).NotTo(HaveOccurred())

			// Create emulator and load program
			e := emu.NewEmulator(
				emu.WithStdout(stdoutBuf),
				emu.WithStackPointer(prog.InitialSP),
			)
			loadProgramIntoEmulator(e, prog)

			// Run and verify
			exitCode := e.Run()
			Expect(exitCode).To(Equal(int64(0)))
			Expect(stdoutBuf.String()).To(Equal("Hello, World!\n"))
		})
	})

	Describe("computation program", func() {
		It("should compute 10 + 5 and exit with result", func() {
			// Create a program that computes 10 + 5 and exits with that value
			elfPath := filepath.Join(tempDir, "compute.elf")
			code := buildComputeProgram(10, 5)
			createTestELF(elfPath, 0x400000, 0x400000, code)

			// Load the ELF
			prog, err := loader.Load(elfPath)
			Expect(err).NotTo(HaveOccurred())

			// Create emulator and load program
			e := emu.NewEmulator(
				emu.WithStdout(stdoutBuf),
				emu.WithStackPointer(prog.InitialSP),
			)
			loadProgramIntoEmulator(e, prog)

			// Run and verify
			exitCode := e.Run()
			Expect(exitCode).To(Equal(int64(15)))
		})
	})

	Describe("loop program", func() {
		It("should count down from 5 to 0 and exit with 0", func() {
			// Create a program with a loop that counts down
			elfPath := filepath.Join(tempDir, "loop.elf")
			code := buildLoopProgram(5)
			createTestELF(elfPath, 0x400000, 0x400000, code)

			// Load the ELF
			prog, err := loader.Load(elfPath)
			Expect(err).NotTo(HaveOccurred())

			// Create emulator and load program
			e := emu.NewEmulator(
				emu.WithStdout(stdoutBuf),
				emu.WithStackPointer(prog.InitialSP),
				emu.WithMaxInstructions(1000), // Safety limit
			)
			loadProgramIntoEmulator(e, prog)

			// Run and verify
			exitCode := e.Run()
			Expect(exitCode).To(Equal(int64(0)))
		})
	})

	Describe("function call program", func() {
		It("should call a function and return correctly", func() {
			// Create a program that uses BL/RET for function call
			elfPath := filepath.Join(tempDir, "funcall.elf")
			code := buildFunctionCallProgram()
			createTestELF(elfPath, 0x400000, 0x400000, code)

			// Load the ELF
			prog, err := loader.Load(elfPath)
			Expect(err).NotTo(HaveOccurred())

			// Create emulator and load program
			e := emu.NewEmulator(
				emu.WithStdout(stdoutBuf),
				emu.WithStackPointer(prog.InitialSP),
			)
			loadProgramIntoEmulator(e, prog)

			// Run and verify
			exitCode := e.Run()
			// Function returns 100, which becomes exit code
			Expect(exitCode).To(Equal(int64(100)))
		})
	})

	Describe("multi-segment ELF", func() {
		It("should load data from separate data segment", func() {
			// Create a program that loads a value from a data segment
			elfPath := filepath.Join(tempDir, "multidata.elf")
			code := buildLoadDataProgram(0x600000)
			data := uint64ToBytes(uint64(77)) // Value to load: 77
			createMultiSegmentTestELF(elfPath, 0x400000, 0x400000, code, 0x600000, data)

			// Load the ELF
			prog, err := loader.Load(elfPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(prog.Segments).To(HaveLen(2))

			// Create emulator and load program
			e := emu.NewEmulator(
				emu.WithStdout(stdoutBuf),
				emu.WithStackPointer(prog.InitialSP),
			)
			loadProgramIntoEmulator(e, prog)

			// Run and verify - program loads 77 from data segment and exits with it
			exitCode := e.Run()
			Expect(exitCode).To(Equal(int64(77)))
		})
	})
})

// loadProgramIntoEmulator loads all segments from an ELF program into the emulator's memory.
func loadProgramIntoEmulator(e *emu.Emulator, prog *loader.Program) {
	mem := e.Memory()

	// Load each segment
	for _, seg := range prog.Segments {
		// Load segment data
		for i, b := range seg.Data {
			mem.Write8(seg.VirtAddr+uint64(i), b)
		}

		// Zero-fill BSS area (MemSize > len(Data))
		for i := uint64(len(seg.Data)); i < seg.MemSize; i++ {
			mem.Write8(seg.VirtAddr+i, 0)
		}
	}

	// Set entry point
	e.RegFile().PC = prog.EntryPoint
}

// Helper functions to build ARM64 programs

// buildExitProgram creates code that exits with the given exit code.
func buildExitProgram(exitCode uint16) []byte {
	program := []byte{}
	// MOV X8, #93 (exit syscall)
	program = append(program, encodeMovImm(8, 93)...)
	// MOV X0, #exitCode
	program = append(program, encodeMovImm(0, exitCode)...)
	// SVC #0
	program = append(program, encodeSVC(0)...)
	return program
}

// buildComputeProgram creates code that computes a + b and exits with the result.
func buildComputeProgram(a, b uint16) []byte {
	program := []byte{}
	// MOV X0, #a
	program = append(program, encodeMovImm(0, a)...)
	// MOV X1, #b
	program = append(program, encodeMovImm(1, b)...)
	// ADD X0, X0, X1
	program = append(program, encodeADDReg(0, 0, 1)...)
	// MOV X8, #93 (exit syscall)
	program = append(program, encodeMovImm(8, 93)...)
	// SVC #0
	program = append(program, encodeSVC(0)...)
	return program
}

// buildLoopProgram creates code that counts down from n to 0.
func buildLoopProgram(n uint16) []byte {
	program := []byte{}
	// MOV X0, #n
	program = append(program, encodeMovImm(0, n)...)
	// loop:
	// SUBS X0, X0, #1 (with flags)
	program = append(program, encodeSUBSImm(0, 0, 1)...)
	// B.NE loop (offset -4, which is -1 instruction)
	program = append(program, encodeBCond(-4, 0x1)...) // CondNE = 0x1
	// MOV X8, #93
	program = append(program, encodeMovImm(8, 93)...)
	// SVC #0
	program = append(program, encodeSVC(0)...)
	return program
}

// buildFunctionCallProgram creates code that calls a function and returns.
func buildFunctionCallProgram() []byte {
	program := []byte{}
	// main:
	// BL func (offset +12 bytes = +3 instructions)
	program = append(program, encodeBL(12)...)
	// After return, X0 has return value
	// MOV X8, #93
	program = append(program, encodeMovImm(8, 93)...)
	// SVC #0
	program = append(program, encodeSVC(0)...)
	// func:
	// MOV X0, #100
	program = append(program, encodeMovImm(0, 100)...)
	// RET
	program = append(program, encodeRET()...)
	return program
}

// buildHelloProgram creates code that writes a message to stdout.
// Returns the code bytes and the data address where the message should be placed.
func buildHelloProgram(codeBase uint64, message string) ([]byte, uint64) {
	// Data will be placed after code, at address 0x600000
	// 0x600000 = 0x600 << 12, so we can use ADD with shift=12
	dataAddr := uint64(0x600000)
	msgLen := uint16(len(message))

	program := []byte{}
	// MOV X8, #64 (write syscall)
	program = append(program, encodeMovImm(8, 64)...)
	// MOV X0, #1 (stdout fd)
	program = append(program, encodeMovImm(0, 1)...)
	// Load address of message into X1 using ADD with shift
	// ADD X1, XZR, #0x600, LSL #12 = 0x600000
	program = append(program, encodeADDImmShift(1, 31, 0x600, 12)...)
	// MOV X2, #msgLen
	program = append(program, encodeMovImm(2, msgLen)...)
	// SVC #0
	program = append(program, encodeSVC(0)...)
	// MOV X8, #93 (exit syscall)
	program = append(program, encodeMovImm(8, 93)...)
	// MOV X0, #0 (exit code)
	program = append(program, encodeMovImm(0, 0)...)
	// SVC #0
	program = append(program, encodeSVC(0)...)

	return program, dataAddr
}

// buildLoadDataProgram creates code that loads a 64-bit value from dataAddr and exits with it.
func buildLoadDataProgram(dataAddr uint64) []byte {
	// dataAddr = 0x600000 = 0x600 << 12
	program := []byte{}
	// Load address into X1 using ADD with shift
	// ADD X1, XZR, #0x600, LSL #12 = 0x600000
	program = append(program, encodeADDImmShift(1, 31, 0x600, 12)...)
	// LDR X0, [X1]
	program = append(program, encodeLDR64(0, 1, 0)...)
	// MOV X8, #93
	program = append(program, encodeMovImm(8, 93)...)
	// SVC #0
	program = append(program, encodeSVC(0)...)
	return program
}

// ARM64 instruction encoding helpers

func uint32ToBytes(v uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf
}

func uint64ToBytes(v uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, v)
	return buf
}

// encodeMovImm encodes: ADD Xd, XZR, #imm (using ADD from ZR as MOV equivalent)
func encodeMovImm(rd uint8, imm uint16) []byte {
	return encodeADDImmShift(rd, 31, imm, 0)
}

// encodeADDImmShift encodes: ADD Xd, Xn, #imm, LSL #shift
// shift must be 0 or 12
func encodeADDImmShift(rd, rn uint8, imm uint16, shift uint8) []byte {
	var inst uint32 = 0
	inst |= 1 << 31        // sf = 1 (64-bit)
	inst |= 0 << 30        // op = 0 (ADD)
	inst |= 0 << 29        // S = 0
	inst |= 0b100010 << 23 // opc for ADD immediate
	if shift == 12 {
		inst |= 1 << 22 // sh = 1 (shift by 12)
	}
	inst |= uint32(imm&0xFFF) << 10
	inst |= uint32(rn&0x1F) << 5
	inst |= uint32(rd & 0x1F)
	return uint32ToBytes(inst)
}

// encodeADDReg encodes: ADD Xd, Xn, Xm
func encodeADDReg(rd, rn, rm uint8) []byte {
	var inst uint32 = 0
	inst |= 1 << 31       // sf = 1 (64-bit)
	inst |= 0 << 30       // op = 0 (ADD)
	inst |= 0 << 29       // S = 0
	inst |= 0b01011 << 24 // opc for ADD register
	inst |= 0 << 22       // shift type
	inst |= 0 << 21       // 0
	inst |= uint32(rm&0x1F) << 16
	inst |= 0 << 10 // imm6
	inst |= uint32(rn&0x1F) << 5
	inst |= uint32(rd & 0x1F)
	return uint32ToBytes(inst)
}

// encodeSUBSImm encodes: SUBS Xd, Xn, #imm (SUB with flags)
func encodeSUBSImm(rd, rn uint8, imm uint16) []byte {
	var inst uint32 = 0
	inst |= 1 << 31        // sf = 1 (64-bit)
	inst |= 1 << 30        // op = 1 (SUB)
	inst |= 1 << 29        // S = 1 (set flags)
	inst |= 0b100010 << 23 // opc for SUB immediate
	inst |= 0 << 22        // shift
	inst |= uint32(imm&0xFFF) << 10
	inst |= uint32(rn&0x1F) << 5
	inst |= uint32(rd & 0x1F)
	return uint32ToBytes(inst)
}

// encodeBL encodes: BL offset
func encodeBL(offset int32) []byte {
	var inst uint32 = 0
	inst |= 0b100101 << 26
	imm26 := uint32(offset/4) & 0x3FFFFFF
	inst |= imm26
	return uint32ToBytes(inst)
}

// encodeBCond encodes: B.cond offset
func encodeBCond(offset int32, cond uint8) []byte {
	var inst uint32 = 0
	inst |= 0b0101010 << 25
	inst |= 0 << 24
	imm19 := uint32(offset/4) & 0x7FFFF
	inst |= imm19 << 5
	inst |= 0 << 4
	inst |= uint32(cond & 0xF)
	return uint32ToBytes(inst)
}

// encodeRET encodes: RET (return from X30)
func encodeRET() []byte {
	var inst uint32 = 0
	inst |= 0b1101011 << 25
	inst |= 0 << 24
	inst |= 0 << 23
	inst |= 0b10 << 21
	inst |= 0b11111 << 16
	inst |= 0b0000 << 12
	inst |= 0 << 11
	inst |= 0 << 10
	inst |= uint32(30) << 5 // Rn = X30 (LR)
	inst |= 0b00000
	return uint32ToBytes(inst)
}

// encodeSVC encodes: SVC #imm
func encodeSVC(imm uint16) []byte {
	var inst uint32 = 0
	inst |= 0b11010100 << 24
	inst |= 0b000 << 21
	inst |= uint32(imm) << 5
	inst |= 0b00001
	return uint32ToBytes(inst)
}

// encodeLDR64 encodes: LDR Xd, [Xn, #offset]
func encodeLDR64(rd, rn uint8, offset uint16) []byte {
	var inst uint32 = 0
	inst |= 0b11 << 30  // size = 11 (64-bit)
	inst |= 0b111 << 27 // opc
	inst |= 0 << 26     // V = 0 (non-SIMD)
	inst |= 0b01 << 24  // opc
	inst |= 0b01 << 22  // unsigned offset
	scaledOffset := offset / 8
	inst |= uint32(scaledOffset&0xFFF) << 10
	inst |= uint32(rn&0x1F) << 5
	inst |= uint32(rd & 0x1F)
	return uint32ToBytes(inst)
}

// ELF creation helpers

// createTestELF creates a minimal ARM64 ELF with a single code segment.
func createTestELF(path string, loadAddr, entryPoint uint64, code []byte) {
	elfHeader := make([]byte, 64)

	copy(elfHeader[0:4], []byte{0x7f, 'E', 'L', 'F'})
	elfHeader[4] = 2                                     // 64-bit
	elfHeader[5] = 1                                     // little endian
	elfHeader[6] = 1                                     // version
	binary.LittleEndian.PutUint16(elfHeader[16:18], 2)   // executable
	binary.LittleEndian.PutUint16(elfHeader[18:20], 183) // AArch64
	binary.LittleEndian.PutUint32(elfHeader[20:24], 1)   // version
	binary.LittleEndian.PutUint64(elfHeader[24:32], entryPoint)
	binary.LittleEndian.PutUint64(elfHeader[32:40], 64) // phoff
	binary.LittleEndian.PutUint16(elfHeader[52:54], 64) // ehsize
	binary.LittleEndian.PutUint16(elfHeader[54:56], 56) // phentsize
	binary.LittleEndian.PutUint16(elfHeader[56:58], 1)  // phnum

	progHeader := make([]byte, 56)
	binary.LittleEndian.PutUint32(progHeader[0:4], 1)                   // PT_LOAD
	binary.LittleEndian.PutUint32(progHeader[4:8], 0x5)                 // PF_R | PF_X
	binary.LittleEndian.PutUint64(progHeader[8:16], 120)                // offset
	binary.LittleEndian.PutUint64(progHeader[16:24], loadAddr)          // vaddr
	binary.LittleEndian.PutUint64(progHeader[24:32], loadAddr)          // paddr
	binary.LittleEndian.PutUint64(progHeader[32:40], uint64(len(code))) // filesz
	binary.LittleEndian.PutUint64(progHeader[40:48], uint64(len(code))) // memsz
	binary.LittleEndian.PutUint64(progHeader[48:56], 0x1000)            // align

	file, _ := os.Create(path)
	defer file.Close()
	file.Write(elfHeader)
	file.Write(progHeader)
	file.Write(code)
}

// createHelloELF creates an ARM64 ELF with code and data segments.
func createHelloELF(path string, codeAddr, entryPoint uint64, code []byte, dataAddr uint64, data []byte) {
	elfHeader := make([]byte, 64)

	copy(elfHeader[0:4], []byte{0x7f, 'E', 'L', 'F'})
	elfHeader[4] = 2                                     // 64-bit
	elfHeader[5] = 1                                     // little endian
	elfHeader[6] = 1                                     // version
	binary.LittleEndian.PutUint16(elfHeader[16:18], 2)   // executable
	binary.LittleEndian.PutUint16(elfHeader[18:20], 183) // AArch64
	binary.LittleEndian.PutUint32(elfHeader[20:24], 1)   // version
	binary.LittleEndian.PutUint64(elfHeader[24:32], entryPoint)
	binary.LittleEndian.PutUint64(elfHeader[32:40], 64) // phoff
	binary.LittleEndian.PutUint16(elfHeader[52:54], 64) // ehsize
	binary.LittleEndian.PutUint16(elfHeader[54:56], 56) // phentsize
	binary.LittleEndian.PutUint16(elfHeader[56:58], 2)  // phnum (2 segments)

	// Code segment (RX)
	progHeader1 := make([]byte, 56)
	binary.LittleEndian.PutUint32(progHeader1[0:4], 1)                   // PT_LOAD
	binary.LittleEndian.PutUint32(progHeader1[4:8], 0x5)                 // PF_R | PF_X
	binary.LittleEndian.PutUint64(progHeader1[8:16], 64+56*2)            // offset
	binary.LittleEndian.PutUint64(progHeader1[16:24], codeAddr)          // vaddr
	binary.LittleEndian.PutUint64(progHeader1[24:32], codeAddr)          // paddr
	binary.LittleEndian.PutUint64(progHeader1[32:40], uint64(len(code))) // filesz
	binary.LittleEndian.PutUint64(progHeader1[40:48], uint64(len(code))) // memsz
	binary.LittleEndian.PutUint64(progHeader1[48:56], 0x1000)            // align

	// Data segment (R)
	progHeader2 := make([]byte, 56)
	binary.LittleEndian.PutUint32(progHeader2[0:4], 1)                          // PT_LOAD
	binary.LittleEndian.PutUint32(progHeader2[4:8], 0x4)                        // PF_R
	binary.LittleEndian.PutUint64(progHeader2[8:16], 64+56*2+uint64(len(code))) // offset
	binary.LittleEndian.PutUint64(progHeader2[16:24], dataAddr)                 // vaddr
	binary.LittleEndian.PutUint64(progHeader2[24:32], dataAddr)                 // paddr
	binary.LittleEndian.PutUint64(progHeader2[32:40], uint64(len(data)))        // filesz
	binary.LittleEndian.PutUint64(progHeader2[40:48], uint64(len(data)))        // memsz
	binary.LittleEndian.PutUint64(progHeader2[48:56], 0x1000)                   // align

	file, _ := os.Create(path)
	defer file.Close()
	file.Write(elfHeader)
	file.Write(progHeader1)
	file.Write(progHeader2)
	file.Write(code)
	file.Write(data)
}

// createMultiSegmentTestELF creates an ARM64 ELF with code and data segments.
func createMultiSegmentTestELF(path string, codeAddr, entryPoint uint64, code []byte, dataAddr uint64, data []byte) {
	createHelloELF(path, codeAddr, entryPoint, code, dataAddr, data)
}
