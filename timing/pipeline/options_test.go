package pipeline_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
	"github.com/sarchlab/m2sim/timing/pipeline"
)

var _ = Describe("Pipeline Options", func() {
	var (
		regFile *emu.RegFile
		memory  *emu.Memory
	)

	BeforeEach(func() {
		regFile = &emu.RegFile{}
		memory = emu.NewMemory()
		regFile.WriteReg(8, 93) // exit syscall
	})

	Describe("WithSyscallHandler", func() {
		It("should use custom syscall handler", func() {
			customHandlerCalled := false
			customHandler := NewMockSyscallHandler(regFile, memory, func() emu.SyscallResult {
				customHandlerCalled = true
				return emu.SyscallResult{Exited: true, ExitCode: 0}
			})

			pipe := pipeline.NewPipeline(regFile, memory,
				pipeline.WithSyscallHandler(customHandler))

			// SVC #0 instruction
			memory.Write32(0x1000, 0xD4000001)
			pipe.SetPC(0x1000)
			pipe.Run()

			Expect(customHandlerCalled).To(BeTrue())
		})

		It("should allow custom syscall return values", func() {
			customHandler := NewMockSyscallHandler(regFile, memory, func() emu.SyscallResult {
				regFile.WriteReg(0, 42) // Set return value in X0
				return emu.SyscallResult{Exited: true, ExitCode: 99}
			})

			pipe := pipeline.NewPipeline(regFile, memory,
				pipeline.WithSyscallHandler(customHandler))

			memory.Write32(0x1000, 0xD4000001) // SVC #0
			pipe.SetPC(0x1000)
			pipe.Run()

			Expect(pipe.ExitCode()).To(Equal(int64(99)))
			Expect(regFile.ReadReg(0)).To(Equal(uint64(42)))
		})
	})

	Describe("WithBranchPredictorConfig", func() {
		It("should create pipeline with custom branch predictor config", func() {
			config := pipeline.BranchPredictorConfig{
				BHTSize:             512,
				BTBSize:             256,
				GlobalHistoryLength: 8,
				UseTournament:       true,
			}

			pipe := pipeline.NewPipeline(regFile, memory,
				pipeline.WithBranchPredictorConfig(config))

			Expect(pipe).NotTo(BeNil())

			// Execute a simple program to verify pipeline works
			memory.Write32(0x1000, 0x910029E0) // ADD X0, XZR, #10
			memory.Write32(0x1004, 0xD4000001) // SVC #0
			pipe.SetPC(0x1000)
			pipe.Run()

			Expect(regFile.ReadReg(0)).To(Equal(uint64(10)))
		})

		It("should use custom config for branch prediction", func() {
			// Small BTB to test non-default config
			config := pipeline.BranchPredictorConfig{
				BHTSize:             16,
				BTBSize:             4,
				GlobalHistoryLength: 4,
				UseTournament:       false,
			}

			pipe := pipeline.NewPipeline(regFile, memory,
				pipeline.WithBranchPredictorConfig(config))

			// Simple test: ADD X0, XZR, #42; SVC #0
			memory.Write32(0x1000, 0x9100ABE0) // ADD X0, XZR, #42
			memory.Write32(0x1004, 0xD4000001) // SVC #0

			pipe.SetPC(0x1000)
			pipe.Run()

			// Verify the pipeline executed correctly with custom config
			Expect(regFile.ReadReg(0)).To(Equal(uint64(42)))
		})
	})

	Describe("Superscalar Configurations", func() {
		Context("OctupleIssue", func() {
			It("should create 8-wide superscalar pipeline", func() {
				pipe := pipeline.NewPipeline(regFile, memory,
					pipeline.WithOctupleIssue())

				Expect(pipe).NotTo(BeNil())
			})

			It("should execute 8 independent instructions efficiently", func() {
				pipe := pipeline.NewPipeline(regFile, memory,
					pipeline.WithOctupleIssue())

				// 8 independent ADD instructions
				memory.Write32(0x1000, 0x910029E0) // ADD X0, XZR, #10
				memory.Write32(0x1004, 0x910053E1) // ADD X1, XZR, #20
				memory.Write32(0x1008, 0x91007BE2) // ADD X2, XZR, #30
				memory.Write32(0x100C, 0x9100A3E3) // ADD X3, XZR, #40
				memory.Write32(0x1010, 0x9100CBE4) // ADD X4, XZR, #50
				memory.Write32(0x1014, 0x9100F3E5) // ADD X5, XZR, #60
				memory.Write32(0x1018, 0x91011BE6) // ADD X6, XZR, #70
				memory.Write32(0x101C, 0x910143E7) // ADD X7, XZR, #80
				memory.Write32(0x1020, 0xD4000001) // SVC #0

				pipe.SetPC(0x1000)
				pipe.Run()

				octuCycles := pipe.Stats().Cycles

				// Verify results
				Expect(regFile.ReadReg(0)).To(Equal(uint64(10)))
				Expect(regFile.ReadReg(1)).To(Equal(uint64(20)))
				Expect(regFile.ReadReg(2)).To(Equal(uint64(30)))
				Expect(regFile.ReadReg(3)).To(Equal(uint64(40)))
				Expect(regFile.ReadReg(4)).To(Equal(uint64(50)))
				Expect(regFile.ReadReg(5)).To(Equal(uint64(60)))
				Expect(regFile.ReadReg(6)).To(Equal(uint64(70)))
				Expect(regFile.ReadReg(7)).To(Equal(uint64(80)))

				// Compare with single-issue
				regFile2 := &emu.RegFile{}
				regFile2.WriteReg(8, 93)
				singlePipe := pipeline.NewPipeline(regFile2, memory)
				singlePipe.SetPC(0x1000)
				singlePipe.Run()
				singleCycles := singlePipe.Stats().Cycles

				// Octuple should be faster than single
				Expect(octuCycles).To(BeNumerically("<", singleCycles))
			})

			It("should handle dependencies correctly with 8-wide", func() {
				pipe := pipeline.NewPipeline(regFile, memory,
					pipeline.WithOctupleIssue())

				// Chain of dependencies: each instruction depends on previous
				memory.Write32(0x1000, 0x910029E0) // ADD X0, XZR, #10
				memory.Write32(0x1004, 0x91000400) // ADD X0, X0, #1
				memory.Write32(0x1008, 0x91000400) // ADD X0, X0, #1
				memory.Write32(0x100C, 0x91000400) // ADD X0, X0, #1
				memory.Write32(0x1010, 0xD4000001) // SVC #0

				pipe.SetPC(0x1000)
				pipe.Run()

				// Should be 10 + 1 + 1 + 1 = 13
				Expect(regFile.ReadReg(0)).To(Equal(uint64(13)))
			})
		})

		Context("OctupleIssueConfig", func() {
			It("should return config with issue width 8", func() {
				config := pipeline.OctupleIssueConfig()
				Expect(config.IssueWidth).To(Equal(8))
			})
		})
	})
})

// MockSyscallHandler implements a custom syscall handler for testing.
type MockSyscallHandler struct {
	regFile  *emu.RegFile
	memory   *emu.Memory
	onHandle func() emu.SyscallResult
}

// NewMockSyscallHandler creates a new mock syscall handler.
func NewMockSyscallHandler(rf *emu.RegFile, mem *emu.Memory, handler func() emu.SyscallResult) *MockSyscallHandler {
	return &MockSyscallHandler{
		regFile:  rf,
		memory:   mem,
		onHandle: handler,
	}
}

func (h *MockSyscallHandler) Handle() emu.SyscallResult {
	if h.onHandle != nil {
		return h.onHandle()
	}
	return emu.SyscallResult{Exited: true, ExitCode: 0}
}
