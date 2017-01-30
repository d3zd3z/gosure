package sure

// Attributes are stored as simple key/value pairs, always as strings.
// The particular attributes are defined by the file readers.  This
// type captures all known attributes in a more space-efficient
// internal structure.
//
// TODO: These types are Linux-specific, and should be built that way
// as well.
type AttMap struct {
	Kind   uint32 // Uses the syscall S_IF* values.
	Uid    uint32
	Gid    uint32
	Perm   uint32
	Devmaj uint32
	Devmin uint32
	Mtime  int64 // TODO: Store better than seconds.
	Ctime  int64 // TODO: Store better than seconds.
	Ino    uint64
	Size   int64
	Sha    []byte
	Targ   string
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
