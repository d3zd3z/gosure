package sha

import (
	"crypto/sha1"
	"testing"
)

var mainresult []byte

// Benchmark the standard library sha1
func BenchmarkGoSha1(b *testing.B) {
	buff := make([]byte, 1024)
	h := sha1.New()
	for i := 0; i < b.N; i++ {
		h.Write(buff)
	}
	mainresult = h.Sum(nil)
}
