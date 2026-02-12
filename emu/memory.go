// Package emu provides functional ARM64 emulation.
package emu

import "encoding/binary"

const (
	// pageSize defines the size of each memory page (64KB)
	pageSize = 64 * 1024
	// pageMask is used to extract the page offset from an address
	pageMask = pageSize - 1
)

// memoryPage represents a single page of memory
type memoryPage struct {
	data [pageSize]byte
}

// Memory provides a page-based byte-addressable memory model for emulation.
// This optimization replaces the map[uint64]byte approach with a page-based system
// to eliminate map access overhead, targeting the 3.95% CPU usage in Memory.Read32.
type Memory struct {
	pages map[uint64]*memoryPage
}

// NewMemory creates a new memory instance.
func NewMemory() *Memory {
	return &Memory{
		pages: make(map[uint64]*memoryPage),
	}
}

// getOrCreatePage gets an existing page or creates a new one for the given address.
func (m *Memory) getOrCreatePage(addr uint64) *memoryPage {
	pageAddr := addr &^ pageMask
	page, exists := m.pages[pageAddr]
	if !exists {
		page = &memoryPage{}
		m.pages[pageAddr] = page
	}
	return page
}

// getPage gets an existing page for the given address, or nil if it doesn't exist.
func (m *Memory) getPage(addr uint64) *memoryPage {
	pageAddr := addr &^ pageMask
	return m.pages[pageAddr]
}

// Read8 reads a single byte from memory.
func (m *Memory) Read8(addr uint64) byte {
	page := m.getPage(addr)
	if page == nil {
		return 0 // Return 0 for uninitialized memory
	}
	return page.data[addr&pageMask]
}

// Write8 writes a single byte to memory.
func (m *Memory) Write8(addr uint64, value byte) {
	page := m.getOrCreatePage(addr)
	page.data[addr&pageMask] = value
}

// Read16 reads a 16-bit little-endian value from memory.
func (m *Memory) Read16(addr uint64) uint16 {
	var buf [2]byte
	for i := uint64(0); i < 2; i++ {
		buf[i] = m.Read8(addr + i)
	}
	return binary.LittleEndian.Uint16(buf[:])
}

// Write16 writes a 16-bit little-endian value to memory.
func (m *Memory) Write16(addr uint64, value uint16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], value)
	for i := uint64(0); i < 2; i++ {
		m.Write8(addr+i, buf[i])
	}
}

// Read32 reads a 32-bit little-endian value from memory.
func (m *Memory) Read32(addr uint64) uint32 {
	var buf [4]byte
	for i := uint64(0); i < 4; i++ {
		buf[i] = m.Read8(addr + i)
	}
	return binary.LittleEndian.Uint32(buf[:])
}

// Write32 writes a 32-bit little-endian value to memory.
func (m *Memory) Write32(addr uint64, value uint32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	for i := uint64(0); i < 4; i++ {
		m.Write8(addr+i, buf[i])
	}
}

// Read64 reads a 64-bit little-endian value from memory.
func (m *Memory) Read64(addr uint64) uint64 {
	var buf [8]byte
	for i := uint64(0); i < 8; i++ {
		buf[i] = m.Read8(addr + i)
	}
	return binary.LittleEndian.Uint64(buf[:])
}

// Write64 writes a 64-bit little-endian value to memory.
func (m *Memory) Write64(addr uint64, value uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], value)
	for i := uint64(0); i < 8; i++ {
		m.Write8(addr+i, buf[i])
	}
}

// LoadProgram loads a binary program into memory at the specified address.
func (m *Memory) LoadProgram(addr uint64, program []byte) {
	for i, b := range program {
		m.Write8(addr+uint64(i), b)
	}
}
