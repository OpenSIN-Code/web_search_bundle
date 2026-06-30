// SPDX-License-Identifier: MIT
// Purpose: Singleflight coalescing for identical concurrent queries.
package orchestrator

import "sync"

// singleFlightCall represents an in-flight or completed query.
type singleFlightCall struct {
	wg  sync.WaitGroup
	val *SearchResult
	err error
}

// singleFlightGroup deduplicates concurrent identical queries.
// Only the first caller executes the query; subsequent callers
// with the same key block until the first completes and share
// the result.
type singleFlightGroup struct {
	mu    sync.Mutex
	calls map[string]*singleFlightCall
}

// Do executes fn only once for a given key among concurrent callers.
// If a call with the same key is already in flight, the caller blocks
// until it completes and receives the same result.
func (g *singleFlightGroup) Do(key string, fn func() (*SearchResult, error)) (*SearchResult, error) {
	g.mu.Lock()
	if g.calls == nil {
		g.calls = make(map[string]*singleFlightCall)
	}

	// If a call is in flight, wait for it.
	if c, ok := g.calls[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	// First caller — create the call and execute.
	c := &singleFlightCall{}
	c.wg.Add(1)
	g.calls[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	// Remove the completed call so future requests re-execute.
	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()

	return c.val, c.err
}
