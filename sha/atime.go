// +build !linux

package sha

import (
	"os"
)

// Open a file with no atime modification, if that is supported by the
// platform.
func openNoAtime(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY, 0)
}
