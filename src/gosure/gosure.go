// File integrity testing.

package main

import (
	"fmt"
	"linuxdir"
	"log"
	"os"
	"sha"
	"strconv"
)

var _ = linuxdir.Readdir
var _ = fmt.Printf
var _ = log.Printf

const magic = "asure-2.0\n-----\n"

func main() {
	dir, err := WalkRoot(".")
	if err != nil {
		log.Fatalf("Unable to walk root directory: %s", err)
	}

	writeSure("2sure.0.gz", dir)
}

type Node struct {
	name   string
	atts   map[string]string
	costly func() map[string]string // Get the atts that are costly to make.
}

const hexDigits = "0123456789abcdef"

func makeNode(path string, info *os.FileInfo) (n *Node) {
	atts := make(map[string]string)
	var costly func() map[string]string

	switch {
	case info.IsDirectory():
		atts["kind"] = "dir"
		atts["uid"] = strconv.Itoa(info.Uid)
		atts["gid"] = strconv.Itoa(info.Gid)
		atts["perm"] = strconv.Uitoa64(uint64(info.Permission()))
	case info.IsRegular():
		atts["kind"] = "file"
		atts["uid"] = strconv.Itoa(info.Uid)
		atts["gid"] = strconv.Itoa(info.Gid)
		atts["perm"] = strconv.Uitoa64(uint64(info.Permission()))
		atts["mtime"] = strconv.Itoa64(info.Mtime_ns / 1000000000)
		atts["ctime"] = strconv.Itoa64(info.Ctime_ns / 1000000000)
		atts["ino"] = strconv.Uitoa64(info.Ino)

		costly = func() (atts map[string]string) {
			atts = make(map[string]string)
			hash, err := sha.HashFile(path + "/" + info.Name)
			if err != nil {
				log.Printf("Unable to hash file: %s", path+"/"+info.Name)
			}
			hex := make([]byte, 40)
			for i, ch := range hash {
				hex[2*i] = hexDigits[ch>>4]
				hex[2*i+1] = hexDigits[ch&0xf]
			}
			atts["sha1"] = string(hex)
			return
		}
	case info.IsSymlink():
		atts["kind"] = "lnk"
		target, err := os.Readlink(path + "/" + info.Name)
		if err != nil {
			log.Printf("Error reading symlink: %s", path+"/"+info.Name)
		} else {
			atts["targ"] = target
		}
	default:
		fmt.Printf("Node: %v\n", *info)
		panic("Unexpected file type")
	}

	n = &Node{name: info.Name, atts: atts, costly: costly}
	return

}
