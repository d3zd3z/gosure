package sha_test

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"davidb.org/x/gosure/sha"
)

func TestFile(t *testing.T) {
	// To make things predictable, make a file with known
	// contents.
	name, err := genFile(256*1024 + 1765)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(name)

	h1, err := sha.HashFile(name)
	if err != nil {
		t.Fatal(err)
	}

	// Use external tools to compute the sha1, assuming these are
	// available.
	h2text, err := exec.Command("sha1sum", name).Output()
	if err != nil {
		// Try the gsha1sum, which may be the case with
		// homebrew.
		h2text, err = exec.Command("gsha1sum", name).Output()
		if err != nil {
			t.Fatal(err)
		}
	}
	fields := strings.Split(string(h2text), " ")
	h2, err := hex.DecodeString(string(fields[0]))
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(h1, h2) != 0 {
		t.Fatalf("hash unequal: % x, % x", h1, h2)
	}
}

// Generate a temp file with 'n' bytes of random data in it, returning
// the name.
func genFile(size int) (string, error) {
	f, err := ioutil.TempFile("/var/tmp", "tfile")
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 1024)

	state := uint32(0x12345678)

	for size > 0 {
		if size < len(buf) {
			buf = buf[0:size]
		}

		for i := 0; i < len(buf); i++ {
			state = nextXor(state)
			buf[i] = byte(state & 0xFF)
		}

		_, err = f.Write(buf)
		if err != nil {
			// Try to clean up
			os.Remove(f.Name())
			return "", err
		}

		size -= len(buf)
	}

	name := f.Name()
	return name, nil
}

// This 32-bit xor-shift (https://en.wikipedia.org/wiki/Xorshift) and
// has a period of 2^32-1.  This is more than adequate for the needs
// of this test.
func nextXor(state uint32) uint32 {
	state ^= state << 13
	state ^= state >> 17
	state ^= state << 5
	return state
}
