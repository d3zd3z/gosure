package weave

import (
	"bufio"
)

// WriteDelta is a Sink that merely writes the plaintext of a given
// delta to the given writer.  Does not close the writer when it is
// finished.
type WriteDelta struct {
	write *bufio.Writer
}

// NewWriteDelta creates a WriteDelta that just writes the plaintext
// to the given writer.
func NewWriteDelta(wr *bufio.Writer) WriteDelta {
	return WriteDelta{
		write: wr,
	}
}

// Insert marks an insert, does nothing.
func (w WriteDelta) Insert(delta int) error { return nil }

// Delete marks a delete, does nothing.
func (w WriteDelta) Delete(delta int) error { return nil }

// End marks the end, does nothing.
func (w WriteDelta) End(delta int) error { return nil }

// Plain possibly adds plaintext.
func (w WriteDelta) Plain(text string, keep bool) error {
	if keep {
		_, err := w.write.WriteString(text)
		if err != nil {
			return err
		}

		return w.write.WriteByte('\n')
	}
	return nil
}
