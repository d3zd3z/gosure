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

// NewWeave creates a new weave file.  The file will be named based on
// the given naming convention.  The 'name' will be used for the
// initial delta, and the tags will be recorded in that delta.  Close
// must be called to finialize the weaving.  Note that the underlying
// writer is buffered.
func NewNewWeave(nc NamingConvention, name string, tags map[string]string) (*NewWeaveWriter, error) {
	file, err := TempFile(nc, nc.IsCompressed())
	if err != nil {
		return nil, err
	}

	var w io.WriteCloser
	if nc.IsCompressed() {
		w = gzip.NewWriter(file)
	} else {
		w = closeWrite{bufio.NewWriter(file)}
	}

	head := NewHeader()
	delta := head.AddDelta(name, tags)

	err = head.Save(w)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}

	_, err = fmt.Fprintf(w, "\x01I %d\n", delta)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}

	return &NewWeaveWriter{
		file:  file,
		w:     w,
		nc:    nc,
		delta: delta,
	}, nil
}

func (w *NewWeaveWriter) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

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
