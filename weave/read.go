package weave

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
)

// weaveOpen opens a weave file for reading, based on a given naming
// convention.  On success, returns the opened File, a reader on that
// file (which will either be the same file, or a gzip.Reader), and
// nil.  Otherwise, an error is returned
func weaveOpen(nc NamingConvention) (*os.File, io.Reader, error) {
	file, err := os.Open(nc.MainFile())
	if err != nil {
		return nil, nil, err
	}

	if nc.IsCompressed() {
		gz, err := gzip.NewReader(file)
		if err != nil {
			file.Close()
			return nil, nil, err
		}
		return file, gz, nil
	}
	return file, file, nil
}

// ReadHeader reads the header from the weave file described by the
// naming convention.
func ReadHeader(nc NamingConvention) (*Header, error) {
	file, rd, err := weaveOpen(nc)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bufrd := bufio.NewReader(rd)

	return LoadHeader(bufrd)
}

// ReadDelta reads the contents of a given delta from a delta file.
// The 'line' function will be called on each line of text from the
// delta.  If 'line' returns a non-nil error, it will be propagated up
// through the call to ReadDelta, otherwise, ReadDelta will return any
// error encountered in reading.
func ReadDelta(nc NamingConvention, delta int, line func(text string) error) error {
	return ReadGeneral(nc, delta, deltaSink(line))
}

// ReadGeneral reads a delta using the specified Sink.
func ReadGeneral(nc NamingConvention, delta int, sink Sink) error {
	file, rd, err := weaveOpen(nc)
	if err != nil {
		return err
	}
	defer file.Close()

	// The Parser will already wrap in a bufio.Reader, so no need
	// to wrap here.
	return NewParser(rd, sink, delta).ParseTo(0)
}

type deltaSink func(text string) error

func (s deltaSink) Insert(delta int) error { return nil }
func (s deltaSink) Delete(delta int) error { return nil }
func (s deltaSink) End(delta int) error    { return nil }

func (s deltaSink) Plain(text string, keep bool) error {
	if keep {
		return s(text)
	}
	return nil
}
