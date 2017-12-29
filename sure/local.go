package sure

import (
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"syscall"
)

// Walk a directory tree, generating a tree structure for it.  All
// attributes are filled in that can be gleaned through lstat (and
// possibly readlink).  The files themselves are not opened.
func ScanFs(path string) (tree *Tree, err error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return
	}

	if !stat.IsDir() {
		err = errors.New("Expecting directory for walk")
		return
	}

	return walkFs("__root__", path, stat)
}

// Walk an already statted (directory) node.
func walkFs(name, fullName string, stat os.FileInfo) (tree *Tree, err error) {
	tree = &Tree{
		Name: name,
		Atts: getAtts(fullName, stat),
	}

	entries, err := readdir(fullName)
	if err != nil {
		return
	}

	sort.Sort(byName(entries))

	for _, ent := range entries {
		// log.Printf("Walk: %q", ent.Name())
		if ent.IsDir() {
			var child *Tree
			child, err = walkFs(ent.Name(),
				path.Join(fullName, ent.Name()), ent)
			if err != nil {
				log.Printf("Unable to stat %q: %v", path.Join(fullName, ent.Name()), err)
				continue
			}
			tree.Children = append(tree.Children, child)
		} else {
			node := &File{
				Name: ent.Name(),
				Atts: getAtts(path.Join(fullName, ent.Name()), ent),
			}
			tree.Files = append(tree.Files, node)
		}
	}

	return
}

// readdir reads all of the entries in the given directory.  This
// works like File.Readdir, but skips entries that aren't able to be
// statted (with a warning) (instead of discarding all of the rest).
// Unlike File.Readdir, this does not return "." or "..", and the
// result can be an empty slice.
func readdir(path string) ([]os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	names, err := file.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	fi := make([]os.FileInfo, 0, len(names))
	for _, filename := range names {
		fip, lerr := os.Lstat(filepath.Join(path, filename))
		if lerr != nil {
			// TODO Should warn here.
			continue
		}
		fi = append(fi, fip)
	}

	return fi, nil
}

func getAtts(name string, info os.FileInfo) AttMap {
	var atts AttMap
	sys := info.Sys().(*syscall.Stat_t)

	switch sys.Mode & syscall.S_IFMT {
	case syscall.S_IFDIR:
		dirAtts := &DirAtts{}
		basePerms(&dirAtts.BaseAtts, sys)
		atts = dirAtts
	case syscall.S_IFREG:
		mtime, ctime := getSysTimes(sys)
		regAtts := &RegAtts{
			Mtime: mtime,
			Ctime: ctime,
			Ino:   sys.Ino,
			Size:  sys.Size,
		}
		basePerms(&regAtts.BaseAtts, sys)
		atts = regAtts
	case syscall.S_IFLNK:
		lnkAtts := &LinkAtts{}
		basePerms(&lnkAtts.BaseAtts, sys)
		target, err := os.Readlink(name)
		if err != nil {
			log.Printf("Error reading symlink: %v", err)
		} else {
			lnkAtts.Targ = target
		}
		atts = lnkAtts
	case syscall.S_IFIFO:
		fifoAtts := &FifoAtts{Kind: syscall.S_IFIFO}
		basePerms(&fifoAtts.BaseAtts, sys)
		atts = fifoAtts
	case syscall.S_IFSOCK:
		fifoAtts := &FifoAtts{Kind: syscall.S_IFSOCK}
		basePerms(&fifoAtts.BaseAtts, sys)
		atts = fifoAtts
	case syscall.S_IFCHR:
		devAtts := &DevAtts{
			Kind: syscall.S_IFCHR,
			Rdev: uint64(sys.Rdev),
		}
		basePerms(&devAtts.BaseAtts, sys)
		atts = devAtts
		// TODO: These should have time info on them?
	case syscall.S_IFBLK:
		devAtts := &DevAtts{
			Kind: syscall.S_IFBLK,
			Rdev: uint64(sys.Rdev),
		}
		basePerms(&devAtts.BaseAtts, sys)
		atts = devAtts
	default:
		log.Printf("Node: %+v", info)
		panic("Unexpected file type")
	}

	return atts
}

// Base permissions shared by most nodes
func basePerms(atts *BaseAtts, sys *syscall.Stat_t) {
	atts.Uid = sys.Uid
	atts.Gid = sys.Gid
	atts.Perm = permission(sys)
}

// The Permission() call in 'os' masks off too many bits.
func permission(sys *syscall.Stat_t) uint32 {
	return uint32(sys.Mode &^ syscall.S_IFMT)
}

// For sorting by name
type byName []os.FileInfo

func (p byName) Len() int           { return len(p) }
func (p byName) Less(i, j int) bool { return p[i].Name() < p[j].Name() }
func (p byName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
