package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/timing/core"
)

var _ = Describe("Core", func() {
	It("should have a Core type", func() {
		var c core.Core
		Expect(c).To(BeZero())
	})
})
