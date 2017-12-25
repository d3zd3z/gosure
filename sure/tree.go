package sure // import "davidb.org/x/gosure/sure"

import (
	"syscall"
)

// The file kinds are imported directly from syscall.
const (
	S_IFDIR  = syscall.S_IFDIR
	S_IFREG  = syscall.S_IFREG
	S_IFLNK  = syscall.S_IFLNK
	S_IFIFO  = syscall.S_IFIFO
	S_IFSOCK = syscall.S_IFSOCK
	S_IFCHR  = syscall.S_IFCHR
	S_IFBLK  = syscall.S_IFBLK
)

// AttMap defines an interface to general attributes for file nodes.
type AttMap interface {
	GetKind() string

	// Compare these attributes with another node.  If the type of
	// the other node differs, report 'kind' and be done.
	// Otherwise, appends the field names of the attributes that
	// are missing.
	// Compare(other interface{}, atts []string) []string
}

// BaseAtts are attributes associated with all most types.
type BaseAtts struct {
	Uid  uint32
	Gid  uint32
	Perm uint32
}

type DirAtts struct {
	Base BaseAtts
}

func (r *DirAtts) GetKind() string { return "dir" }

type RegAtts struct {
	Base  BaseAtts
	Mtime int64 // TODO: Store better than seconds.
	Ctime int64 // TODO: Store better than seconds.
	Ino   uint64
	Size  int64
	Sha1  []byte
}

func (r *RegAtts) GetKind() string { return "file" }

type LinkAtts struct {
	Base BaseAtts
	Targ string
}

func (r *LinkAtts) GetKind() string { return "lnk" }

// FifoAtts is for both fifos and sockets.
type FifoAtts struct {
	Kind uint32
	Base BaseAtts
}

func (a *FifoAtts) GetKind() string {
	if a.Kind == S_IFIFO {
		return "fifo"
	} else {
		return "sock"
	}
}

// DevAtts is for block and character nodes.
type DevAtts struct {
	Kind uint32
	Base BaseAtts
	Rdev uint64
}

func (a *DevAtts) GetKind() string {
	if a.Kind == S_IFBLK {
		return "blk"
	} else {
		return "chr"
	}
}

// Attributes are stored as simple key/value pairs, always as strings.
// The particular attributes are defined by the file readers.  This
// type captures all known attributes in a more space-efficient
// internal structure.
//
// TODO: These types are Linux-specific, and should be built that way
// as well.
/*
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
*/

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
