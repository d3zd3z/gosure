package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"davidb.org/code/gosure/weave"
)

var deltaLines int
var deltaDeltas int
var weaveFile string = "w.delta"

func delta() {
	if _, err := os.Lstat(playFile); err == nil {
		log.Fatal("plan file %q is present, remove and rerun", playFile)
	}

	// Clean up prior runs.
	os.Remove(sccsFile)

	data := make([]int, deltaLines)

	for i := range data {
		data[i] = i + 1
	}

	// Generate initial SCCS file.
	ss := NewSccs(playFile)
	if err := ss.Initial(data); err != nil {
		log.Fatal(err)
	}

	// Generate initial weave file.
	if err := writeWeave(data); err != nil {
		log.Fatal(err)
	}

	// Make sure they are the same.
	if err := compareWeaves(sccsFile, weaveFile); err != nil {
		log.Fatal(err)
	}
}

// Write an initial weave file, writing the data as the first delta.
func writeWeave(data []int) error {
	f, err := os.Create(weaveFile)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	defer buf.Flush()

	wr := weave.NewWriter(buf)

	if err = wr.Insert(1); err != nil {
		return err
	}

	for _, x := range data {
		if err = wr.Plain(fmt.Sprintf("%d", x), true); err != nil {
			return nil
		}
	}

	if err = wr.End(1); err != nil {
		return err
	}

	return nil
}

func compareWeaves(a, b string) error {
	af, err := os.Open(a)
	if err != nil {
		return err
	}
	defer af.Close()

	bf, err := os.Open(b)
	if err != nil {
		return err
	}
	defer bf.Close()

	achan := make(chan controlItem, 0)
	bchan := make(chan controlItem, 0)

	aparse := weave.NewParser(af, chanWriter(achan), 0)
	bparse := weave.NewParser(bf, chanWriter(bchan), 0)

	go func() {
		err := aparse.ParseTo(0)
		if err != io.EOF {
			log.Fatal(err, " achan")
		}
		close(achan)
	}()

	go func() {
		err := bparse.ParseTo(0)
		if err != io.EOF {
			log.Fatal(err, " bchan")
		}
		close(bchan)
	}()

	for {
		aitem, aok := <-achan
		bitem, bok := <-bchan

		// Both are finished.
		if !aok && !bok {
			break
		}

		if aok != bok {
			return errors.New("Early EOF")
		}

		if aitem.kind != bitem.kind || aitem.delta != bitem.delta || aitem.text != bitem.text {
			return errors.New("Item mismatch")
		}
	}

	return nil
}

// A ChanWriter writes the encoded data to a channel.
type controlType int

const (
	controlInsert controlType = iota
	controlDelete
	controlEnd
	controlPlain
)

// Either delta or text should be set to the zero value.
type controlItem struct {
	kind  controlType
	delta int
	text  string
}

type chanWriter chan<- controlItem

func (w chanWriter) Insert(delta int) error {
	w <- controlItem{
		kind:  controlInsert,
		delta: delta,
	}
	return nil
}

func (w chanWriter) Delete(delta int) error {
	w <- controlItem{
		kind:  controlDelete,
		delta: delta,
	}
	return nil
}

func (w chanWriter) End(delta int) error {
	w <- controlItem{
		kind:  controlEnd,
		delta: delta,
	}
	return nil
}

func (w chanWriter) Plain(text string, keep bool) error {
	w <- controlItem{
		kind: controlPlain,
		text: text,
	}
	return nil
}
