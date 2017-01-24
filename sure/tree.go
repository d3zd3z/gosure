package sure

import (
	"errors"
	"strconv"
)

// Attributes are stored as simple key/value pairs, always as strings.
// The particular attributes are defined by the file readers.
type AttMap map[string]string

// NoKey is the error returned by the below queries if the key is not
// present.
var NoKey = errors.New("NoKey")

// Attempt to retrieve a key from the map, and interpret it as a
// uint64.  Returns NoKey if the field is not present.  Will return
// another error (from the strconv.ParseUint) if the not is not a
// valid uint64.
func (a AttMap) GetUint64(key string) (uint64, error) {
	value, ok := a[key]
	if !ok {
		return 0, NoKey
	}

	return strconv.ParseUint(value, 10, 64)
}

// A directory tree node.
type Tree struct {
	Name     string
	Atts     AttMap
	Children []*Tree
	Files    []*File
}

// A non-directory node (not necessarily a plain file)
type File struct {
	Name string
	Atts AttMap
}
