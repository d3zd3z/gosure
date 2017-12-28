package weave

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"
)

// A Header is at the beginning of ever weave file.  It describes each
// of the deltas in the file.  The version describes the version of
// the header, in this case, 1.
type Header struct {
	Version int      `json:"version"`
	Deltas  []*Delta `json:"deltas"`
}

// A Delta describes a single version of the data stored in the weave
// file.  They are numbered, ideally sequentially, starting with 1,
// have a name, a timestamp, and a set of tags.
type Delta struct {
	Name   string            `json:"name"`
	Number int               `json:"number"`
	Tags   map[string]string `json:"tags"`
	Time   time.Time         `json:"time"`
}

// NewHeader creates a blank header describing zero deltas.
func NewHeader() Header {
	return Header{
		Version: 1,
	}
}

// AddDelta adds a new delta to this header, of a given name and tags.
// The tags will be copied, the Time filled in, and the number
// returned.
func (h *Header) AddDelta(name string, tags map[string]string) int {
	newTags := make(map[string]string)

	for k, v := range tags {
		newTags[k] = v
	}

	delta := Delta{
		Name:   name,
		Number: len(h.Deltas) + 1,
		Tags:   newTags,
		Time:   time.Now().UTC().Round(0),
	}
	h.Deltas = append(h.Deltas, &delta)

	return delta.Number
}

// Save writes the header to the stream in the format used in the
// weave files.
func (h *Header) Save(w io.Writer) error {
	_, err := w.Write([]byte("\x01t"))
	if err != nil {
		return err
	}

	// The json encoder adds a newline.
	enc := json.NewEncoder(w)
	return enc.Encode(h)
}

// LatestDelta returns the most recent delta in the store.  Zero means
// there are no deltas (this is valid for a new blank header).
func (h *Header) LatestDelta() int {
	max := 0
	for _, d := range h.Deltas {
		if d.Number > max {
			max = d.Number
		}
	}
	return max
}

// PenultimateDelta returns the penultimate delta, if there is one.
// Will return an error if this it not the case.
func (h *Header) PenultimateDelta() (int, error) {
	var deltas []int

	for _, d := range h.Deltas {
		deltas = append(deltas, d.Number)
	}

	if len(deltas) < 2 {
		return 0, errors.New("Cannot load previous version, there is only one")
	}

	sort.Sort(sort.Reverse(sort.IntSlice(deltas)))

	return deltas[1], nil
}

// ErrInvalidHeader indicates that the header in the weave file was
// malformed.  This usually means this isn't actually a weave file.
var ErrInvalidHeader = fmt.Errorf("invalid weave header")

// A BytesReader is something that can be read a line at a time.
type BytesReader interface {
	ReadBytes(delim byte) ([]byte, error)
}

// LoadHeader reads the header from the stream.
func LoadHeader(r BytesReader) (*Header, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	if len(line) < 2 || line[0] != 1 || line[1] != 't' {
		return nil, ErrInvalidHeader
	}

	var header Header
	err = json.Unmarshal(line[2:], &header)
	if err != nil {
		return nil, err
	}

	return &header, nil
}
