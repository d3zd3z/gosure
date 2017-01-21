package linuxdir

import (
	"os"
	"sort"
)

// Read all of the entries in the named directory, returning the
// built-up fileinfo for them.
// This improves performance on some filesystem by first sorting the
// directory entries by inode, statting them in that order.  The
// result is then sorted by name.
func Readdir(name string) (entries []os.FileInfo, err error) {
	dir, err := Open(name)
	if err != nil {
		return
	}
	defer dir.Close()

	names := make([]*Dirent, 0, 100)
	for {
		var entry *Dirent
		entry, err = dir.Readdir()
		if entry == nil {
			if err != nil {
				// TODO: This should probably log or
				// something.
			}
			break
		}
		if entry.Name != "." && entry.Name != ".." {
			names = append(names, entry)
		}
	}
	sort.Sort((*InodeSort)(&names))

	entries = make([]os.FileInfo, 0, len(names))
	for i := range names {
		var tmp os.FileInfo
		tmp, err = os.Lstat(name + "/" + names[i].Name)
		if err != nil {
			// TODO: Warn again.  But, skip the name if we
			// can't stat it.
		} else {
			entries = append(entries, tmp)
		}
	}
	sort.Sort((*NameSort)(&entries))
	return
}

// Sorting Dirent by inode number.
type InodeSort []*Dirent

func (p InodeSort) Len() int           { return len(p) }
func (p InodeSort) Less(i, j int) bool { return p[i].Ino < p[j].Ino }
func (p InodeSort) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type NameSort []os.FileInfo

func (p NameSort) Len() int           { return len(p) }
func (p NameSort) Less(i, j int) bool { return p[i].Name() < p[j].Name() }
func (p NameSort) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
