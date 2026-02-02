package cache_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/timing/cache"
)

var _ = Describe("Cache", func() {
	It("should have a Cache type", func() {
		var c cache.Cache
		Expect(c).To(BeZero())
	})
})
