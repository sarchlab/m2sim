package loader_test

import (
	"encoding/binary"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sarchlab/m2sim/loader"
)

var _ = Describe("ELF Loader", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "elf-loader-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("Load", func() {
		Context("with a valid ARM64 ELF binary", func() {
			var elfPath string

			BeforeEach(func() {
				elfPath = filepath.Join(tempDir, "test.elf")
				createMinimalARM64ELF(elfPath, 0x400000, 0x400080, []byte{
					// Simple ARM64 code: mov x0, #42; ret
					0x40, 0x05, 0x80, 0xd2, // mov x0, #42
					0xc0, 0x03, 0x5f, 0xd6, // ret
				})
			})

			It("should load without error", func() {
				prog, err := loader.Load(elfPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(prog).NotTo(BeNil())
			})

			It("should extract the correct entry point", func() {
				prog, err := loader.Load(elfPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(prog.EntryPoint).To(Equal(uint64(0x400080)))
			})

			It("should load segments into memory", func() {
				prog, err := loader.Load(elfPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(prog.Segments)).To(BeNumerically(">", 0))
			})

			It("should set up initial stack pointer", func() {
				prog, err := loader.Load(elfPath)
				Expect(err).NotTo(HaveOccurred())
				// Stack should be set to a reasonable high address
				Expect(prog.InitialSP).To(BeNumerically(">", 0x7f0000000000))
			})
		})

		Context("with segment data", func() {
			It("should correctly load segment contents", func() {
				elfPath := filepath.Join(tempDir, "code.elf")
				codeData := []byte{
					0x40, 0x05, 0x80, 0xd2, // mov x0, #42
					0xc0, 0x03, 0x5f, 0xd6, // ret
				}
				createMinimalARM64ELF(elfPath, 0x400000, 0x400000, codeData)

				prog, err := loader.Load(elfPath)
				Expect(err).NotTo(HaveOccurred())

				// Find the segment containing our code
				var foundSegment *loader.Segment
				for i := range prog.Segments {
					if prog.Segments[i].VirtAddr == 0x400000 {
						foundSegment = &prog.Segments[i]
						break
					}
				}
				Expect(foundSegment).NotTo(BeNil())
				Expect(foundSegment.Data).To(HaveLen(len(codeData)))
			})
		})

		Context("with an invalid file", func() {
			It("should return error for non-existent file", func() {
				_, err := loader.Load("/nonexistent/path/to/file.elf")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to open"))
			})

			It("should return error for non-ELF file", func() {
				notElfPath := filepath.Join(tempDir, "not-elf.bin")
				err := os.WriteFile(notElfPath, []byte("not an elf file"), 0644)
				Expect(err).NotTo(HaveOccurred())

				_, err = loader.Load(notElfPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ELF"))
			})

			It("should return error for empty file", func() {
				emptyPath := filepath.Join(tempDir, "empty.elf")
				err := os.WriteFile(emptyPath, []byte{}, 0644)
				Expect(err).NotTo(HaveOccurred())

				_, err = loader.Load(emptyPath)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with non-ARM64 ELF", func() {
			It("should return error for x86-64 ELF", func() {
				elfPath := filepath.Join(tempDir, "x86.elf")
				createMinimalx86ELF(elfPath)

				_, err := loader.Load(elfPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not an ARM64"))
			})
		})

		Context("with 32-bit ELF", func() {
			It("should return error for 32-bit ELF", func() {
				elfPath := filepath.Join(tempDir, "elf32.elf")
				createMinimal32BitELF(elfPath)

				_, err := loader.Load(elfPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not a 64-bit"))
			})
		})
	})

	Describe("Program", func() {
		It("should provide LoadIntoMemory helper", func() {
			elfPath := filepath.Join(tempDir, "test.elf")
			codeData := []byte{0x40, 0x05, 0x80, 0xd2, 0xc0, 0x03, 0x5f, 0xd6}
			createMinimalARM64ELF(elfPath, 0x400000, 0x400000, codeData)

			prog, err := loader.Load(elfPath)
			Expect(err).NotTo(HaveOccurred())

			// Verify segments can be iterated for loading
			totalBytes := uint64(0)
			for _, seg := range prog.Segments {
				totalBytes += seg.MemSize
			}
			Expect(totalBytes).To(BeNumerically(">", 0))
		})
	})

	Describe("Segment", func() {
		It("should have correct virtual address", func() {
			elfPath := filepath.Join(tempDir, "test.elf")
			createMinimalARM64ELF(elfPath, 0x500000, 0x500000, []byte{0x00})

			prog, err := loader.Load(elfPath)
			Expect(err).NotTo(HaveOccurred())

			found := false
			for _, seg := range prog.Segments {
				if seg.VirtAddr == 0x500000 {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should correctly report permissions", func() {
			elfPath := filepath.Join(tempDir, "test.elf")
			createMinimalARM64ELF(elfPath, 0x400000, 0x400000, []byte{0x00})

			prog, err := loader.Load(elfPath)
			Expect(err).NotTo(HaveOccurred())

			// At least one segment should be executable (code)
			hasExecutable := false
			for _, seg := range prog.Segments {
				if seg.Flags&loader.SegmentFlagExecute != 0 {
					hasExecutable = true
					break
				}
			}
			Expect(hasExecutable).To(BeTrue())
		})
	})
})

// createMinimalARM64ELF creates a minimal valid ARM64 ELF64 binary.
func createMinimalARM64ELF(path string, loadAddr, entryPoint uint64, code []byte) {
	// ELF Header (64 bytes)
	elfHeader := make([]byte, 64)

	// Magic number
	copy(elfHeader[0:4], []byte{0x7f, 'E', 'L', 'F'})
	// Class: 64-bit
	elfHeader[4] = 2
	// Data: little endian
	elfHeader[5] = 1
	// Version
	elfHeader[6] = 1
	// OS/ABI
	elfHeader[7] = 0
	// Type: executable
	binary.LittleEndian.PutUint16(elfHeader[16:18], 2)
	// Machine: AArch64
	binary.LittleEndian.PutUint16(elfHeader[18:20], 183)
	// Version
	binary.LittleEndian.PutUint32(elfHeader[20:24], 1)
	// Entry point
	binary.LittleEndian.PutUint64(elfHeader[24:32], entryPoint)
	// Program header offset (right after ELF header)
	binary.LittleEndian.PutUint64(elfHeader[32:40], 64)
	// Section header offset (none)
	binary.LittleEndian.PutUint64(elfHeader[40:48], 0)
	// Flags
	binary.LittleEndian.PutUint32(elfHeader[48:52], 0)
	// ELF header size
	binary.LittleEndian.PutUint16(elfHeader[52:54], 64)
	// Program header entry size
	binary.LittleEndian.PutUint16(elfHeader[54:56], 56)
	// Number of program headers
	binary.LittleEndian.PutUint16(elfHeader[56:58], 1)
	// Section header entry size
	binary.LittleEndian.PutUint16(elfHeader[58:60], 64)
	// Number of section headers
	binary.LittleEndian.PutUint16(elfHeader[60:62], 0)
	// Section name string table index
	binary.LittleEndian.PutUint16(elfHeader[62:64], 0)

	// Program Header (56 bytes) - PT_LOAD
	progHeader := make([]byte, 56)
	// Type: PT_LOAD
	binary.LittleEndian.PutUint32(progHeader[0:4], 1)
	// Flags: PF_X | PF_R (readable + executable)
	binary.LittleEndian.PutUint32(progHeader[4:8], 0x5)
	// Offset in file (after headers)
	binary.LittleEndian.PutUint64(progHeader[8:16], 120)
	// Virtual address
	binary.LittleEndian.PutUint64(progHeader[16:24], loadAddr)
	// Physical address
	binary.LittleEndian.PutUint64(progHeader[24:32], loadAddr)
	// File size
	binary.LittleEndian.PutUint64(progHeader[32:40], uint64(len(code)))
	// Memory size
	binary.LittleEndian.PutUint64(progHeader[40:48], uint64(len(code)))
	// Alignment
	binary.LittleEndian.PutUint64(progHeader[48:56], 0x1000)

	// Write the ELF file
	file, _ := os.Create(path)
	defer file.Close()

	file.Write(elfHeader)
	file.Write(progHeader)
	file.Write(code)
}

// createMinimalx86ELF creates a minimal x86-64 ELF to test rejection.
func createMinimalx86ELF(path string) {
	elfHeader := make([]byte, 64)

	copy(elfHeader[0:4], []byte{0x7f, 'E', 'L', 'F'})
	elfHeader[4] = 2                                    // 64-bit
	elfHeader[5] = 1                                    // little endian
	elfHeader[6] = 1                                    // version
	binary.LittleEndian.PutUint16(elfHeader[16:18], 2)  // executable
	binary.LittleEndian.PutUint16(elfHeader[18:20], 62) // x86-64
	binary.LittleEndian.PutUint32(elfHeader[20:24], 1)  // version
	binary.LittleEndian.PutUint64(elfHeader[24:32], 0)  // entry
	binary.LittleEndian.PutUint64(elfHeader[32:40], 64) // phoff
	binary.LittleEndian.PutUint16(elfHeader[52:54], 64) // ehsize
	binary.LittleEndian.PutUint16(elfHeader[54:56], 56) // phentsize
	binary.LittleEndian.PutUint16(elfHeader[56:58], 0)  // phnum

	file, _ := os.Create(path)
	defer file.Close()
	file.Write(elfHeader)
}

// createMinimal32BitELF creates a minimal 32-bit ELF to test rejection.
func createMinimal32BitELF(path string) {
	elfHeader := make([]byte, 52)

	copy(elfHeader[0:4], []byte{0x7f, 'E', 'L', 'F'})
	elfHeader[4] = 1                                     // 32-bit (ELFCLASS32)
	elfHeader[5] = 1                                     // little endian
	elfHeader[6] = 1                                     // version
	binary.LittleEndian.PutUint16(elfHeader[16:18], 2)   // executable
	binary.LittleEndian.PutUint16(elfHeader[18:20], 183) // ARM64 (won't matter)
	binary.LittleEndian.PutUint32(elfHeader[20:24], 1)   // version

	file, _ := os.Create(path)
	defer file.Close()
	file.Write(elfHeader)
}
