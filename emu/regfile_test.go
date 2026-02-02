package emu_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
)

var _ = Describe("RegFile", func() {
	var regFile *emu.RegFile

	BeforeEach(func() {
		regFile = &emu.RegFile{}
	})

	Describe("64-bit register access", func() {
		It("should read and write general-purpose registers", func() {
			regFile.WriteReg(0, 0x1234567890ABCDEF)
			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x1234567890ABCDEF)))

			regFile.WriteReg(30, 42)
			Expect(regFile.ReadReg(30)).To(Equal(uint64(42)))
		})

		It("should always read zero from XZR (X31)", func() {
			Expect(regFile.ReadReg(31)).To(Equal(uint64(0)))
		})

		It("should ignore writes to XZR (X31)", func() {
			regFile.WriteReg(31, 0xFFFFFFFFFFFFFFFF)
			Expect(regFile.ReadReg(31)).To(Equal(uint64(0)))
		})
	})

	Describe("32-bit register access", func() {
		It("should read lower 32 bits", func() {
			regFile.WriteReg(0, 0xFFFFFFFF12345678)
			Expect(regFile.ReadReg32(0)).To(Equal(uint32(0x12345678)))
		})

		It("should zero-extend on 32-bit write", func() {
			regFile.WriteReg(0, 0xFFFFFFFFFFFFFFFF)
			regFile.WriteReg32(0, 0x12345678)
			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x12345678)))
		})

		It("should handle XZR for 32-bit operations", func() {
			Expect(regFile.ReadReg32(31)).To(Equal(uint32(0)))
			regFile.WriteReg32(31, 0xFFFFFFFF)
			Expect(regFile.ReadReg32(31)).To(Equal(uint32(0)))
		})
	})

	Describe("PSTATE flags", func() {
		It("should initialize all flags to false", func() {
			Expect(regFile.PSTATE.N).To(BeFalse())
			Expect(regFile.PSTATE.Z).To(BeFalse())
			Expect(regFile.PSTATE.C).To(BeFalse())
			Expect(regFile.PSTATE.V).To(BeFalse())
		})

		It("should allow setting and clearing flags", func() {
			regFile.PSTATE.N = true
			regFile.PSTATE.Z = true
			regFile.PSTATE.C = true
			regFile.PSTATE.V = true

			Expect(regFile.PSTATE.N).To(BeTrue())
			Expect(regFile.PSTATE.Z).To(BeTrue())
			Expect(regFile.PSTATE.C).To(BeTrue())
			Expect(regFile.PSTATE.V).To(BeTrue())
		})
	})

	Describe("SP and PC", func() {
		It("should allow reading and writing SP", func() {
			regFile.SP = 0x7FFFFF000000
			Expect(regFile.SP).To(Equal(uint64(0x7FFFFF000000)))
		})

		It("should allow reading and writing PC", func() {
			regFile.PC = 0x100000
			Expect(regFile.PC).To(Equal(uint64(0x100000)))
		})
	})
})
