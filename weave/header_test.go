package weave_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"davidb.org/x/gosure/weave"
)

func TestHeader(t *testing.T) {
	h := weave.NewHeader()

	tags := map[string]string{
		"key": "value 1",
	}

	for rev := 1; rev < 10; rev++ {
		// Add a tag to make sure the tags are not being
		// shared.
		tags["key"] = fmt.Sprintf("value %d", rev)
		tags[fmt.Sprintf("key%d", rev)] = fmt.Sprintf("value %d", rev)
		num := h.AddDelta(fmt.Sprintf("Sample %d", rev), tags)
		delete(tags, fmt.Sprintf("key%d", rev))
		if num != rev {
			t.Fatalf("Unexpected number %d != %d", num, rev)
		}
	}

	var buf bytes.Buffer
	err := h.Save(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// Add a magic token to the end of the buffer to make sure it
	// doesn't get read.
	_, err = buf.WriteString("\x01E\n")

	hdr2, err := weave.LoadHeader(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure that it only read one line.
	rest := buf.String()
	if rest != "\x01E\n" {
		t.Fatalf("Didn't find expected remainder: %q != %q", rest, "\x01E\n")
	}

	if !reflect.DeepEqual(&h, hdr2) {
		t.Fatalf("Header did not read back correctly")
	}
}
