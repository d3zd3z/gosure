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
		"", "", false,
	}},
	{"/usr/bin", nil, store.Store{
		"/usr/bin", "", false,
	}},
	{"/usr/bin/2sure.dat.gz", nil, store.Store{
		"/usr/bin", "2sure", false,
	}},
	{"/usr/bin/2sure.bak.gz", nil, store.Store{
		"/usr/bin", "2sure", false,
	}},
	{"/usr/bin/2sure.dat", nil, store.Store{
		"/usr/bin", "2sure", true,
	}},
	{"/usr/bin/2sure.bak", nil, store.Store{
		"/usr/bin", "2sure", true,
	}},
	{"/usr/bin/fred.dat.gz", nil, store.Store{
		"/usr/bin", "fred", false,
	}},
	{"/usr/bin/fred.dat", nil, store.Store{
		"/usr/bin", "fred", true,
	}},
	{"/usr/bin/fred", nil, store.Store{
		"/usr/bin", "fred", false,
	}},
	{"/usr/bin/bogus.ext", store.InvalidName("/usr/bin/bogus.ext"), store.Store{
		"", "", false,
	}},
}

func TestParse(t *testing.T) {
	for _, pt := range parseTests {
		var st store.Store
		err := st.Parse(pt.name)
		if err != pt.err {
			t.Fatalf("Unexpected error, got %v, expect %v", err, pt.err)
		}

		if st != pt.store {
			t.Fatalf("Store mismatch, got %+v, expect %+v", st, pt.store)
		}
	}
}
