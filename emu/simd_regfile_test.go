package emu_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
)

var _ = Describe("SIMDRegFile", func() {
	var simdRegFile *emu.SIMDRegFile

	BeforeEach(func() {
		simdRegFile = emu.NewSIMDRegFile()
	})

	Describe("Q register (128-bit) operations", func() {
		It("should read and write Q registers", func() {
			simdRegFile.WriteQ(0, 0x1234567890ABCDEF, 0xFEDCBA0987654321)

			low, high := simdRegFile.ReadQ(0)
			Expect(low).To(Equal(uint64(0x1234567890ABCDEF)))
			Expect(high).To(Equal(uint64(0xFEDCBA0987654321)))
		})

		It("should have independent registers", func() {
			simdRegFile.WriteQ(0, 0x1111111111111111, 0x2222222222222222)
			simdRegFile.WriteQ(1, 0x3333333333333333, 0x4444444444444444)

			low0, high0 := simdRegFile.ReadQ(0)
			low1, high1 := simdRegFile.ReadQ(1)

			Expect(low0).To(Equal(uint64(0x1111111111111111)))
			Expect(high0).To(Equal(uint64(0x2222222222222222)))
			Expect(low1).To(Equal(uint64(0x3333333333333333)))
			Expect(high1).To(Equal(uint64(0x4444444444444444)))
		})

		It("should support all 32 registers", func() {
			for i := uint8(0); i < 32; i++ {
				simdRegFile.WriteQ(i, uint64(i), uint64(i+100))
			}

			for i := uint8(0); i < 32; i++ {
				low, high := simdRegFile.ReadQ(i)
				Expect(low).To(Equal(uint64(i)))
				Expect(high).To(Equal(uint64(i + 100)))
			}
		})
	})

	Describe("D register (64-bit) operations", func() {
		It("should read and write D registers", func() {
			simdRegFile.WriteD(0, 0x1234567890ABCDEF)

			Expect(simdRegFile.ReadD(0)).To(Equal(uint64(0x1234567890ABCDEF)))
		})

		It("should zero upper bits when writing D register", func() {
			simdRegFile.WriteQ(0, 0x1111111111111111, 0x2222222222222222)
			simdRegFile.WriteD(0, 0xAAAAAAAAAAAAAAAA)

			low, high := simdRegFile.ReadQ(0)
			Expect(low).To(Equal(uint64(0xAAAAAAAAAAAAAAAA)))
			Expect(high).To(Equal(uint64(0))) // Should be zeroed
		})
	})

	Describe("S register (32-bit) operations", func() {
		It("should read and write S registers", func() {
			simdRegFile.WriteS(0, 0x12345678)

			Expect(simdRegFile.ReadS(0)).To(Equal(uint32(0x12345678)))
		})

		It("should zero upper bits when writing S register", func() {
			simdRegFile.WriteQ(0, 0x1111111111111111, 0x2222222222222222)
			simdRegFile.WriteS(0, 0xAABBCCDD)

			low, high := simdRegFile.ReadQ(0)
			Expect(low).To(Equal(uint64(0xAABBCCDD)))
			Expect(high).To(Equal(uint64(0)))
		})
	})

	Describe("Lane operations", func() {
		Context("8-bit lanes", func() {
			It("should read and write 8-bit lanes in low half", func() {
				simdRegFile.WriteQ(0, 0x0706050403020100, 0)

				Expect(simdRegFile.ReadLane8(0, 0)).To(Equal(uint8(0x00)))
				Expect(simdRegFile.ReadLane8(0, 1)).To(Equal(uint8(0x01)))
				Expect(simdRegFile.ReadLane8(0, 7)).To(Equal(uint8(0x07)))
			})

			It("should read and write 8-bit lanes in high half", func() {
				simdRegFile.WriteQ(0, 0, 0x0F0E0D0C0B0A0908)

				Expect(simdRegFile.ReadLane8(0, 8)).To(Equal(uint8(0x08)))
				Expect(simdRegFile.ReadLane8(0, 15)).To(Equal(uint8(0x0F)))
			})

			It("should write individual 8-bit lanes", func() {
				simdRegFile.WriteLane8(0, 0, 0xAA)
				simdRegFile.WriteLane8(0, 8, 0xBB)

				Expect(simdRegFile.ReadLane8(0, 0)).To(Equal(uint8(0xAA)))
				Expect(simdRegFile.ReadLane8(0, 8)).To(Equal(uint8(0xBB)))
			})
		})

		Context("16-bit lanes", func() {
			It("should read and write 16-bit lanes", func() {
				simdRegFile.WriteQ(0, 0x0003000200010000, 0x0007000600050004)

				Expect(simdRegFile.ReadLane16(0, 0)).To(Equal(uint16(0x0000)))
				Expect(simdRegFile.ReadLane16(0, 1)).To(Equal(uint16(0x0001)))
				Expect(simdRegFile.ReadLane16(0, 4)).To(Equal(uint16(0x0004)))
				Expect(simdRegFile.ReadLane16(0, 7)).To(Equal(uint16(0x0007)))
			})

			It("should write individual 16-bit lanes", func() {
				simdRegFile.WriteLane16(0, 0, 0x1234)
				simdRegFile.WriteLane16(0, 4, 0x5678)

				Expect(simdRegFile.ReadLane16(0, 0)).To(Equal(uint16(0x1234)))
				Expect(simdRegFile.ReadLane16(0, 4)).To(Equal(uint16(0x5678)))
			})
		})

		Context("32-bit lanes", func() {
			It("should read and write 32-bit lanes", func() {
				simdRegFile.WriteQ(0, 0x0000000100000000, 0x0000000300000002)

				Expect(simdRegFile.ReadLane32(0, 0)).To(Equal(uint32(0x00000000)))
				Expect(simdRegFile.ReadLane32(0, 1)).To(Equal(uint32(0x00000001)))
				Expect(simdRegFile.ReadLane32(0, 2)).To(Equal(uint32(0x00000002)))
				Expect(simdRegFile.ReadLane32(0, 3)).To(Equal(uint32(0x00000003)))
			})

			It("should write individual 32-bit lanes", func() {
				simdRegFile.WriteLane32(0, 0, 0x12345678)
				simdRegFile.WriteLane32(0, 2, 0xABCDEF00)

				Expect(simdRegFile.ReadLane32(0, 0)).To(Equal(uint32(0x12345678)))
				Expect(simdRegFile.ReadLane32(0, 2)).To(Equal(uint32(0xABCDEF00)))
			})
		})

		Context("64-bit lanes", func() {
			It("should read and write 64-bit lanes", func() {
				simdRegFile.WriteQ(0, 0x1111111111111111, 0x2222222222222222)

				Expect(simdRegFile.ReadLane64(0, 0)).To(Equal(uint64(0x1111111111111111)))
				Expect(simdRegFile.ReadLane64(0, 1)).To(Equal(uint64(0x2222222222222222)))
			})

			It("should write individual 64-bit lanes", func() {
				simdRegFile.WriteLane64(0, 0, 0xAAAAAAAAAAAAAAAA)
				simdRegFile.WriteLane64(0, 1, 0xBBBBBBBBBBBBBBBB)

				Expect(simdRegFile.ReadLane64(0, 0)).To(Equal(uint64(0xAAAAAAAAAAAAAAAA)))
				Expect(simdRegFile.ReadLane64(0, 1)).To(Equal(uint64(0xBBBBBBBBBBBBBBBB)))
			})
		})
	})

	Describe("Clear", func() {
		It("should zero all registers", func() {
			for i := uint8(0); i < 32; i++ {
				simdRegFile.WriteQ(i, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF)
			}

			simdRegFile.Clear()

			for i := uint8(0); i < 32; i++ {
				low, high := simdRegFile.ReadQ(i)
				Expect(low).To(Equal(uint64(0)))
				Expect(high).To(Equal(uint64(0)))
			}
		})
	})
})
