// Hash a file, using native Go libraries.

package main

import (
	"os"
	"syscall"
	"./sha/_obj/sha"
)

func HashFile(path string) (result []byte, err os.Error) {
	hash := sha.NewSha1()
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

		hash.Update(buffer[0:n])
	}
	result = hash.Final()
	return
}
