package sure

import (
	"bytes"
	"fmt"
	"log"
	"path"
	"reflect"
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

	ov := reflect.ValueOf(oa).Elem()
	nv := reflect.ValueOf(na).Elem()

	if ov.Type() != nv.Type() {
		mismatch = append(mismatch, "kind")
	} else {
		mismatch = compAttWalk(ov, nv, nil)
	}

	if len(mismatch) == 0 {
		return
	}

	sort.Sort(sort.StringSlice(mismatch))

	attText := strings.Join(mismatch, ",")
	fmt.Printf("  [%-20s] %s\n", attText, name)
}

// Walk through the structures (which are assumed to be the same
// type), and compare all of the items.
func compAttWalk(ov, nv reflect.Value, mismatch []string) []string {
	t := ov.Type()
	nField := t.NumField()

	for i := 0; i < nField; i++ {
		ofld := ov.Field(i)
		nfld := nv.Field(i)
		ftyp := t.Field(i)

		// Walk down the struct
		if ftyp.Type.Kind() == reflect.Struct {
			mismatch = compAttWalk(ofld, nfld, mismatch)
			continue
		}

		name := strings.ToLower(ftyp.Name)

		// Special case to ignore ctime and ino.
		if name == "ctime" || name == "ino" {
			continue
		}

		bad := false

		// Type based comparison.
		switch ofld.Interface().(type) {
		case uint32:
			if ofld.Uint() != nfld.Uint() {
				bad = true
			}
		case uint64:
			if ofld.Uint() != nfld.Uint() {
				bad = true
			}
		case int64:
			if ofld.Int() != nfld.Int() {
				bad = true
			}
		case []byte:
			if bytes.Compare(ofld.Bytes(), nfld.Bytes()) != 0 {
				bad = true
			}
		case string:
			if ofld.String() != nfld.String() {
				bad = true
			}
		default:
			log.Fatalf("unknown field type: %v", ftyp)
		}

		if bad {
			mismatch = append(mismatch, strings.ToLower(ftyp.Name))
		}
	}

	return mismatch
}

func warnAtt(key, kind string) {
}
