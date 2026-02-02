package emu_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
)

var _ = Describe("Memory", func() {
	var mem *emu.Memory

	BeforeEach(func() {
		mem = emu.NewMemory()
	})

	Describe("NewMemory", func() {
		It("should create a new memory instance", func() {
			Expect(mem).NotTo(BeNil())
		})
	})

	Describe("8-bit operations", func() {
		It("should read and write single bytes", func() {
			mem.Write8(0x1000, 0xAB)
			Expect(mem.Read8(0x1000)).To(Equal(byte(0xAB)))
		})

		It("should return zero for unwritten addresses", func() {
			Expect(mem.Read8(0x2000)).To(Equal(byte(0x00)))
		})

		It("should handle maximum byte value", func() {
			mem.Write8(0x1000, 0xFF)
			Expect(mem.Read8(0x1000)).To(Equal(byte(0xFF)))
		})

		It("should handle address 0", func() {
			mem.Write8(0x0, 0x42)
			Expect(mem.Read8(0x0)).To(Equal(byte(0x42)))
		})

		It("should handle high addresses", func() {
			mem.Write8(0xFFFFFFFFFFFF0000, 0x99)
			Expect(mem.Read8(0xFFFFFFFFFFFF0000)).To(Equal(byte(0x99)))
		})
	})

	Describe("16-bit operations", func() {
		It("should read and write 16-bit values in little-endian", func() {
			mem.Write16(0x1000, 0x1234)
			Expect(mem.Read16(0x1000)).To(Equal(uint16(0x1234)))
		})

		It("should store bytes in little-endian order", func() {
			mem.Write16(0x1000, 0xABCD)
			Expect(mem.Read8(0x1000)).To(Equal(byte(0xCD))) // Low byte first
			Expect(mem.Read8(0x1001)).To(Equal(byte(0xAB))) // High byte second
		})

		It("should reconstruct from individual bytes", func() {
			mem.Write8(0x1000, 0x34) // Low byte
			mem.Write8(0x1001, 0x12) // High byte
			Expect(mem.Read16(0x1000)).To(Equal(uint16(0x1234)))
		})

		It("should return zero for unwritten addresses", func() {
			Expect(mem.Read16(0x3000)).To(Equal(uint16(0x0000)))
		})

		It("should handle maximum 16-bit value", func() {
			mem.Write16(0x1000, 0xFFFF)
			Expect(mem.Read16(0x1000)).To(Equal(uint16(0xFFFF)))
		})
	})

	Describe("32-bit operations", func() {
		It("should read and write 32-bit values in little-endian", func() {
			mem.Write32(0x1000, 0x12345678)
			Expect(mem.Read32(0x1000)).To(Equal(uint32(0x12345678)))
		})

		It("should store bytes in little-endian order", func() {
			mem.Write32(0x1000, 0xDEADBEEF)
			Expect(mem.Read8(0x1000)).To(Equal(byte(0xEF)))
			Expect(mem.Read8(0x1001)).To(Equal(byte(0xBE)))
			Expect(mem.Read8(0x1002)).To(Equal(byte(0xAD)))
			Expect(mem.Read8(0x1003)).To(Equal(byte(0xDE)))
		})

		It("should reconstruct from individual bytes", func() {
			mem.Write8(0x1000, 0x78)
			mem.Write8(0x1001, 0x56)
			mem.Write8(0x1002, 0x34)
			mem.Write8(0x1003, 0x12)
			Expect(mem.Read32(0x1000)).To(Equal(uint32(0x12345678)))
		})

		It("should return zero for unwritten addresses", func() {
			Expect(mem.Read32(0x4000)).To(Equal(uint32(0x00000000)))
		})

		It("should handle maximum 32-bit value", func() {
			mem.Write32(0x1000, 0xFFFFFFFF)
			Expect(mem.Read32(0x1000)).To(Equal(uint32(0xFFFFFFFF)))
		})
	})

	Describe("64-bit operations", func() {
		It("should read and write 64-bit values in little-endian", func() {
			mem.Write64(0x1000, 0x123456789ABCDEF0)
			Expect(mem.Read64(0x1000)).To(Equal(uint64(0x123456789ABCDEF0)))
		})

		It("should store bytes in little-endian order", func() {
			mem.Write64(0x1000, 0x0102030405060708)
			Expect(mem.Read8(0x1000)).To(Equal(byte(0x08)))
			Expect(mem.Read8(0x1001)).To(Equal(byte(0x07)))
			Expect(mem.Read8(0x1002)).To(Equal(byte(0x06)))
			Expect(mem.Read8(0x1003)).To(Equal(byte(0x05)))
			Expect(mem.Read8(0x1004)).To(Equal(byte(0x04)))
			Expect(mem.Read8(0x1005)).To(Equal(byte(0x03)))
			Expect(mem.Read8(0x1006)).To(Equal(byte(0x02)))
			Expect(mem.Read8(0x1007)).To(Equal(byte(0x01)))
		})

		It("should reconstruct from individual bytes", func() {
			mem.Write8(0x1000, 0xF0)
			mem.Write8(0x1001, 0xDE)
			mem.Write8(0x1002, 0xBC)
			mem.Write8(0x1003, 0x9A)
			mem.Write8(0x1004, 0x78)
			mem.Write8(0x1005, 0x56)
			mem.Write8(0x1006, 0x34)
			mem.Write8(0x1007, 0x12)
			Expect(mem.Read64(0x1000)).To(Equal(uint64(0x123456789ABCDEF0)))
		})

		It("should return zero for unwritten addresses", func() {
			Expect(mem.Read64(0x5000)).To(Equal(uint64(0x0000000000000000)))
		})

		It("should handle maximum 64-bit value", func() {
			mem.Write64(0x1000, 0xFFFFFFFFFFFFFFFF)
			Expect(mem.Read64(0x1000)).To(Equal(uint64(0xFFFFFFFFFFFFFFFF)))
		})
	})

	Describe("memory isolation", func() {
		It("should not affect adjacent addresses when writing 8-bit", func() {
			mem.Write8(0x1000, 0xAA)
			mem.Write8(0x1002, 0xBB)
			Expect(mem.Read8(0x1001)).To(Equal(byte(0x00)))
		})

		It("should overwrite previous values", func() {
			mem.Write32(0x1000, 0x12345678)
			mem.Write32(0x1000, 0xDEADBEEF)
			Expect(mem.Read32(0x1000)).To(Equal(uint32(0xDEADBEEF)))
		})

		It("should handle overlapping writes correctly", func() {
			mem.Write32(0x1000, 0xFFFFFFFF)
			mem.Write8(0x1001, 0x00)
			Expect(mem.Read32(0x1000)).To(Equal(uint32(0xFFFF00FF)))
		})
	})

	Describe("LoadProgram", func() {
		It("should load an empty program", func() {
			program := []byte{}
			mem.LoadProgram(0x1000, program)
			Expect(mem.Read8(0x1000)).To(Equal(byte(0x00)))
		})

		It("should load a single byte", func() {
			program := []byte{0xAB}
			mem.LoadProgram(0x1000, program)
			Expect(mem.Read8(0x1000)).To(Equal(byte(0xAB)))
		})

		It("should load multiple bytes", func() {
			program := []byte{0x01, 0x02, 0x03, 0x04}
			mem.LoadProgram(0x1000, program)
			Expect(mem.Read8(0x1000)).To(Equal(byte(0x01)))
			Expect(mem.Read8(0x1001)).To(Equal(byte(0x02)))
			Expect(mem.Read8(0x1002)).To(Equal(byte(0x03)))
			Expect(mem.Read8(0x1003)).To(Equal(byte(0x04)))
		})

		It("should load at arbitrary address", func() {
			program := []byte{0xDE, 0xAD, 0xBE, 0xEF}
			mem.LoadProgram(0x80000, program)
			Expect(mem.Read32(0x80000)).To(Equal(uint32(0xEFBEADDE)))
		})

		It("should allow reading loaded program as instructions", func() {
			// ARM64 NOP instruction: 0xD503201F
			program := []byte{0x1F, 0x20, 0x03, 0xD5}
			mem.LoadProgram(0x100000, program)
			Expect(mem.Read32(0x100000)).To(Equal(uint32(0xD503201F)))
		})

		It("should load at address 0", func() {
			program := []byte{0x42, 0x43}
			mem.LoadProgram(0x0, program)
			Expect(mem.Read8(0x0)).To(Equal(byte(0x42)))
			Expect(mem.Read8(0x1)).To(Equal(byte(0x43)))
		})

		It("should not affect memory outside program range", func() {
			mem.Write8(0x1000, 0xFF)
			mem.Write8(0x1005, 0xFF)
			program := []byte{0x01, 0x02, 0x03, 0x04}
			mem.LoadProgram(0x1001, program)
			Expect(mem.Read8(0x1000)).To(Equal(byte(0xFF)))
			Expect(mem.Read8(0x1005)).To(Equal(byte(0xFF)))
		})
	})

	Describe("cross-width operations", func() {
		It("should read 16-bit value as part of 32-bit value", func() {
			mem.Write32(0x1000, 0xAABBCCDD)
			Expect(mem.Read16(0x1000)).To(Equal(uint16(0xCCDD)))
			Expect(mem.Read16(0x1002)).To(Equal(uint16(0xAABB)))
		})

		It("should read 32-bit value as part of 64-bit value", func() {
			mem.Write64(0x1000, 0x1122334455667788)
			Expect(mem.Read32(0x1000)).To(Equal(uint32(0x55667788)))
			Expect(mem.Read32(0x1004)).To(Equal(uint32(0x11223344)))
		})

		It("should handle partial reads", func() {
			mem.Write64(0x1000, 0xFFEEDDCCBBAA9988)
			// Little-endian: 88 99 AA BB CC DD EE FF
			Expect(mem.Read8(0x1003)).To(Equal(byte(0xBB)))
			Expect(mem.Read16(0x1002)).To(Equal(uint16(0xBBAA)))
		})
	})
})
