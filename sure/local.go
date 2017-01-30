package sure

import (
	"errors"
	"log"
	"os"
	"path"
	"strconv"
	"syscall"

	"davidb.org/code/gosure/linuxdir"
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

	atts.Kind = sys.Mode & syscall.S_IFMT
	switch sys.Mode & syscall.S_IFMT {
	case syscall.S_IFDIR:
		basePerms(&atts, sys)
	case syscall.S_IFREG:
		basePerms(&atts, sys)
		timeInfo(&atts, sys)
		atts.Ino = sys.Ino
		atts.Size = sys.Size
	case syscall.S_IFLNK:
		target, err := os.Readlink(name)
		if err != nil {
			log.Printf("Error reading symlink: %v", err)
		} else {
			atts.Targ = target
		}
	case syscall.S_IFIFO:
		basePerms(&atts, sys)
	case syscall.S_IFSOCK:
		basePerms(&atts, sys)
	case syscall.S_IFCHR:
		basePerms(&atts, sys)
		devInfo(&atts, sys)
		// TODO: These should have time info on them?
	case syscall.S_IFBLK:
		basePerms(&atts, sys)
		devInfo(&atts, sys)
	default:
		log.Printf("Node: %+v", info)
		panic("Unexpected file type")
	}

	return atts
}

// Base permissions shared by most nodes
func basePerms(atts *AttMap, sys *syscall.Stat_t) {
	atts.Uid = sys.Uid
	atts.Gid = sys.Gid
	atts.Perm = permission(sys)
}

func devInfo(atts *AttMap, sys *syscall.Stat_t) {
	atts.Devmaj = uint32(linuxdir.Major(sys.Rdev))
	atts.Devmin = uint32(linuxdir.Minor(sys.Rdev))
}

func timeInfo(atts *AttMap, sys *syscall.Stat_t) {
	// TODO: Store sub-second times.
	atts.Mtime = sys.Mtim.Sec
	atts.Ctime = sys.Ctim.Sec
}

// The Permission() call in 'os' masks off too many bits.
func permission(sys *syscall.Stat_t) uint32 {
	return sys.Mode &^ syscall.S_IFMT
}

func i64toa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func u64toa(i uint64) string {
	return strconv.FormatUint(i, 10)
}
