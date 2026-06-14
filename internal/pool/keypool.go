// SPDX-License-Identifier: MIT
// Purpose: Rotate a pool of API keys with basic rate-limit handling.
// Docs: internal/pool/keypool.doc.md
package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// KeyPool manages a set of API keys with round-robin access.
type KeyPool struct {
	keys   []string
	offset uint64
	mu     sync.RWMutex

	// ban tracks keys that returned 429/403 until a backoff time.
	banned map[string]time.Time
}

// New creates a pool from the given keys, removing empty entries.
func New(keys []string) *KeyPool {
	var cleaned []string
	for _, k := range keys {
		if k != "" {
			cleaned = append(cleaned, k)
		}
	}
	return &KeyPool{
		keys:   cleaned,
		banned: make(map[string]time.Time),
	}
}

// IsEmpty returns true if no keys are available.
func (p *KeyPool) IsEmpty() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.keys) == 0
}

// Next returns the next available key, skipping banned ones.
func (p *KeyPool) Next() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	// Clean expired bans.
	for k, t := range p.banned {
		if now.After(t) {
			delete(p.banned, k)
		}
	}

	if len(p.keys) == 0 {
		return "", errors.New("no keys available")
	}

	// The modulo guarantees the value is < len(p.keys), so the int conversion
	// is always within Go's int range for any realistic key pool.
	offset := atomic.AddUint64(&p.offset, 1)
	startIdx := int(offset % uint64(len(p.keys))) // #nosec G115
	for i := 0; i < len(p.keys); i++ {
		idx := (startIdx + i) % len(p.keys)
		key := p.keys[idx]
		if _, banned := p.banned[key]; !banned {
			return key, nil
		}
	}

	return "", errors.New("all keys temporarily banned")
}

// Ban marks a key as unavailable for the given duration.
func (p *KeyPool) Ban(key string, d time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.banned[key] = time.Now().Add(d)
}

// Count returns the number of active (non-banned) keys.
func (p *KeyPool) Count() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for k, t := range p.banned {
		if now.After(t) {
			delete(p.banned, k)
		}
	}

	active := 0
	for _, k := range p.keys {
		if _, banned := p.banned[k]; !banned {
			active++
		}
	}
	return active
}
