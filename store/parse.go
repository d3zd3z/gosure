package store

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
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
	switch ext {
	case ".weave":
		base = base[:len(base)-6]
		s.Ext = "weave"
	case ".dat", ".bak":
		base = base[:len(base)-4]
	case "":
		// If no extension was given, use compression by
		// default.
		plain = false
	default:
		return InvalidName(name)
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

// Set sets the name of this store.  Used to parse the command line.
func (s *Store) Set(value string) error {
	return s.Parse(value)
}

// Type returns a short description of the type of the store.  Used by
// the command parsing to print help.
func (s *Store) Type() string {
	return "surefile path"
}

// Tags wraps the store as a 'Value' to use the tags as a command line
// argument.
type Tags struct {
	store *Store
}

// NewTags create a new Tags struct, wrapping a given store.
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

// Set adds a new tag, from the command line parsing.  Must be of the
// form key=value.
func (t Tags) Set(value string) error {
	f := strings.SplitN(value, "=", 2)
	if len(f) != 2 {
		return errors.New("--tag value must contain an '='")
	}

	// TODO: Check for duplicates.
	if t.store.Tags == nil {
		t.store.Tags = make(map[string]string)
	}
	t.store.Tags[f[0]] = f[1]
	return nil
}

// Type returns a descriptive name for this type, for help messages.
func (t Tags) Type() string {
	return "tags"
}

// FixTags adjusts the tags appropriately after any options from the
// user have been processed.  If 'name' is not given, it will have the
// same value as time.  The name will be deleted from the tags, and
// placed into the Name value in the struct.
func (s *Store) FixTags() {
	if s.Tags == nil {
		s.Tags = make(map[string]string)
	}

	// To make this function idempotent, move the name back into
	// the Tags list.
	_, ok := s.Tags["name"]
	if s.Name != "" && !ok {
		s.Tags["name"] = s.Name
		s.Name = ""
	}

	// If the name isn't set, use the time, otherwise extract from
	// the explicitly given name field.
	name, ok := s.Tags["name"]
	if ok {
		s.Name = name
		delete(s.Tags, "name")
	} else {
		s.Name = time.Now().UTC().Format(time.RFC3339Nano)
	}
}
