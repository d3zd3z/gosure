package store

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"

	"davidb.org/code/gosure/sure"
)

// The Store represents the current surefile store.  The default
// values will result in a the files "./2sure.dat.gz" and the likes
// being used.
type Store struct {
	Path  string // The directory where the surefiles will be written.
	Base  string // The initial part of the name.
	Plain bool   // Plain indicates the files should not be compressed.
}

// Write the tree to the surefile, archiving a previous version.
func (s *Store) Write(tree *sure.Tree) error {
	tname, err := s.writeTemp(tree)
	if err != nil {
		// Depending on where the failure happened, the file
		// may have been written, so try to erase it, ignoring
		// any error.
		if tname != "" {
			os.Remove(tname)
		}
		return err
	}

	os.Rename(s.datName(), s.bakName())
	err = os.Rename(tname, s.datName())
	if err != nil {
		return err
	}

	return nil
}

// Read a tree from the data file.
func (s *Store) ReadDat() (*sure.Tree, error) {
	return s.readNamed(s.datName())
}

// Read a tree from the backup file.
func (s *Store) ReadBak() (*sure.Tree, error) {
	return s.readNamed(s.bakName())
}

// Read a tree from the given pathname.
func (s *Store) readNamed(name string) (*sure.Tree, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rd io.Reader
	if s.Plain {
		rd = f
	} else {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		rd = gz
	}

	return sure.Decode(rd)
}

// Write out the tree to a temp file, returning the name of the temp
// file.
func (s *Store) writeTemp(tree *sure.Tree) (string, error) {
	f, err := s.tmpFile()
	if err != nil {
		return "", err
	}
	defer f.Close()

	name := f.Name()

	var wr io.Writer
	if s.Plain {
		wr = f
	} else {
		gz := gzip.NewWriter(f)
		defer gz.Close()
		wr = gz
	}

	err = tree.Encode(wr)
	if err != nil {
		return "", err
	}

	return name, nil
}

// tmpFile tries to open a new file, without overwriting a file.
func (s *Store) tmpFile() (*os.File, error) {
	n := 0
	for {
		name := s.makeName(strconv.Itoa(n))

		f, err := os.OpenFile(name, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0644)
		if err == nil {
			return f, nil
		}
		if os.IsExist(err) {
			n++
			continue
		}
		return nil, err
	}
}

// makeName generates a filename with the given string as the
// extension part of the name
func (s *Store) makeName(ext string) string {
	base := s.Base
	if base == "" {
		base = "2sure"
	}

	gz := ".gz"
	if s.Plain {
		gz = ""
	}

	return path.Join(s.Path, fmt.Sprintf("%s.%s%s", base, ext, gz))
}

// datName returns the pathname for the primary dat file.
func (s *Store) datName() string {
	return s.makeName("dat")
}

// bakName returns the pathname for the backup file.
func (s *Store) bakName() string {
	return s.makeName("bak")
}
