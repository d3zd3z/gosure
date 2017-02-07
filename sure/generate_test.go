package sure

import (
	"bytes"
	"math/rand"
	"testing"
)

// Generate some random trees, encode and decode them, and compare the
// results.
func TestGenerate(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	for i := 0; i < 3; i++ {
		tr := GenerateTree(r, 20, 3)

		// Encode it to a buffer.
		var buf bytes.Buffer
		tr.Encode(&buf)

		tr2, err := Decode(&buf)
		if err != nil {
			t.Fatal(err)
		}

		var compBuf bytes.Buffer
		NewComparer(&compBuf).CompareTrees(tr, tr2)

		if compBuf.Len() > 0 {
			t.Log("delta output:\n%s", compBuf.String())
			t.Fatal("Trees differ")
		}
	}
}
