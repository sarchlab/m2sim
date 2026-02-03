// Package emu provides functional ARM64 emulation.
package emu

// SIMDRegFile represents the ARM64 SIMD/NEON register file.
// It contains 32 vector registers (V0-V31), each 128 bits wide.
// These registers can be accessed as:
//   - Q registers (128-bit quadword)
//   - D registers (64-bit doubleword, lower half)
//   - S registers (32-bit single, lower quarter)
//   - H registers (16-bit half, lower eighth)
//   - B registers (8-bit byte, lowest byte)
type SIMDRegFile struct {
	// V holds the 32 vector registers.
	// Each register is represented as a pair of uint64 (low, high).
	V [32][2]uint64
}

// NewSIMDRegFile creates a new SIMD register file.
func NewSIMDRegFile() *SIMDRegFile {
	return &SIMDRegFile{}
}

// ReadQ reads a 128-bit Q register as two uint64 values (low, high).
func (s *SIMDRegFile) ReadQ(reg uint8) (low, high uint64) {
	return s.V[reg][0], s.V[reg][1]
}

// WriteQ writes a 128-bit Q register from two uint64 values (low, high).
func (s *SIMDRegFile) WriteQ(reg uint8, low, high uint64) {
	s.V[reg][0] = low
	s.V[reg][1] = high
}

// ReadD reads the lower 64 bits (D register).
func (s *SIMDRegFile) ReadD(reg uint8) uint64 {
	return s.V[reg][0]
}

// WriteD writes the lower 64 bits (D register), zeroing the upper 64 bits.
func (s *SIMDRegFile) WriteD(reg uint8, value uint64) {
	s.V[reg][0] = value
	s.V[reg][1] = 0
}

// ReadS reads the lower 32 bits (S register).
func (s *SIMDRegFile) ReadS(reg uint8) uint32 {
	return uint32(s.V[reg][0])
}

// WriteS writes the lower 32 bits (S register), zeroing upper bits.
func (s *SIMDRegFile) WriteS(reg uint8, value uint32) {
	s.V[reg][0] = uint64(value)
	s.V[reg][1] = 0
}

// ReadLane8 reads an 8-bit lane from a vector register.
// Lane index: 0-15 (0 is lowest byte).
func (s *SIMDRegFile) ReadLane8(reg uint8, lane uint8) uint8 {
	if lane < 8 {
		return uint8(s.V[reg][0] >> (lane * 8))
	}
	return uint8(s.V[reg][1] >> ((lane - 8) * 8))
}

// WriteLane8 writes an 8-bit lane to a vector register.
func (s *SIMDRegFile) WriteLane8(reg uint8, lane uint8, value uint8) {
	mask := uint64(0xFF) << (lane * 8)
	if lane < 8 {
		s.V[reg][0] = (s.V[reg][0] &^ mask) | (uint64(value) << (lane * 8))
	} else {
		lane -= 8
		mask = uint64(0xFF) << (lane * 8)
		s.V[reg][1] = (s.V[reg][1] &^ mask) | (uint64(value) << (lane * 8))
	}
}

// ReadLane16 reads a 16-bit lane from a vector register.
// Lane index: 0-7 (0 is lowest halfword).
func (s *SIMDRegFile) ReadLane16(reg uint8, lane uint8) uint16 {
	if lane < 4 {
		return uint16(s.V[reg][0] >> (lane * 16))
	}
	return uint16(s.V[reg][1] >> ((lane - 4) * 16))
}

// WriteLane16 writes a 16-bit lane to a vector register.
func (s *SIMDRegFile) WriteLane16(reg uint8, lane uint8, value uint16) {
	if lane < 4 {
		mask := uint64(0xFFFF) << (lane * 16)
		s.V[reg][0] = (s.V[reg][0] &^ mask) | (uint64(value) << (lane * 16))
	} else {
		lane -= 4
		mask := uint64(0xFFFF) << (lane * 16)
		s.V[reg][1] = (s.V[reg][1] &^ mask) | (uint64(value) << (lane * 16))
	}
}

// ReadLane32 reads a 32-bit lane from a vector register.
// Lane index: 0-3 (0 is lowest word).
func (s *SIMDRegFile) ReadLane32(reg uint8, lane uint8) uint32 {
	if lane < 2 {
		return uint32(s.V[reg][0] >> (lane * 32))
	}
	return uint32(s.V[reg][1] >> ((lane - 2) * 32))
}

// WriteLane32 writes a 32-bit lane to a vector register.
func (s *SIMDRegFile) WriteLane32(reg uint8, lane uint8, value uint32) {
	if lane < 2 {
		mask := uint64(0xFFFFFFFF) << (lane * 32)
		s.V[reg][0] = (s.V[reg][0] &^ mask) | (uint64(value) << (lane * 32))
	} else {
		lane -= 2
		mask := uint64(0xFFFFFFFF) << (lane * 32)
		s.V[reg][1] = (s.V[reg][1] &^ mask) | (uint64(value) << (lane * 32))
	}
}

// ReadLane64 reads a 64-bit lane from a vector register.
// Lane index: 0-1 (0 is low 64 bits, 1 is high 64 bits).
func (s *SIMDRegFile) ReadLane64(reg uint8, lane uint8) uint64 {
	return s.V[reg][lane]
}

// WriteLane64 writes a 64-bit lane to a vector register.
func (s *SIMDRegFile) WriteLane64(reg uint8, lane uint8, value uint64) {
	s.V[reg][lane] = value
}

// Clear zeros all SIMD registers.
func (s *SIMDRegFile) Clear() {
	for i := range s.V {
		s.V[i][0] = 0
		s.V[i][1] = 0
	}
}
