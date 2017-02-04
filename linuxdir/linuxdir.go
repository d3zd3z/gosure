package linuxdir

// Go's readdir doesn't return the inode number.  Sorting entries by
// inode number before statting them can prevent lots of seeks during
// stat on some filesystems.

// #include <sys/types.h>
// #include <dirent.h>
import "C"

import "unsafe"

type Dir C.DIR

func Open(name string) (result *Dir, err error) {
	tmp, err := C.opendir(C.CString(name))
	if err != nil {
		return
	}

	result = (*Dir)(tmp)
	return
}

func (p *Dir) Close() {
	C.closedir((*C.DIR)(p))
}

type Dirent struct {
	Name string
	Ino  uint64
}

// TODO, change this to readdir_r.
func (p *Dir) Readdir() (entry *Dirent, err error) {
	ent, err := C.readdir((*C.DIR)(p))
	if ent == nil {
		entry = nil
		// TODO: Correctly determine eof.
		return
	}
	entry = &Dirent{Ino: uint64(ent.d_ino)}

	// Convert the name.
	bytes := (*[10000]byte)(unsafe.Pointer(&ent.d_name[0]))
	entry.Name = string(bytes[0:clen(bytes[:])])
	return
}

func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}
