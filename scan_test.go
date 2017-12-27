package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"davidb.org/x/gosure/store"
	"davidb.org/x/gosure/sure"
)

// Verify that we "gracefully" handle a snapshot when files are
// unreadable.
func TestUnreadableFile(t *testing.T) {
	tdir, err := ioutil.TempDir("", "gosure-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	// TODO: Once #4 is fixed, we should be able to scan the dir
	// without having to enter it.
	err = os.Chdir(tdir)
	if err != nil {
		t.Fatal(err)
	}

	// Create a single file, and make it unreadable.
	makeUnreadableFile(t, tdir)

	tree, err := sure.ScanFs(".")
	if err != nil {
		t.Fatal(err)
	}

	hashUpdate(tree)

	var st store.Store
	err = st.Parse("2sure.dat.gz")
	if err != nil {
		t.Fatal(err)
	}

	err = st.Write(tree)
	if err != nil {
		t.Fatal(err)
	}
}

func makeUnreadableFile(t *testing.T, dir string) {
	name := filepath.Join(dir, "unreadable.txt")

	file, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	fmt.Fprintf(file, "Sample contents\n")

	err = file.Chmod(0)
	if err != nil {
		t.Fatal(err)
	}
}
