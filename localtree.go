// Walk a local tree.

package main

import (
	"os"
	"fmt"
)

func walkTree(base string) (NodeReader, os.Error) {
	fd, err := os.Open(base, os.O_RDONLY, 0)
	if err != nil {
		goto error
	}
	defer fd.Close()

	// Note that Readdir doesn't process names in inode order, so
	// will probably be very inefficient on large directories.
	children, err := fd.Readdir(-1)
	if err != nil {
		goto error
	}
	dirs := make([]os.FileInfo, 0, len(children))
	files := make([]os.FileInfo, 0, len(children))
	for _, child := range children {
		if child.IsDirectory() {
			dirs = append(dirs, child)
		} else {
			files = append(files, child)
		}
	}

	fmt.Printf("Dirs:\n")
	for _, node := range dirs {
		fmt.Printf("   %s\n", node.Name)
	}
	fmt.Printf("Files\n")
	for _, node := range files {
		fmt.Printf("   %s\n", node.Name)
	}

	return nil, os.NewError("TODO")

error:
	return nil, err
}
