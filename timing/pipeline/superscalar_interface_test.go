package pipeline

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Superscalar Register Interfaces", func() {
	Describe("SecondaryMEMWBRegister WritebackSlot interface", func() {
		var reg *SecondaryMEMWBRegister

		BeforeEach(func() {
			reg = &SecondaryMEMWBRegister{
				Valid:     true,
				RegWrite:  true,
				Rd:        5,
				ALUResult: 42,
				MemData:   100,
				MemToReg:  true,
			}
		})

		It("should return MemData via GetMemData", func() {
			Expect(reg.GetMemData()).To(Equal(uint64(100)))
		})

		It("should return false for GetIsFused (fusion only in slot 0)", func() {
			Expect(reg.GetIsFused()).To(BeFalse())
		})
	})

	Describe("TertiaryMEMWBRegister WritebackSlot interface", func() {
		var reg *TertiaryMEMWBRegister

		BeforeEach(func() {
			reg = &TertiaryMEMWBRegister{
				Valid:     true,
				RegWrite:  true,
				Rd:        6,
				ALUResult: 123,
				MemData:   200,
				MemToReg:  false,
			}
		})

		It("should return MemData via GetMemData", func() {
			Expect(reg.GetMemData()).To(Equal(uint64(200)))
		})

		It("should return false for GetIsFused", func() {
			Expect(reg.GetIsFused()).To(BeFalse())
		})
	})

	Describe("QuaternaryMEMWBRegister WritebackSlot interface", func() {
		var reg *QuaternaryMEMWBRegister

		BeforeEach(func() {
			reg = &QuaternaryMEMWBRegister{
				Valid:     true,
				RegWrite:  true,
				Rd:        7,
				ALUResult: 456,
				MemData:   300,
				MemToReg:  true,
			}
		})

		It("should return MemData via GetMemData", func() {
			Expect(reg.GetMemData()).To(Equal(uint64(300)))
		})

		It("should return false for GetIsFused", func() {
			Expect(reg.GetIsFused()).To(BeFalse())
		})
	})

	Describe("QuinaryMEMWBRegister WritebackSlot interface", func() {
		var reg *QuinaryMEMWBRegister

		BeforeEach(func() {
			reg = &QuinaryMEMWBRegister{
				Valid:     true,
				RegWrite:  true,
				Rd:        8,
				ALUResult: 789,
				MemData:   400,
				MemToReg:  false,
			}
		})

		It("should return MemData via GetMemData", func() {
			Expect(reg.GetMemData()).To(Equal(uint64(400)))
		})

		It("should return false for GetIsFused", func() {
			Expect(reg.GetIsFused()).To(BeFalse())
		})
	})

	Describe("SenaryMEMWBRegister WritebackSlot interface", func() {
		var reg *SenaryMEMWBRegister

		BeforeEach(func() {
			reg = &SenaryMEMWBRegister{
				Valid:     true,
				RegWrite:  true,
				Rd:        9,
				ALUResult: 111,
				MemData:   500,
				MemToReg:  true,
			}
		})

		It("should return MemData via GetMemData", func() {
			Expect(reg.GetMemData()).To(Equal(uint64(500)))
		})

		It("should return false for GetIsFused", func() {
			Expect(reg.GetIsFused()).To(BeFalse())
		})
	})

	Describe("SecondaryEXMEMRegister ExMemSlot interface", func() {
		var reg *SecondaryEXMEMRegister

		BeforeEach(func() {
			reg = &SecondaryEXMEMRegister{
				Valid:      true,
				RegWrite:   true,
				Rd:         10,
				MemRead:    true,
				MemWrite:   false,
				ALUResult:  222,
				StoreValue: 333,
			}
		})

		It("should return IsValid correctly", func() {
			Expect(reg.IsValid()).To(BeTrue())
		})

		It("should return GetMemRead correctly", func() {
			Expect(reg.GetMemRead()).To(BeTrue())
		})

		It("should return GetMemWrite correctly", func() {
			Expect(reg.GetMemWrite()).To(BeFalse())
		})

		It("should return GetALUResult correctly", func() {
			Expect(reg.GetALUResult()).To(Equal(uint64(222)))
		})

		It("should return GetStoreValue correctly", func() {
			Expect(reg.GetStoreValue()).To(Equal(uint64(333)))
		})
	})

	Describe("TertiaryEXMEMRegister ExMemSlot interface", func() {
		var reg *TertiaryEXMEMRegister

		BeforeEach(func() {
			reg = &TertiaryEXMEMRegister{
				Valid:      true,
				RegWrite:   true,
				Rd:         11,
				MemRead:    false,
				MemWrite:   true,
				ALUResult:  444,
				StoreValue: 555,
			}
		})

		It("should return IsValid correctly", func() {
			Expect(reg.IsValid()).To(BeTrue())
		})

		It("should return GetMemRead correctly", func() {
			Expect(reg.GetMemRead()).To(BeFalse())
		})

		It("should return GetMemWrite correctly", func() {
			Expect(reg.GetMemWrite()).To(BeTrue())
		})

		It("should return GetALUResult correctly", func() {
			Expect(reg.GetALUResult()).To(Equal(uint64(444)))
		})

		It("should return GetStoreValue correctly", func() {
			Expect(reg.GetStoreValue()).To(Equal(uint64(555)))
		})
	})

	Describe("QuaternaryEXMEMRegister ExMemSlot interface", func() {
		var reg *QuaternaryEXMEMRegister

		BeforeEach(func() {
			reg = &QuaternaryEXMEMRegister{
				Valid:      true,
				RegWrite:   false,
				Rd:         12,
				MemRead:    true,
				MemWrite:   false,
				ALUResult:  666,
				StoreValue: 777,
			}
		})

		It("should return IsValid correctly", func() {
			Expect(reg.IsValid()).To(BeTrue())
		})

		It("should return GetMemRead correctly", func() {
			Expect(reg.GetMemRead()).To(BeTrue())
		})

		It("should return GetMemWrite correctly", func() {
			Expect(reg.GetMemWrite()).To(BeFalse())
		})

		It("should return GetALUResult correctly", func() {
			Expect(reg.GetALUResult()).To(Equal(uint64(666)))
		})

		It("should return GetStoreValue correctly", func() {
			Expect(reg.GetStoreValue()).To(Equal(uint64(777)))
		})
	})

	Describe("QuinaryEXMEMRegister ExMemSlot interface", func() {
		var reg *QuinaryEXMEMRegister

		BeforeEach(func() {
			reg = &QuinaryEXMEMRegister{
				Valid:      true,
				RegWrite:   true,
				Rd:         13,
				MemRead:    false,
				MemWrite:   true,
				ALUResult:  888,
				StoreValue: 999,
			}
		})

		It("should return IsValid correctly", func() {
			Expect(reg.IsValid()).To(BeTrue())
		})

		It("should return GetMemRead correctly", func() {
			Expect(reg.GetMemRead()).To(BeFalse())
		})

		It("should return GetMemWrite correctly", func() {
			Expect(reg.GetMemWrite()).To(BeTrue())
		})

		It("should return GetALUResult correctly", func() {
			Expect(reg.GetALUResult()).To(Equal(uint64(888)))
		})

		It("should return GetStoreValue correctly", func() {
			Expect(reg.GetStoreValue()).To(Equal(uint64(999)))
		})
	})

	Describe("SenaryEXMEMRegister ExMemSlot interface", func() {
		var reg *SenaryEXMEMRegister

		BeforeEach(func() {
			reg = &SenaryEXMEMRegister{
				Valid:      true,
				RegWrite:   true,
				Rd:         14,
				MemRead:    true,
				MemWrite:   true,
				ALUResult:  1111,
				StoreValue: 2222,
			}
		})

		It("should return IsValid correctly", func() {
			Expect(reg.IsValid()).To(BeTrue())
		})

		It("should return GetMemRead correctly", func() {
			Expect(reg.GetMemRead()).To(BeTrue())
		})

		It("should return GetMemWrite correctly", func() {
			Expect(reg.GetMemWrite()).To(BeTrue())
		})

		It("should return GetALUResult correctly", func() {
			Expect(reg.GetALUResult()).To(Equal(uint64(1111)))
		})

		It("should return GetStoreValue correctly", func() {
			Expect(reg.GetStoreValue()).To(Equal(uint64(2222)))
		})
	})
})
