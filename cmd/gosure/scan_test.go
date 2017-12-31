package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"davidb.org/x/gosure/status"
	"davidb.org/x/gosure/store"
	"davidb.org/x/gosure/suredrive"
)

// Verify that we "gracefully" handle a snapshot when files are
// unreadable.
func TestUnreadableFile(t *testing.T) {
	tdir, err := ioutil.TempDir("", "gosure-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	// Create a single file, and make it unreadable.
	makeUnreadableFile(t, tdir)

	var st store.Store
	err = st.Parse(filepath.Join(tdir, "2sure.dat.gz"))
	if err != nil {
		t.Fatal(err)
	}

	mgr := status.NewManager()
	defer mgr.Close()

	err = suredrive.Scan(&st, tdir, mgr)
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
