// +build linux

package sha

import (
	"os"
	"syscall"
)

// Open a file with no atime modification, if that is supported by the
// platform.
func openNoAtime(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOATIME, 0)
	if err != nil {
		file, err = os.OpenFile(path, os.O_RDONLY, 0)
	}
	return file, err
}
