// SPDX-License-Identifier: MIT
// Purpose: Benchmark key pool rotation and active-key counting.
// Docs: keypool.doc.md
package pool

import "testing"

func BenchmarkKeyPoolNext10(b *testing.B) {
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	p := New(keys)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Next()
	}
}

func BenchmarkKeyPoolCount10(b *testing.B) {
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	p := New(keys)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Count()
	}
}
