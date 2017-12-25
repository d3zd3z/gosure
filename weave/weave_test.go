package weave_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"testing"

	"davidb.org/x/gosure/weave"
)

func TestDeltas(t *testing.T) {
	tdir, err := ioutil.TempDir("", "weave-test-")
	if err != nil {
		t.Fatal(err)
	}
	if true {
		defer os.RemoveAll(tdir)
	}
	// fmt.Printf("tmp: %s\n", tdir)

	data := NewDataSet(tdir, 100)

	err = data.SaveNew()
	if err != nil {
		t.Fatal(err)
	}

	err = data.Check(1)
	if err != nil {
		t.Fatal(err)
	}

	for i := 2; i < 100; i++ {
		data.Scramble()
		err = data.SaveDelta()
		if err != nil {
			t.Fatal(err)
		}

		err = data.Check(i)
		if err != nil {
			t.Fatal(err)
		}
	}
}

type DataSet struct {
	Data   []int
	Name   string
	Tags   map[string]string
	NC     weave.SimpleNaming
	Deltas map[int][]int
}

func NewDataSet(dir string, limit int) *DataSet {
	nums := make([]int, limit)
	for i := 0; i < limit; i++ {
		nums[i] = i
	}
	tags := map[string]string{
		"kind": "initial",
	}

	return &DataSet{
		Data: nums,
		Name: "initial",
		Tags: tags,
		NC: weave.SimpleNaming{
			Path:       dir,
			Base:       "test",
			Ext:        "weave",
			Compressed: true,
		},
		Deltas: make(map[int][]int),
	}
}

func (d *DataSet) SaveNew() error {
	wr, err := weave.NewNewWeave(&d.NC, d.Name, d.Tags)
	if err != nil {
		return err
	}
	// Can't be deferred, because we need it closed to be able to
	// read the header back.
	// defer wr.Close()

	err = d.SaveData(wr)
	wr.Close()
	if err != nil {
		return err
	}

	// Verify the header.
	hd, err := weave.ReadHeader(&d.NC)
	if err != nil {
		return err
	}

	if len(hd.Deltas) != 1 {
		return fmt.Errorf("Expecting 1 delta not %d", len(hd.Deltas))
	}

	if hd.Deltas[0].Name != d.Name {
		return fmt.Errorf("Delta name doesn't match %q != %q", hd.Deltas[0].Name, d.Name)
	}

	if !reflect.DeepEqual(hd.Deltas[0].Tags, d.Tags) {
		return fmt.Errorf("Tags don't match %v != %v", hd.Deltas[0].Tags, d.Tags)
	}

	newNums := make([]int, len(d.Data))
	copy(newNums, d.Data)
	d.Deltas[1] = newNums

	return nil
}

func (d *DataSet) SaveDelta() error {
	base := len(d.Deltas)
	wr, err := weave.NewDeltaWriter(&d.NC, base, d.Name, d.Tags)
	if err != nil {
		return err
	}
	err = d.SaveData(wr)
	err2 := wr.Close()
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}

	max := 0
	for k := range d.Deltas {
		if k > max {
			max = k
		}
	}

	newNums := make([]int, len(d.Data))
	copy(newNums, d.Data)
	d.Deltas[max+1] = newNums

	return nil
}

// Check the given delta against our notion of that delta.
func (d *DataSet) Check(delta int) error {
	expect, ok := d.Deltas[delta]
	if !ok {
		panic("Trying to check delta that hasn't been written")
	}

	nums := make([]int, 0)
	err := weave.ReadDelta(&d.NC, delta, func(line string) error {
		num, err := strconv.Atoi(line)
		if err != nil {
			return err
		}
		nums = append(nums, num)
		return nil
	})
	if err == nil {
		return fmt.Errorf("Unexpected empty return from ReadDelta")
	}
	if err != io.EOF {
		return err
	}

	if !reflect.DeepEqual(nums, expect) {
		return fmt.Errorf("Mismatch reading nums:\ngot: %v\nExpect: %v\n", nums, expect)
	}

	return nil
}

// Modify some of the data.  For now, just replace a sequence with
// some new values.
func (d *DataSet) Scramble() {
	max := 0
	for _, n := range d.Data {
		if n > max {
			max = n
		}
	}

	a := rand.Intn(len(d.Data))
	b := rand.Intn(len(d.Data))
	if a > b {
		a, b = b, a
	}

	for i := a; i <= b; i++ {
		max++
		d.Data[i] = max
	}
}

func (d *DataSet) SaveData(w io.Writer) error {
	for _, num := range d.Data {
		_, err := fmt.Fprintf(w, "%d\n", num)
		if err != nil {
			return err
		}
	}
	return nil
}
