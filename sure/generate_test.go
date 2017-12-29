package sure

import (
	"bytes"
	"fmt"
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
			t.Logf("delta output:\n%s", compBuf.String())
			t.Fatal("Trees differ")
		}
	}
}

// Benchmark extracting attributes out of a set of files.
func BenchmarkGetAtts(b *testing.B) {
	r := rand.New(rand.NewSource(1))

	var files []*File

	for i := 0; i < b.N; i++ {
		files = append(files, generateFile(r))
	}
	b.ResetTimer()

	var atts []string
	for _, f := range files {
		atts = append(atts, encodeAtts(f.Atts))
	}
	result1 = atts
}

// Benchmark decoding the attribute maps.
func BenchmarkSetAtts(b *testing.B) {
	r := rand.New(rand.NewSource(1))

	var atts []string

	for i := 0; i < b.N; i++ {
		f := generateFile(r)
		atts = append(atts, "simple ["+encodeAtts(f.Atts)+"]")
	}
	b.ResetTimer()

	var decoded []AttMap
	for _, at := range atts {
		var name string
		var atts AttMap
		err := parseNameAtts(at, &name, &atts)
		if err != nil {
			b.Logf("att: %q", at)
			b.Fatal(err)
		}
		decoded = append(decoded, atts)
	}
	allatts = decoded
}

// Benchmark decoding the attribute maps as maps
func BenchmarkMapAtts(b *testing.B) {
	r := rand.New(rand.NewSource(1))

	var atts []string

	for i := 0; i < b.N; i++ {
		f := generateFile(r)
		atts = append(atts, "simple ["+encodeAtts(f.Atts)+"]")
	}
	b.ResetTimer()

	var decoded [][]stringPair
	for _, at := range atts {
		_, atts, err := parseNameAtts2(at)
		if err != nil {
			b.Logf("att: %q", at)
			b.Fatal(err)
		}
		decoded = append(decoded, atts)
	}
	allatts2 = decoded
}

var result1 []string
var allatts []AttMap
var allatts2 [][]stringPair

// parseNameAtts2 parses the name and attributes from the encoded
// line, returning them.  This is copied from the main source here an
// modified for the alternate output.
func parseNameAtts2(line string) (string, []stringPair, error) {
	i := 0
	tname, err := scanName(line, &i)
	if err != nil {
		return "", nil, err
	}

	if i >= len(line) || line[i] != '[' {
		return "", nil, fmt.Errorf("Expecting '['")
	}
	i++

	var allAtts []stringPair
	for {
		if i >= len(line) {
			return "", nil, SyntaxError
		}
		if line[i] == ']' {
			return tname, allAtts, nil
		}

		key, err := scanName(line, &i)
		if err != nil {
			return "", nil, err
		}
		value, err := scanName(line, &i)
		if err != nil {
			return "", nil, err
		}

		allAtts = append(allAtts, stringPair{
			key:   key,
			value: value,
		})
	}
}
