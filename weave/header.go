package weave

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Header struct {
	Version int      `json:"version"`
	Deltas  []*Delta `json:"deltas"`
}

type Delta struct {
	Name   string            `json:"name"`
	Number int               `json:"number"`
	Tags   map[string]string `json:"tags"`
	Time   string            `json:"time"`
}

func NewHeader() Header {
	return Header{
		Version: 1,
	}
}

// Add a new delta to this header, of a given name and tags.  The tags
// will be copied, the Time filled in, and the number returned.
func (h *Header) AddDelta(name string, tags map[string]string) int {
	newTags := make(map[string]string)

	for k, v := range tags {
		newTags[k] = v
	}

	delta := Delta{
		Name:   name,
		Number: len(h.Deltas) + 1,
		Tags:   newTags,
		Time:   time.Now().UTC().Format(time.RFC3339Nano),
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

var InvalidHeader = fmt.Errorf("Invalid weave header")

// A BytesReader is something that can be read a line at a time.
type BytesReader interface {
	ReadBytes(delim byte) ([]byte, error)
}

// Load reads the header from the stream.
func LoadHeader(r BytesReader) (*Header, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	if len(line) < 2 || line[0] != 1 || line[1] != 't' {
		return nil, InvalidHeader
	}

	var header Header
	err = json.Unmarshal(line[2:], &header)
	if err != nil {
		return nil, err
	}

	return &header, nil
}
