// Hash a file, using native Go libraries.

package sha

import (
	"crypto/sha1"
	"io"
)

// HashFile computes the sha1 hash of the named file.  If successful,
// the result will be a 20-byte string of hash data, otherwise err
// will be set to the cause of the failure.
//
// On some platforms (notably Linux), this will try to not update the
// atime on the file, so that it is still useful.
func HashFile(path string) (result []byte, err error) {
	hash := sha1.New()
	file, err := openNoAtime(path)
	if err != nil {
		return
	}
	defer file.Close()

	buffer := getBuffer()
	for {
		var n int
		n, err = file.Read(buffer)
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			// TODO: Warn
			return
		}

		_, _ = hash.Write(buffer[0:n])
	}
	putBuffer(buffer)
	result = hash.Sum(nil)
	return
}

// Keep a pool of buffers.  The size will be the number of potential
// buffers in the pool.
var bufPool = make(chan []byte, 16)

// Fetch a buffer for use, allocating if the pool is empty.
func getBuffer() []byte {
	select {
	case buf := <-bufPool:
		return buf
	default:
		return make([]byte, 65536)
	}
}

// Return the buffer to the pool, discarding it if the pool is full.
func putBuffer(buf []byte) {
	select {
	case bufPool <- buf:
		// Pushed.  Default discards and lets the GC clean it
		// up.
	}
}
