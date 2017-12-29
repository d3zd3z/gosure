package sure

import (
	"bytes"
	"math/rand"
	"syscall"
)

// For regular file attributes, generate values and SHA1 that are
// valid random values.
func generateRegAtts(rand *rand.Rand) AttMap {
	var a RegAtts
	a.Uid = rand.Uint32()
	a.Gid = rand.Uint32()
	a.Perm = uint32(rand.Int31n(010000))
	a.Mtime = rand.Int63n(1e10)
	a.Ctime = rand.Int63n(1e10)
	a.Ino = uint64(rand.Int63())
	a.Size = rand.Int63()
	a.Sha1 = make([]byte, 20)
	rand.Read(a.Sha1)

	return &a
}

// Generate a random (but valid UTF-8) name.
func randomName(rand *rand.Rand) string {
	var buf bytes.Buffer

	limit := rand.Intn(255) + 1
	for i := 0; i < limit; i++ {
		// Most of the time, generate an ascii character
		// (possibly control).
		if rand.Intn(1000) > 0 {
			buf.WriteRune(rand.Int31n(96) + 32)
		} else {
			buf.WriteRune(rand.Int31n(0xd8000))
		}
	}

	return buf.String()
}

func generateLinkAtts(rand *rand.Rand) AttMap {
	var a LinkAtts
	a.Targ = randomName(rand)

	return &a
}

func generateFifoAtts(rand *rand.Rand) AttMap {
	var a FifoAtts
	a.Kind = syscall.S_IFIFO
	if rand.Intn(2) == 0 {
		a.Kind = syscall.S_IFSOCK
	}
	a.Uid = rand.Uint32()
	a.Gid = rand.Uint32()
	a.Perm = uint32(rand.Int31n(010000))

	return &a
}

func generateDevAtts(rand *rand.Rand) AttMap {
	var a DevAtts
	a.Kind = syscall.S_IFBLK
	if rand.Intn(2) == 0 {
		a.Kind = syscall.S_IFCHR
	}
	a.Uid = rand.Uint32()
	a.Gid = rand.Uint32()
	a.Perm = uint32(rand.Int31n(010000))
	a.Rdev = uint64(rand.Int63())

	return &a
}

func generateDirAtts(rand *rand.Rand) AttMap {
	var a DirAtts
	a.Uid = rand.Uint32()
	a.Gid = rand.Uint32()
	a.Perm = uint32(rand.Int31n(010000))

	return &a
}

func generateFile(rand *rand.Rand) *File {
	var f File
	f.Name = randomName(rand)

	switch rand.Intn(4) {
	case 0:
		f.Atts = generateRegAtts(rand)
	case 1:
		f.Atts = generateLinkAtts(rand)
	case 2:
		f.Atts = generateFifoAtts(rand)
	case 3:
		f.Atts = generateDevAtts(rand)
	}

	return &f
}

// Generate a random tree for testing.
func GenerateTree(rand *rand.Rand, size, depth int) *Tree {
	var t Tree
	t.Name = randomName(rand)
	t.Atts = generateDirAtts(rand)

	nchildren := rand.Intn(size) + 1
	if depth <= 1 {
		nchildren = 0
	}
	for ch := 0; ch < nchildren; ch++ {
		t.Children = append(t.Children, GenerateTree(rand, size, depth-1))
	}

	nchildren = rand.Intn(size) + 1
	for ch := 0; ch < nchildren; ch++ {
		t.Files = append(t.Files, generateFile(rand))
	}

	return &t
}
