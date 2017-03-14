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

func NewWriteDelta(wr *bufio.Writer) WriteDelta {
	return WriteDelta{
		write: wr,
	}
}

func (w WriteDelta) Insert(delta int) error { return nil }
func (w WriteDelta) Delete(delta int) error { return nil }
func (w WriteDelta) End(delta int) error    { return nil }

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
