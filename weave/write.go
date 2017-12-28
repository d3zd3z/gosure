package weave

import (
	"fmt"
	"io"
)

// A Writer is a sync that outputs weave data to the given output.
// The particular delta is ignored, and the entire output will be
// emitted.
type Writer struct {
	io.Writer
}

// Insert marks a new insert.
func (w Writer) Insert(delta int) error {
	_, err := fmt.Fprintf(w, "\x01I %d\n", delta)
	return err
}

// Delete marks a new delete.
func (w Writer) Delete(delta int) error {
	_, err := fmt.Fprintf(w, "\x01D %d\n", delta)
	return err
}

// End marks a new end.
func (w Writer) End(delta int) error {
	_, err := fmt.Fprintf(w, "\x01E %d\n", delta)
	return err
}

// Plain marks a new line of plaintext.
func (w Writer) Plain(text string, keep bool) error {
	_, err := fmt.Fprintln(w, text)
	return err
}
