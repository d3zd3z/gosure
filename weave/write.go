package weave

import (
	"bufio"
	"fmt"
)

// A Writer is a sync that outputs weave data to the given output.
// The particular delta is ignored, and the entire output will be
// emitted.
type Writer struct {
	write *bufio.Writer
}

func NewWriter(write *bufio.Writer) Writer {
	return Writer{
		write: write,
	}
}

func (w Writer) Insert(delta int) error {
	_, err := fmt.Fprintf(w.write, "\x01I %d\n", delta)
	return err
}

func (w Writer) Delete(delta int) error {
	_, err := fmt.Fprintf(w.write, "\x01D %d\n", delta)
	return err
}

func (w Writer) End(delta int) error {
	_, err := fmt.Fprintf(w.write, "\x01E %d\n", delta)
	return err
}

func (w Writer) Plain(text string, keep bool) error {
	_, err := fmt.Fprintln(w.write, text)
	return err
}
