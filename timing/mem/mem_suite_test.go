package mem_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMem(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mem Suite")
}
