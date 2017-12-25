package sure

import (
	"errors"
	"log"
	"os"
	"path"
	"syscall"

	"davidb.org/x/gosure/linuxdir"
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

	entries, err := linuxdir.Readdir(fullName)
	if err != nil {
		return
	}

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

func getAtts(name string, info os.FileInfo) AttMap {
	var atts AttMap
	sys := info.Sys().(*syscall.Stat_t)

	switch sys.Mode & syscall.S_IFMT {
	case syscall.S_IFDIR:
		dirAtts := &DirAtts{}
		basePerms(&dirAtts.Base, sys)
		atts = dirAtts
	case syscall.S_IFREG:
		mtime, ctime := getSysTimes(sys)
		regAtts := &RegAtts{
			Mtime: mtime,
			Ctime: ctime,
			Ino:   sys.Ino,
			Size:  sys.Size,
		}
		basePerms(&regAtts.Base, sys)
		atts = regAtts
	case syscall.S_IFLNK:
		lnkAtts := &LinkAtts{}
		basePerms(&lnkAtts.Base, sys)
		target, err := os.Readlink(name)
		if err != nil {
			log.Printf("Error reading symlink: %v", err)
		} else {
			lnkAtts.Targ = target
		}
		atts = lnkAtts
	case syscall.S_IFIFO:
		fifoAtts := &FifoAtts{Kind: S_IFIFO}
		basePerms(&fifoAtts.Base, sys)
		atts = fifoAtts
	case syscall.S_IFSOCK:
		fifoAtts := &FifoAtts{Kind: S_IFSOCK}
		basePerms(&fifoAtts.Base, sys)
		atts = fifoAtts
	case syscall.S_IFCHR:
		devAtts := &DevAtts{
			Kind: S_IFCHR,
			Rdev: uint64(sys.Rdev),
		}
		basePerms(&devAtts.Base, sys)
		atts = devAtts
		// TODO: These should have time info on them?
	case syscall.S_IFBLK:
		devAtts := &DevAtts{
			Kind: S_IFBLK,
			Rdev: uint64(sys.Rdev),
		}
		basePerms(&devAtts.Base, sys)
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
