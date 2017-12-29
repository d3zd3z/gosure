package sure_test

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"syscall"
	"testing"

	"davidb.org/x/gosure/sure"
)

var generate = flag.Bool("generate", false, "Generate test tree save")

func TestEncode(t *testing.T) {
	// Marshal the tree.
	var buf bytes.Buffer
	err := tdata1.Encode(&buf)
	if err != nil {
		t.Fatal(err)
	}

	if *generate {
		f, err := os.Create("testdata/tdata1.golden")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		err = tdata1.Encode(f)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		b2, err := ioutil.ReadFile("testdata/tdata1.golden")
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(buf.Bytes(), b2) != 0 {
			t.Fatal("Mismatch on marshalled data.")
		}
	}
}

var tdata1 *sure.Tree = &sure.Tree{
	Name: "__root__",
	Atts: &sure.DirAtts{
		BaseAtts: sure.BaseAtts{
			Uid:  12345,
			Gid:  54321,
			Perm: 0755,
		},
	},
	Children: []*sure.Tree{},
	Files: []*sure.File{
		&sure.File{
			Name: "regular file",
			Atts: &sure.RegAtts{
				BaseAtts: sure.BaseAtts{
					Uid:  0xffffffff,
					Gid:  0x80000000,
					Perm: 0644,
				},
				Mtime: 1485993509,
				Ctime: 1485993510,
				Ino:   0x123456789abcdef0,
				Size:  5827423,
				Sha1: []byte{
					0x8f, 0x55, 0x2e, 0x8f, 0x26,
					0x4d, 0x2a, 0x9c, 0x61, 0x9b,
					0xfc, 0xaa, 0x1f, 0x87, 0xf7,
					0xb0, 0x6c, 0x43, 0x45, 0x82,
				},
			},
		},
		&sure.File{
			Name: "a symlink\xff", // Invalid UTF-8.
			Atts: &sure.LinkAtts{
				Targ: genBytes(),
			},
		},
		&sure.File{
			Name: "A fifo",
			Atts: &sure.FifoAtts{
				Kind: syscall.S_IFIFO,
				BaseAtts: sure.BaseAtts{
					Uid:  52,
					Gid:  74,
					Perm: 12345,
				},
			},
		},
		&sure.File{
			Name: "A socket",
			Atts: &sure.FifoAtts{
				Kind: syscall.S_IFSOCK,
				BaseAtts: sure.BaseAtts{
					Uid:  7194783,
					Gid:  719648,
					Perm: 71964873,
				},
			},
		},
		&sure.File{
			Name: "A block file",
			Atts: &sure.DevAtts{
				Kind: syscall.S_IFBLK,
				BaseAtts: sure.BaseAtts{
					Uid:  174872,
					Gid:  9174863,
					Perm: 9717483,
				},
				Rdev: 82749,
			},
		},
		&sure.File{
			Name: "A character file",
			Atts: &sure.DevAtts{
				Kind: syscall.S_IFCHR,
				BaseAtts: sure.BaseAtts{
					Uid:  9817497,
					Gid:  7196487,
					Perm: 971964,
				},
				Rdev: 987461,
			},
		},
	},
}

func genBytes() string {
	buf := make([]byte, 256)
	for i := 0; i < 256; i++ {
		buf[i] = byte(i)
	}

	return string(buf)
}
