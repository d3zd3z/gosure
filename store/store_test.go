package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"

	"davidb.org/x/gosure/sure"
)

func TestTmpFile(t *testing.T) {
	tdir, err := ioutil.TempDir("", "store-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	var st Store

	st.Path = tdir

	// Make sure we can create a large number of these
	for i := 0; i < 100; i++ {
		f, err := st.tmpFile()
		if err != nil {
			t.Fatal(err)
		}
		f.Close()

		// Verify that the name is correct.
		name := path.Join(tdir, fmt.Sprintf("2sure.%d.gz", i))
		if f.Name() != name {
			t.Fatalf("Tmp name mismatch, expect %q, got %q", f.Name(), name)
		}
	}

	// Verify that we can't write to an invalid name.
	st.Path = path.Join(tdir, "notdir")
	f, err := st.tmpFile()
	if err == nil {
		name := f.Name()
		f.Close()
		t.Fatalf("Should not have been able to create file %q", name)
	}

	// Try with a file instead of a name.
	st.Path = path.Join(tdir, "2sure.0.gz")
	f, err = st.tmpFile()
	if err == nil {
		name := f.Name()
		f.Close()
		t.Fatalf("Should not have been able to create file %q", name)
	}
}

func TestWrite(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	tdir, err := ioutil.TempDir("", "store-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	// Write 3 times, and then make sure the files we get back are
	// right.
	var trees []*sure.Tree

	var st Store
	st.Path = tdir

	for i := 0; i < 3; i++ {
		tr := sure.GenerateTree(r, 10, 2)
		trees = append(trees, tr)

		err := st.Write(tr)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Check that we only have the two names.
	files, err := ioutil.ReadDir(tdir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("Wrong number of files: %d, expect %d", len(files), 2)
	}

	for _, fi := range files {
		if fi.Name() == "2sure.dat.gz" || fi.Name() == "2sure.bak.gz" {
			continue
		}
		t.Fatalf("File: %q unexpected", fi.Name())
	}

	t2, err := st.ReadDat()
	if err != nil {
		t.Fatal(err)
	}
	treesSame(t, trees[2], t2)

	t1, err := st.ReadBak()
	if err != nil {
		t.Fatal(err)
	}
	treesSame(t, trees[1], t1)
}

func treesSame(t *testing.T, a, b *sure.Tree) {
	var buf bytes.Buffer
	sure.NewComparer(&buf).CompareTrees(a, b)

	if buf.Len() > 0 {
		t.Logf("delta output: %q", buf.String())
		t.Fatal("Trees differ")
	}
}
