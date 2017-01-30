package sure

import (
	"bytes"
	"fmt"
	"path"
	"sort"
	"strings"
)

// Traverse an old tree and a new tree, printing out everything that
// is different between them.
func CompareTrees(older, newer *Tree) {
	compWalk(older, newer, ".")
}

func compWalk(older, newer *Tree, name string) {
	// First make a map of the old ones.
	oldc := make(map[string]*Tree)

	for _, och := range older.Children {
		oldc[och.Name] = och
	}

	for _, nch := range newer.Children {
		och, ok := oldc[nch.Name]
		chname := path.Join(name, nch.Name)
		if ok {
			// Recursively compare the children.
			compWalk(och, nch, chname)

			// Compare the attributes.
			compAtts(chname, och.Atts, nch.Atts)

			// Remove this from the old list so we don't
			// show it as deleted.
			delete(oldc, och.Name)
		} else {
			// Not present in old, this names a new
			// directory.
			fmt.Printf("+ %-22s %s\n", "dir", chname)
		}
	}

	// Any remaining directories have been removed.
	oldNames := make([]string, 0, len(oldc))
	for n, _ := range oldc {
		oldNames = append(oldNames, n)
	}
	sort.Sort(sort.StringSlice(oldNames))

	for _, subname := range oldNames {
		chname := path.Join(name, subname)
		fmt.Printf("- %-22s %s\n", "dir", chname)
	}

	compFiles(older.Files, newer.Files, name)
}

func compFiles(older, newer []*File, name string) {
	oldf := make(map[string]*File)

	for _, ofi := range older {
		oldf[ofi.Name] = ofi
	}

	for _, nfi := range newer {
		ofi, ok := oldf[nfi.Name]
		chname := path.Join(name, nfi.Name)
		if ok {
			compAtts(chname, ofi.Atts, nfi.Atts)
			delete(oldf, ofi.Name)
		} else {
			fmt.Printf("+ %-22s %s\n", "file", chname)
		}
	}

	// Any remaining files have been removed.
	oldNames := make([]string, 0, len(oldf))
	for n, _ := range oldf {
		oldNames = append(oldNames, n)
	}
	sort.Sort(sort.StringSlice(oldNames))

	for _, subname := range oldNames {
		chname := path.Join(name, subname)
		fmt.Printf("- %-22s %s\n", "file", chname)
	}
}

// Compare attributes, and if any differ, print them out and the file
// name.  Ignores attributes "ctime" and "ino" because these will not
// be the same when restored from a backup.
func compAtts(name string, oa, na AttMap) {
	var mismatch []string

	if oa.Devmaj != na.Devmaj || oa.Devmin != na.Devmin {
		mismatch = append(mismatch, "dev")
	}
	if oa.Gid != na.Gid {
		mismatch = append(mismatch, "gid")
	}
	if oa.Kind != na.Kind {
		mismatch = append(mismatch, "kind")
	}
	if oa.Mtime != na.Mtime {
		mismatch = append(mismatch, "mtime")
	}
	if oa.Perm != na.Perm {
		mismatch = append(mismatch, "perm")
	}
	if !bytes.Equal(oa.Sha, na.Sha) {
		mismatch = append(mismatch, "sha1")
	}
	if oa.Size != na.Size {
		mismatch = append(mismatch, "size")
	}
	if oa.Targ != na.Targ {
		mismatch = append(mismatch, "targ")
	}
	if oa.Uid != na.Uid {
		mismatch = append(mismatch, "uid")
	}

	if len(mismatch) == 0 {
		return
	}

	attText := strings.Join(mismatch, ",")
	fmt.Printf("  [%-20s] %s\n", attText, name)
}

func warnAtt(key, kind string) {
}
