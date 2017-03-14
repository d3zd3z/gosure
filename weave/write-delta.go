package weave

import (
	"bufio"
	"fmt"
)

// A delta write writes the output of a single delta.
type DeltaWriter struct {
	write *bufio.Writer
}

func NewDeltaWriter(write *bufio.Writer) DeltaWriter {
	return DeltaWriter{
		write: write,
	}
}

func (w DeltaWriter) Insert(delta int) error { return nil }
func (w DeltaWriter) Delete(delta int) error { return nil }
func (w DeltaWriter) End(delta int) error    { return nil }

func (w DeltaWriter) Plain(text string, keep bool) error {
	if keep {
		_, err := fmt.Fprintln(w.write, text)
		return err
	}
	return nil
}
