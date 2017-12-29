package sure // import "davidb.org/x/gosure/sure"

import (
	"syscall"
)

// BaseAtts are attributes associated with all most types.
type BaseAtts struct {
	Uid  uint32
	Gid  uint32
	Perm uint32
}

type DirAtts struct {
	BaseAtts
}

func (r *DirAtts) GetKind() string { return "dir" }

type RegAtts struct {
	BaseAtts
	Mtime int64 // TODO: Store better than seconds.
	Ctime int64 // TODO: Store better than seconds.
	Ino   uint64
	Size  int64
	Sha1  []byte
}

func (r *RegAtts) GetKind() string { return "file" }

type LinkAtts struct {
	BaseAtts
	Targ string
}

func (r *LinkAtts) GetKind() string { return "lnk" }

// FifoAtts is for both fifos and sockets.
type FifoAtts struct {
	Kind uint32
	BaseAtts
}

func (a *FifoAtts) GetKind() string {
	if a.Kind == syscall.S_IFIFO {
		return "fifo"
	} else {
		return "sock"
	}
}

// DevAtts is for block and character nodes.
type DevAtts struct {
	Kind uint32
	BaseAtts
	Rdev uint64
}

func (a *DevAtts) GetKind() string {
	if a.Kind == syscall.S_IFBLK {
		return "blk"
	} else {
		return "chr"
	}
}
