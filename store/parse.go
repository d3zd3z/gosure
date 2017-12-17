package store

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

// Parse attempts to determine the parameters of the Store structure
// based on a user-specified path.  The path specified can be the path
// to a directory.  In this case, we will look at possible filenames
// to determine the other parameters.  The path can also give a
// filename of one of the surefiles, and we will derive the name
// information from that.
//
// If the error result is nil, the Store will have the parameters set
// according to user preferences.  Otherwise, the error will give
// details, and the Store will remain unchanged.
func (s *Store) Parse(name string) error {
	if fi, err := os.Stat(name); err == nil && fi.IsDir() {
		// The path given is a directory, use the defaults.
		s.Path = name
		s.Base = ""
		s.Plain = false
		return nil
	}

	// Try splitting off the last filename from the path, and see
	// if we find that directory.
	dir := path.Dir(name)
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return NotDir(name)
	}

	plain := true
	base := path.Base(name)

	if strings.HasSuffix(base, ".gz") {
		base = base[:len(base)-3]
		plain = false
	}

	// Strip off the known suffixes.
	ext := path.Ext(base)
	if ext == ".dat" || ext == ".bak" {
		base = base[:len(base)-4]
	} else if ext != "" {
		return InvalidName(name)
	} else {
		// If no extension was given, use compression by
		// default.
		plain = false
	}

	s.Path = dir
	s.Base = base
	s.Plain = plain

	return nil
}

// NotDir is an error returned when the name doesn't describe an
// existing path.
type NotDir string

func (n NotDir) Error() string {
	return fmt.Sprintf("Path %q is not in an existant directory", string(n))
}

// InvalidName is an error returned if the name contains an unknown
// extension.
type InvalidName string

func (n InvalidName) Error() string {
	return fmt.Sprintf("Path %q has an unknown extension", string(n))
}

// Implement 'Value' from spf13/pflag so the store can be directly
// used as a command line argument.

func (s *Store) String() string {
	base := s.Base
	if base == "" {
		base = "2sure"
	}
	ext := ".gz"
	if s.Plain {
		ext = ""
	}
	return path.Join(s.Path, base+".dat"+ext)
}

func (s *Store) Set(value string) error {
	return s.Parse(value)
}

func (s *Store) Type() string {
	return "surefile path"
}

// Wrap the store as a 'Value' to use the tags as a command line
// argument.
type Tags struct {
	store *Store
}

func NewTags(store *Store) Tags {
	return Tags{
		store: store,
	}
}

func (t Tags) String() string {
	var buf bytes.Buffer

	first := true
	for k, v := range t.store.Tags {
		if !first {
			fmt.Fprintf(&buf, ", ")
		}
		first = false
		fmt.Fprintf(&buf, "%s=%s", k, v)
	}

	return buf.String()
}

func (t Tags) Set(value string) error {
	f := strings.SplitN(value, "=", 2)
	if len(f) != 2 {
		return errors.New("--tag value must contain an '='")
	}

	// TODO: Check for duplicates.
	t.store.Tags[f[0]] = f[1]
	return nil
}

func (t Tags) Type() string {
	return "tags"
}
