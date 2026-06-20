package bench

import "testing"

func BenchmarkFast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = i + 1
	}
}
