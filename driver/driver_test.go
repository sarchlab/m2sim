package driver_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/driver"
)

var _ = Describe("Driver", func() {
	It("should have a Driver type", func() {
		var d driver.Driver
		Expect(d).To(BeZero())
	})
})
