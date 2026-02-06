// Package emu provides functional ARM64 emulation.
package emu_test

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
)

var _ = Describe("Syscall Handler", func() {
	var (
		regFile *emu.RegFile
		memory  *emu.Memory
		stdout  *bytes.Buffer
		stderr  *bytes.Buffer
		handler *emu.DefaultSyscallHandler
	)

	BeforeEach(func() {
		regFile = &emu.RegFile{}
		memory = emu.NewMemory()
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)
		handler = emu.NewDefaultSyscallHandler(regFile, memory, stdout, stderr)
	})

	Describe("Unknown syscall", func() {
		It("should return ENOSYS for unknown syscall numbers", func() {
			// Set X8 to an unknown syscall number (e.g., 999)
			regFile.WriteReg(8, 999)

			result := handler.Handle()

			// Should not exit
			Expect(result.Exited).To(BeFalse())

			// X0 should contain -ENOSYS (38) as two's complement
			x0 := regFile.ReadReg(0)
			var enosys int64 = 38
			expectedError := uint64(-enosys) // -ENOSYS
			Expect(x0).To(Equal(expectedError))
		})

		It("should handle syscall 0 as unknown", func() {
			// Set X8 to syscall 0 (not implemented)
			regFile.WriteReg(8, 0)

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			// X0 should be -ENOSYS
			x0 := regFile.ReadReg(0)
			var enosys int64 = 38
			expectedError := uint64(-enosys)
			Expect(x0).To(Equal(expectedError))
		})
	})

	Describe("Write syscall with bad fd", func() {
		It("should return EBADF for invalid file descriptor", func() {
			// Set up write syscall with invalid fd
			regFile.WriteReg(8, 64) // SyscallWrite
			regFile.WriteReg(0, 42) // Invalid fd (not 1 or 2)
			regFile.WriteReg(1, 0)  // buf pointer
			regFile.WriteReg(2, 5)  // count

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			// X0 should contain -EBADF (9)
			x0 := regFile.ReadReg(0)
			var ebadf int64 = 9
			expectedError := uint64(-ebadf) // -EBADF
			Expect(x0).To(Equal(expectedError))
		})
	})

	Describe("Exit syscall", func() {
		It("should exit with specified code", func() {
			regFile.WriteReg(8, 93) // SyscallExit
			regFile.WriteReg(0, 42) // Exit code

			result := handler.Handle()

			Expect(result.Exited).To(BeTrue())
			Expect(result.ExitCode).To(Equal(int64(42)))
		})

		It("should handle zero exit code", func() {
			regFile.WriteReg(8, 93) // SyscallExit
			regFile.WriteReg(0, 0)  // Exit code 0

			result := handler.Handle()

			Expect(result.Exited).To(BeTrue())
			Expect(result.ExitCode).To(Equal(int64(0)))
		})
	})

	Describe("Write syscall to stdout", func() {
		It("should write buffer to stdout", func() {
			// Store "hello" in memory
			memory.Write8(0x1000, 'h')
			memory.Write8(0x1001, 'e')
			memory.Write8(0x1002, 'l')
			memory.Write8(0x1003, 'l')
			memory.Write8(0x1004, 'o')

			// Set up write syscall
			regFile.WriteReg(8, 64)     // SyscallWrite
			regFile.WriteReg(0, 1)      // stdout
			regFile.WriteReg(1, 0x1000) // buf pointer
			regFile.WriteReg(2, 5)      // count

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			Expect(stdout.String()).To(Equal("hello"))
			// X0 should contain bytes written
			Expect(regFile.ReadReg(0)).To(Equal(uint64(5)))
		})
	})

	Describe("Write syscall to stderr", func() {
		It("should write buffer to stderr", func() {
			// Store "err" in memory
			memory.Write8(0x2000, 'e')
			memory.Write8(0x2001, 'r')
			memory.Write8(0x2002, 'r')

			// Set up write syscall
			regFile.WriteReg(8, 64)     // SyscallWrite
			regFile.WriteReg(0, 2)      // stderr
			regFile.WriteReg(1, 0x2000) // buf pointer
			regFile.WriteReg(2, 3)      // count

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			Expect(stderr.String()).To(Equal("err"))
			Expect(regFile.ReadReg(0)).To(Equal(uint64(3)))
		})
	})

	Describe("Read syscall from stdin", func() {
		It("should read buffer from stdin", func() {
			stdin := strings.NewReader("hello")
			handler.SetStdin(stdin)

			// Set up read syscall
			regFile.WriteReg(8, 63)     // SyscallRead
			regFile.WriteReg(0, 0)      // stdin
			regFile.WriteReg(1, 0x1000) // buf pointer
			regFile.WriteReg(2, 5)      // count

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			// X0 should contain bytes read
			Expect(regFile.ReadReg(0)).To(Equal(uint64(5)))
			// Verify memory contains "hello"
			Expect(memory.Read8(0x1000)).To(Equal(byte('h')))
			Expect(memory.Read8(0x1001)).To(Equal(byte('e')))
			Expect(memory.Read8(0x1002)).To(Equal(byte('l')))
			Expect(memory.Read8(0x1003)).To(Equal(byte('l')))
			Expect(memory.Read8(0x1004)).To(Equal(byte('o')))
		})

		It("should handle partial reads", func() {
			stdin := strings.NewReader("hi")
			handler.SetStdin(stdin)

			// Request more bytes than available
			regFile.WriteReg(8, 63)     // SyscallRead
			regFile.WriteReg(0, 0)      // stdin
			regFile.WriteReg(1, 0x1000) // buf pointer
			regFile.WriteReg(2, 100)    // count (request 100 bytes)

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			// X0 should contain actual bytes read (2)
			Expect(regFile.ReadReg(0)).To(Equal(uint64(2)))
			// Verify memory contains "hi"
			Expect(memory.Read8(0x1000)).To(Equal(byte('h')))
			Expect(memory.Read8(0x1001)).To(Equal(byte('i')))
		})

		It("should return 0 when stdin is nil (EOF)", func() {
			// No stdin configured (nil by default)
			regFile.WriteReg(8, 63)     // SyscallRead
			regFile.WriteReg(0, 0)      // stdin
			regFile.WriteReg(1, 0x1000) // buf pointer
			regFile.WriteReg(2, 10)     // count

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			// X0 should be 0 (EOF)
			Expect(regFile.ReadReg(0)).To(Equal(uint64(0)))
		})

		It("should return 0 on EOF from exhausted reader", func() {
			stdin := strings.NewReader("")
			handler.SetStdin(stdin)

			regFile.WriteReg(8, 63)     // SyscallRead
			regFile.WriteReg(0, 0)      // stdin
			regFile.WriteReg(1, 0x1000) // buf pointer
			regFile.WriteReg(2, 10)     // count

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			// X0 should be 0 (EOF)
			Expect(regFile.ReadReg(0)).To(Equal(uint64(0)))
		})

		It("should handle zero count read", func() {
			stdin := strings.NewReader("hello")
			handler.SetStdin(stdin)

			regFile.WriteReg(8, 63)     // SyscallRead
			regFile.WriteReg(0, 0)      // stdin
			regFile.WriteReg(1, 0x1000) // buf pointer
			regFile.WriteReg(2, 0)      // count = 0

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			// X0 should be 0 (no bytes read)
			Expect(regFile.ReadReg(0)).To(Equal(uint64(0)))
		})
	})

	Describe("Read syscall with bad fd", func() {
		It("should return EBADF for invalid file descriptor", func() {
			// Set up read syscall with invalid fd
			regFile.WriteReg(8, 63) // SyscallRead
			regFile.WriteReg(0, 42) // Invalid fd (not 0)
			regFile.WriteReg(1, 0)  // buf pointer
			regFile.WriteReg(2, 5)  // count

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			// X0 should contain -EBADF (9)
			x0 := regFile.ReadReg(0)
			var ebadf int64 = 9
			expectedError := uint64(-ebadf) // -EBADF
			Expect(x0).To(Equal(expectedError))
		})

		It("should return EBADF for stdout fd on read", func() {
			regFile.WriteReg(8, 63) // SyscallRead
			regFile.WriteReg(0, 1)  // stdout (invalid for read)
			regFile.WriteReg(1, 0)  // buf pointer
			regFile.WriteReg(2, 5)  // count

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			x0 := regFile.ReadReg(0)
			var ebadf int64 = 9
			expectedError := uint64(-ebadf)
			Expect(x0).To(Equal(expectedError))
		})

		It("should return EBADF for stderr fd on read", func() {
			regFile.WriteReg(8, 63) // SyscallRead
			regFile.WriteReg(0, 2)  // stderr (invalid for read)
			regFile.WriteReg(1, 0)  // buf pointer
			regFile.WriteReg(2, 5)  // count

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			x0 := regFile.ReadReg(0)
			var ebadf int64 = 9
			expectedError := uint64(-ebadf)
			Expect(x0).To(Equal(expectedError))
		})
	})
})
