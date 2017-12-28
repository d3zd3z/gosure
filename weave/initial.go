package weave

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

// closeWrite wraps bufio.Writer and adds a Close method so that it
// can also be a io.WriteCloser and be used in place of a gzip.Writer,
// which requires Close to be called for the gzipped file to be valid.
type closeWrite struct {
	*bufio.Writer
}

func (cl closeWrite) Close() error {
	return cl.Flush()
}

// NewWeaveWriter implements io.Writer (buffered) to write to a new
// weave file.  The client should call the Close method to finalize
// writing.
type NewWeaveWriter struct {
	file  *os.File       // The temp file where the data is written.
	w     io.WriteCloser // Writer, buffered possibly with compression.
	nc    NamingConvention
	delta int
}

// weaveCreate opens a new weave file for writing, writing the given
// header to the file, and returning the file, and a bufferfed writer
// (possibly with compression).
func weaveCreate(nc NamingConvention, head *Header) (*os.File, io.WriteCloser, error) {
	file, err := TempFile(nc, nc.IsCompressed())
	if err != nil {
		return nil, nil, err
	}

	var wr io.WriteCloser
	if nc.IsCompressed() {
		wr = gzip.NewWriter(file)
	} else {
		wr = closeWrite{bufio.NewWriter(file)}
	}

	err = head.Save(wr)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, nil, err
	}

	return file, wr, err
}

// NewNewWeave creates a new weave file.  The file will be named based on
// the given naming convention.  The 'name' will be used for the
// initial delta, and the tags will be recorded in that delta.  Close
// must be called to finialize the weaving.  Note that the underlying
// writer is buffered.
func NewNewWeave(nc NamingConvention, name string, tags map[string]string) (*NewWeaveWriter, error) {
	head := NewHeader()
	delta := head.AddDelta(name, tags)

	file, wr, err := weaveCreate(nc, &head)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(wr, "\x01I %d\n", delta)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}

	return &NewWeaveWriter{
		file:  file,
		w:     wr,
		nc:    nc,
		delta: delta,
	}, nil
}

// writeWeave opens a temp weave file for writing,

func (w *NewWeaveWriter) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

// Close closes the NewWeaveWriter, renaming the new file to the
// finalized name.
func (w *NewWeaveWriter) Close() error {
	_, err := fmt.Fprintf(w, "\x01E %d\n", w.delta)
	if err != nil {
		return err
	}

	err = w.w.Close()
	if err != nil {
		return err
	}

	err = w.file.Close()
	if err != nil {
		return err
	}

	os.Rename(w.nc.MainFile(), w.nc.BackupFile())
	return os.Rename(w.file.Name(), w.nc.MainFile())
}
