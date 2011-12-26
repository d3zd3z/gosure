// Hash a file, using native Go libraries.

package sha

import (
	"os"
	"syscall"
)

func HashFile(path string) (result []byte, err os.Error) {
	hash := NewSha1()
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOATIME, 0)
	if err != nil {
		file, err = os.OpenFile(path, os.O_RDONLY, 0)
	}
	if err != nil {
		return
	}
	defer file.Close()

	buffer := make([]byte, 65536)
	for {
		var n int
		n, err = file.Read(buffer)
		if err == os.EOF {
			break
		}
		if err != nil {
			// TODO: Warn
			return
		}

		hash.Update(buffer[0:n])
	}
	result = hash.Final()
	return
}
