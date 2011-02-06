// Hash a file, using native Go libraries.

package main

import (
	"os"
	"crypto/sha1"
	"syscall"
)

func HashFile(path string) (result []byte, err os.Error) {
	hash := sha1.New()
	file, err := os.Open(path, os.O_RDONLY | syscall.O_NOATIME, 0)
	if err != nil {
		file, err = os.Open(path, os.O_RDONLY, 0)
	}
	if err != nil {
		return
	}
	defer file.Close()

	buffer := make([]byte, 65536)
	for {
		n, err := file.Read(buffer)
		if err == os.EOF {
			break
		}
		if err != nil {
			// TODO: Warn
			return
		}

		nn, err := hash.Write(buffer[0:n])
		if err != nil { return }
		if nn != n {
			err = os.NewError("Short write on hash")
			return
		}
	}
	result = hash.Sum()
	return
}
