// Hash a file, using native Go libraries.

package sha

import (
	"io"
)

func HashFile(path string) (result []byte, err error) {
	hash := NewSha1()
	/*
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOATIME, 0)
	if err != nil {
		file, err = os.OpenFile(path, os.O_RDONLY, 0)
	}
	*/
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

		hash.Update(buffer[0:n])
	}
	putBuffer(buffer)
	result = hash.Final()
	return
}

// Keep a pool of buffers.  The size will be the number of potential
// buffers in the pool.
var bufPool chan []byte = make(chan []byte, 16)

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
