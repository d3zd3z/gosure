package weave_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"davidb.org/x/gosure/weave"
)

// Test that we can create basic temp names.
func TestTemps(t *testing.T) {
	tdir, err := ioutil.TempDir("", "weave-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	nam := weave.SimpleNaming{
		Path:       tdir,
		Base:       "hahafile",
		Ext:        "weave",
		Compressed: true,
	}

	// Make sure we can create a bunch.
	for i := 0; i < 10; i++ {
		file1, err := weave.TempFile(&nam, true)
		if err != nil {
			t.Fatal(err)
		}
		file1.Close()
		name1 := file1.Name()
		expect := fmt.Sprintf("%s/hahafile.%d.gz", tdir, i)
		if name1 != expect {
			t.Fatalf("Expecting name %q, got %q", expect, name1)
		}
	}
}

// Test that temp fails properly if the directory is invalid.
func TestInvalidName(t *testing.T) {
	nam := weave.SimpleNaming{
		Path:       "/invalid/path/to/file",
		Base:       "stuff",
		Ext:        "weave",
		Compressed: true,
	}

	_, err := weave.TempFile(&nam, true)
	if err == nil {
		t.Fatal("Should not have been able to create temp file")
	}
}
