package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
	"github.com/sarchlab/m2sim/timing/core"
)

var _ = Describe("Core", func() {
	var (
		regFile *emu.RegFile
		memory  *emu.Memory
		c       *core.Core
	)

	BeforeEach(func() {
		regFile = &emu.RegFile{}
		memory = emu.NewMemory()
		c = core.NewCore(regFile, memory)
	})

	It("should create a core with pipeline", func() {
		Expect(c).NotTo(BeNil())
		Expect(c.Pipeline).NotTo(BeNil())
	})

	It("should set and get PC", func() {
		c.SetPC(0x1000)
		Expect(c.Pipeline.PC()).To(Equal(uint64(0x1000)))
	})

	It("should not be halted initially", func() {
		Expect(c.Halted()).To(BeFalse())
	})

	It("should execute instructions through tick", func() {
		// ADD X1, XZR, #42
		memory.Write32(0x1000, 0x9100A821)
		// NOP instructions to flush pipeline.
		memory.Write32(0x1004, 0xD503201F)
		memory.Write32(0x1008, 0xD503201F)
		memory.Write32(0x100C, 0xD503201F)
		memory.Write32(0x1010, 0xD503201F)

		c.SetPC(0x1000)

		for i := 0; i < 10; i++ {
			c.Tick()
		}

		Expect(regFile.X[1]).To(Equal(uint64(42)))
	})

	It("should return stats", func() {
		memory.Write32(0x1000, 0x9100A821) // ADD X1, XZR, #42
		memory.Write32(0x1004, 0xD503201F) // NOP

		c.SetPC(0x1000)
		c.Tick()
		c.Tick()

		stats := c.Stats()
		Expect(stats.Cycles).To(Equal(uint64(2)))
	})
})
