package weave

import (
	"fmt"
	"os"
	"strconv"
)

// A NamingConvention determines the names of various temp files.  The
// SCCS conventions are not followed, because they are not safe (this
// code will never write to a file that already exists).
type NamingConvention interface {
	// Generate a temporary name.  The file may or may not exist
	// (not checked).  If the name is already present, this can be
	// called again with a different num to generate a different
	// name.  Compressed is a hint as to whether this name should
	// match compression.
	TempFile(num int, compressed bool) string

	// Return the pathname of the primary file.
	MainFile() string

	// Return the pathname of the backup file.
	BackupFile() string

	// Return if compression is requested on the main file.
	IsCompressed() bool
}

// The SimpleNaming is a NamingConvention that has a basename, with
// the main file having a specified extension, the backup file having
// a ".bak" extension, and the temp files using a numbered extension
// starting with ".0".  If the names are intended to be compressed, a
// ".gz" suffix can also be added.
type SimpleNaming struct {
	Path       string // The directory for the files to be written.
	Base       string // The base of the filename.
	Ext        string // The extension to use for the main name.
	Compressed bool   // Are these names to indicate compression.
}

func (sn *SimpleNaming) MakeName(ext string, compressed bool) string {
	gz := ""
	if sn.Compressed && compressed {
		gz = ".gz"
	}
	return fmt.Sprintf("%s/%s.%s%s", sn.Path, sn.Base, ext, gz)
}

func (sn *SimpleNaming) MainFile() string {
	return sn.MakeName(sn.Ext, true)
}

func (sn *SimpleNaming) BackupFile() string {
	return sn.MakeName("bak", true)
}

func (sn *SimpleNaming) TempFile(num int, compressed bool) string {
	return sn.MakeName(strconv.Itoa(num), compressed)
}

// MakeTemp will create a tempfile, the file, and an error.
func TempFile(nc NamingConvention, compressed bool) (file *os.File, err error) {
	n := 0
	for {
		name := nc.TempFile(n, compressed)

		file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err == nil {
			return file, nil
		}

		// Only continue if the error we get is because the
		// file already exists.  Any other error is returned.
		if !os.IsExist(err) {
			return nil, err
		}

		n += 1
	}
}

func (sn *SimpleNaming) IsCompressed() bool {
	return sn.Compressed
}
