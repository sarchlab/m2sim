// Package cache provides cache hierarchy modeling using Akita cache components.
package cache

import (
	"github.com/sarchlab/m2sim/emu"
)

// MemoryBacking wraps emu.Memory as a BackingStore.
type MemoryBacking struct {
	memory *emu.Memory
}

// NewMemoryBacking creates a new MemoryBacking adapter.
func NewMemoryBacking(memory *emu.Memory) *MemoryBacking {
	return &MemoryBacking{memory: memory}
}

// Read fetches data from the backing memory.
func (m *MemoryBacking) Read(addr uint64, size int) []byte {
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = m.memory.Read8(addr + uint64(i))
	}
	return data
}

// Write stores data to the backing memory.
func (m *MemoryBacking) Write(addr uint64, data []byte) {
	for i, b := range data {
		m.memory.Write8(addr+uint64(i), b)
	}
}

// CacheBacking wraps a Cache as a BackingStore.
// This enables hierarchical cache configurations (e.g., L1 → L2 → Memory).
type CacheBacking struct {
	cache *Cache
}

// NewCacheBacking creates a new CacheBacking adapter.
func NewCacheBacking(cache *Cache) *CacheBacking {
	return &CacheBacking{cache: cache}
}

// Read fetches data from the backing cache.
func (c *CacheBacking) Read(addr uint64, size int) []byte {
	data := make([]byte, size)
	// Read byte by byte to handle sizes properly
	for i := 0; i < size; i++ {
		result := c.cache.Read(addr+uint64(i), 1)
		data[i] = byte(result.Data)
	}
	return data
}

// Write stores data to the backing cache.
func (c *CacheBacking) Write(addr uint64, data []byte) {
	for i, b := range data {
		c.cache.Write(addr+uint64(i), 1, uint64(b))
	}
}

// Cache returns the underlying cache for statistics access.
func (c *CacheBacking) Cache() *Cache {
	return c.cache
}
