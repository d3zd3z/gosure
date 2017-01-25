package sure

import (
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
//
// There are some challenges.  It is possible that the old and new
// scans were made with different versions of this program, or even
// with different programs entirely, and therefore may have different
// attributes available.  We want to warn the user about these
// scenarios, but not otherwise report them for each file.  The warn
// is done separately to track this.
func compAtts(name string, oa, na AttMap) {
	allAttMap := make(map[string]bool)

	for k, _ := range oa {
		allAttMap[k] = true
	}
	for k, _ := range na {
		allAttMap[k] = true
	}

	delete(allAttMap, "ctime")
	delete(allAttMap, "ino")

	// Sort the whole attribute list.
	allAtts := make([]string, 0, len(allAttMap))
	for k, _ := range allAttMap {
		allAtts = append(allAtts, k)
	}
	sort.Sort(sort.StringSlice(allAtts))

	var mismatch []string

	for _, k := range allAtts {
		ov, ook := oa[k]
		nv, nok := na[k]
		switch {
		case ook && !nok:
			warnAtt(k, "removed")
		case !ook && nok:
			warnAtt(k, "added")
		default:
			if ov != nv {
				mismatch = append(mismatch, k)
			}
		}
	}

	if len(mismatch) == 0 {
		return
	}

	attText := strings.Join(mismatch, ",")
	fmt.Printf("  [%-20s] %s\n", attText, name)
}

func warnAtt(key, kind string) {
}
