package store_test

import (
	"testing"

	"davidb.org/x/gosure/store"
)

var parseTests = []struct {
	name  string
	err   error
	store store.Store
}{
	{"/invalid/path", store.NotDir("/invalid/path"), store.Store{
		"", "", "", false, simpleTags(), "",
	}},
	{"/usr/bin", nil, store.Store{
		"/usr/bin", "", "", false, simpleTags(), "",
	}},
	{"/usr/bin/2sure.dat.gz", nil, store.Store{
		"/usr/bin", "2sure", "", false, simpleTags(), "",
	}},
	{"/usr/bin/2sure.bak.gz", nil, store.Store{
		"/usr/bin", "2sure", "", false, simpleTags(), "",
	}},
	{"/usr/bin/2sure.dat", nil, store.Store{
		"/usr/bin", "2sure", "", true, simpleTags(), "",
	}},
	{"/usr/bin/2sure.bak", nil, store.Store{
		"/usr/bin", "2sure", "", true, simpleTags(), "",
	}},
	{"/usr/bin/fred.dat.gz", nil, store.Store{
		"/usr/bin", "fred", "", false, simpleTags(), "",
	}},
	{"/usr/bin/fred.dat", nil, store.Store{
		"/usr/bin", "fred", "", true, simpleTags(), "",
	}},
	{"/usr/bin/fred", nil, store.Store{
		"/usr/bin", "fred", "", false, simpleTags(), "",
	}},
	{"/usr/bin/bogus.ext", store.InvalidName("/usr/bin/bogus.ext"), store.Store{
		"", "", "", false, simpleTags(), "",
	}},
	{"/usr/bin/stuff.weave.gz", nil, store.Store{
		"/usr/bin", "stuff", "weave", false, simpleTags(), "",
	}},
	{"/usr/bin/stuff.weave", nil, store.Store{
		"/usr/bin", "stuff", "weave", true, simpleTags(), "",
	}},
}

// Construct a simple tag with just a name.
func simpleTags() map[string]string {
	result := make(map[string]string)
	result["name"] = "test"
	return result
}

func TestParse(t *testing.T) {
	for _, pt := range parseTests {
		var st store.Store
		err := st.Parse(pt.name)
		if err != pt.err {
			t.Fatalf("Unexpected error, got %v, expect %v", err, pt.err)
		}

		if st.Path != pt.store.Path || st.Base != pt.store.Base || st.Plain != pt.store.Plain {
			t.Fatalf("Store mismatch, got %+v, expect %+v", st, pt.store)
		}
	}
}
