package weave_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"testing"

	"davidb.org/x/gosure/weave"
)

func TestDeltas(t *testing.T) {
	data := NewDataSet(100)

	tdir, err := ioutil.TempDir("", "weave-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	sname := weave.SimpleNaming{
		Path:       tdir,
		Base:       "test",
		Ext:        "weave",
		Compressed: true,
	}

	wr, err := weave.NewNewWeave(&sname, data.Name, data.Tags)
	if err != nil {
		t.Fatal(err)
	}

	err = data.Save(wr)
	if err != nil {
		t.Fatal(err)
	}

	err = wr.Close()
	if err != nil {
		t.Fatal(err)
	}

	hd, err := weave.ReadHeader(&sname)
	if err != nil {
		t.Fatal(err)
	}

	if len(hd.Deltas) != 1 || hd.Deltas[0].Name != data.Name || !reflect.DeepEqual(hd.Deltas[0].Tags, data.Tags) {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(hd)
		t.Fatalf("Mismatch reading back delta header")
	}

	nums := make([]int, 0)
	sink := func(line string) error {
		num, err := strconv.Atoi(line)
		if err != nil {
			return err
		}
		nums = append(nums, num)
		return nil
	}
	err = weave.ReadDelta(&sname, 1, sink)
	if err == nil {
		t.Fatalf("Unexpected empty return from ReadDelta")
	}
	if err != io.EOF {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(nums, data.Data) {
		t.Fatalf("Mismatch reading nums:\ngot: %v\nExpect: %v\n", nums, data.Data)
	}
}

type DataSet struct {
	Data []int
	Name string
	Tags map[string]string
}

func NewDataSet(limit int) *DataSet {
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
	}
}

func (d *DataSet) Save(w io.Writer) error {
	for _, num := range d.Data {
		_, err := fmt.Fprintf(w, "%d\n", num)
		if err != nil {
			return err
		}
	}
	return nil
}
