package store // import "davidb.org/x/gosure/store"

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"

	"davidb.org/x/gosure/sure"
	"davidb.org/x/gosure/weave"
)

// The Store represents the current surefile store.  The default
// values will result in a the files "./2sure.dat.gz" and the likes
// being used.
type Store struct {
	Path  string            // The directory where the surefiles will be written.
	Base  string            // The initial part of the name.
	Ext   string            // The extension to use "normally dat"
	Plain bool              // Plain indicates the files should not be compressed.
	Tags  map[string]string // For delta stores, indicates tags for next delta written.
	Name  string            // The name used to describe this capture.
}

// Write writes a new version to the surefile.
func (s *Store) Write(tree *sure.Tree) error {
	s.FixTags()

	// TODO: Check for an existing file and make a delta.
	base, err := s.GetDelta(DeltaLatest)
	if err == nil {
		return s.WriteDelta(tree, base)
	}
	if !os.IsNotExist(err) {
		return err
	}

	wr, err := weave.NewNewWeave(s, s.Name, s.Tags)
	if err != nil {
		return err
	}
	// Note that we explicitly don't close this if there is a
	// problem.  It is better to leave files around than write a
	// blank one.  TODO: Figure out how to handle this better.

	err = tree.Encode(wr)
	if err != nil {
		return err
	}

	return wr.Close()
}

// WriteDelta writes a new delta to the surefile, knowing the previous
// version.
func (s *Store) WriteDelta(tree *sure.Tree, base int) error {
	wr, err := weave.NewDeltaWriter(s, base, s.Name, s.Tags)
	if err != nil {
		return err
	}
	// Don't explicitly close to avoid corrupt file.

	err = tree.Encode(wr)
	if err != nil {
		return err
	}

	return wr.Close()
}

// Magic delta numbers to refer to previous deltas
// TODO: Interpret these the same as slices in python to be more
// flexible.
const (
	DeltaLatest = -1 // The most recent version.
	DeltaPrior  = -2 // The second to most recent version.
)

// ReadDat reads a tree from the data file.
func (s *Store) ReadDat() (*sure.Tree, error) {
	return s.ReadDelta(DeltaLatest)
}

// ReadBak reads a tree from the backup file.
func (s *Store) ReadBak() (*sure.Tree, error) {
	return s.ReadDelta(DeltaPrior)
}

// ReadDelta reads a given delta number.  The delta can be a number
// retrieved from the header, or it can be one of the above
// DeltaLatest, or DeltaPrior constants to read specific recent or
// nearly recent versions.
func (s *Store) ReadDelta(num int) (*sure.Tree, error) {
	num, err := s.GetDelta(num)
	if err != nil {
		return nil, err
	}

	pd := sure.NewPushDecoder()

	err = weave.ReadDelta(s, num, func(text string) error {
		return pd.Add(text)
	})
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	return pd.Tree()
}

// GetDelta canonicalizes a delta number.  Returns an error if there
// was no delta to read.
func (s *Store) GetDelta(num int) (int, error) {
	hdr, err := weave.ReadHeader(s)
	if err != nil {
		return 0, err
	}

	// Adjust the delta number to handle ones near the end.
	switch num {
	case DeltaLatest:
		num = hdr.LatestDelta()
		if num == 0 {
			return 0, errors.New("No versions in surefile")
		}
	case DeltaPrior:
		num, err = hdr.PenultimateDelta()
		if err != nil {
			return 0, err
		}
	}

	return num, nil
}

// ReadHeader attempts to read the header from the storefile.
func (s *Store) ReadHeader() (*weave.Header, error) {
	return weave.ReadHeader(s)
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
		name := s.makeName(strconv.Itoa(n), !s.Plain)

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
func (s *Store) makeName(ext string, compressed bool) string {
	base := s.Base
	if base == "" {
		base = "2sure"
	}

	gz := ".gz"
	if !compressed {
		gz = ""
	}

	return path.Join(s.Path, fmt.Sprintf("%s.%s%s", base, ext, gz))
}

// datName returns the pathname for the primary dat file.
func (s *Store) datName() string {
	ext := "dat"
	if s.Ext != "" {
		ext = s.Ext
	}
	return s.makeName(ext, !s.Plain)
}

// bakName returns the pathname for the backup file.
func (s *Store) bakName() string {
	return s.makeName("bak", !s.Plain)
}

// TempFile is used by the naming convention to generate temp files.
// We just use the number as the extension.
func (s *Store) TempFile(num int, compressed bool) string {
	return s.makeName(strconv.Itoa(num), compressed)
}

// MainFile return the main file name.
func (s *Store) MainFile() string {
	return s.datName()
}

// BackupFile returns the name of the backup file.
func (s *Store) BackupFile() string {
	return s.bakName()
}

// IsCompressed returns whether or not this store is compressed.
func (s *Store) IsCompressed() bool {
	return !s.Plain
}
