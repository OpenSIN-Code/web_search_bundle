// Purpose: Unit tests for the key pool.
// Docs: internal/pool/keypool_test.doc.md
package pool

import (
	"testing"
	"time"
)

func TestKeyPool(t *testing.T) {
	p := New([]string{"key1", "key2", "key3"})
	if p.IsEmpty() {
		t.Fatal("pool should not be empty")
	}
	key, err := p.Next()
	if err != nil {
		t.Fatal(err)
	}
	if key == "" {
		t.Fatal("expected a key")
	}
	p.Ban(key, time.Minute)
	if p.Count() != 2 {
		t.Fatalf("expected 2 active keys, got %d", p.Count())
	}
}
