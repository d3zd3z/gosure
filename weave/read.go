package weave

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
)

// ReadHeader reads the header from the weave file described by the
// naming convention.
func ReadHeader(nc NamingConvention) (*Header, error) {
	file, err := os.Open(nc.MainFile())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rd BytesReader
	if nc.IsCompressed() {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		rd = bufio.NewReader(gz)
	} else {
		rd = bufio.NewReader(file)
	}

	return LoadHeader(rd)
}

// ReadDelta reads the contents of a given delta from a delta file.
// The 'line' function will be called on each line of text from the
// delta.  If 'line' returns a non-nil error, it will be propagated up
// through the call to ReadDelta, otherwise, ReadDelta will return any
// error encountered in reading.
func ReadDelta(nc NamingConvention, delta int, line func(text string) error) error {
	return ReadGeneral(nc, delta, deltaSink(line))
}

// GeneralReader reads a delta using the specified Sink.
func ReadGeneral(nc NamingConvention, delta int, sink Sink) error {
	file, err := os.Open(nc.MainFile())
	if err != nil {
		return err
	}
	defer file.Close()

	var rd io.Reader
	if nc.IsCompressed() {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		rd = gz
	} else {
		// The Parser will already wrap in a bufio.Reader, so no need
		// to wrap here.
		rd = file
	}

	p := NewParser(rd, sink, delta)
	return p.ParseTo(0)
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
