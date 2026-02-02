package mem_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/timing/mem"
)

var _ = Describe("Mem", func() {
	It("should have a MemoryController type", func() {
		var m mem.MemoryController
		Expect(m).To(BeZero())
	})
})
