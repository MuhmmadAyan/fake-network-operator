package ibkubernetes

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// GUIDAllocator manages the allocation and release of GUIDs.
type GUIDAllocator struct {
	mu        sync.Mutex
	allocated map[string]string
	available []string
}

// NewGUIDAllocator creates a new GUIDAllocator with a pool generated between start and end.
func NewGUIDAllocator(start, end string) *GUIDAllocator {
	pool := generateGUIDPool(start, end)
	return &GUIDAllocator{
		allocated: make(map[string]string),
		available: pool,
	}
}

// Allocate assigns a GUID to the specified podKey.
func (g *GUIDAllocator) Allocate(podKey string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if guid, exists := g.allocated[podKey]; exists {
		return guid, nil
	}

	if len(g.available) == 0 {
		return "", errors.New("no available GUIDs")
	}

	guid := g.available[0]
	g.available = g.available[1:]
	g.allocated[podKey] = guid

	return guid, nil
}

// Release returns the GUID associated with podKey back to the pool.
func (g *GUIDAllocator) Release(podKey string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if guid, exists := g.allocated[podKey]; exists {
		g.available = append(g.available, guid)
		delete(g.allocated, podKey)
		return nil
	}

	return fmt.Errorf("GUID for podKey %s not found", podKey)
}

func generateGUIDPool(start, end string) []string {
	// Parse colon-separated GUIDs to integer range if possible,
	// or fallback to a default list if parsing fails.
	startVal, err1 := guidToUint64(start)
	endVal, err2 := guidToUint64(end)

	var pool []string
	if err1 == nil && err2 == nil && startVal <= endVal {
		// Limit to max 1000 generated GUIDs to avoid OOM
		count := 0
		for i := startVal; i <= endVal && count < 1000; i++ {
			pool = append(pool, uint64ToGUID(i))
			count++
		}
	} else {
		// Fallback static pool
		for i := 1; i <= 256; i++ {
			pool = append(pool, fmt.Sprintf("02:00:00:00:00:00:00:%02x", i))
		}
	}
	return pool
}

func guidToUint64(guid string) (uint64, error) {
	clean := strings.ReplaceAll(guid, ":", "")
	return strconv.ParseUint(clean, 16, 64)
}

func uint64ToGUID(val uint64) string {
	hexStr := fmt.Sprintf("%016x", val)
	var parts []string
	for i := 0; i < 16; i += 2 {
		parts = append(parts, hexStr[i:i+2])
	}
	return strings.Join(parts, ":")
}
