package emu_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
)

var _ = Describe("Emu", func() {
	It("should have an Emulator type", func() {
		var e emu.Emulator
		Expect(e).To(BeZero())
	})
})
