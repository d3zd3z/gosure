// Walking the surefile itself.

package main

import (
	"fmt"
	"io"
	"os"
	"compress/gzip"
	"bufio"
)

const magic = "asure-2.0\n-----\n"

const (
	ENTER = iota
	NODE
	MARK
	LEAVE
	)

type Node interface {
	GetKind() int
	GetName() string
	GetAtts() map[string]string
	GetExpensiveAtts() map[string]string
}

func GetAllAtts(node Node) (result map[string]string) {
	result = make(map[string]string)
	for k, v := range node.GetAtts() {
		result[k] = v
	}
	for k, v := range node.GetExpensiveAtts() {
		result[k] = v
	}
	return
}

type SimpleNode struct {
	kind int
	name string
	atts map[string]string
}

type VerySimpleNode struct {
	kind int
}

var theMark = VerySimpleNode {MARK}
var theLeave = VerySimpleNode {MARK}

func (n *VerySimpleNode) GetKind() int { return n.kind }
func (n *VerySimpleNode) GetName() string { return "" }
func (n *VerySimpleNode) GetAtts() map[string]string { return make(map[string]string) }
func (n *VerySimpleNode) GetExpensiveAtts() map[string]string { return make(map[string]string) }

func (n *SimpleNode) GetKind() int { return n.kind }
func (n *SimpleNode) GetName() string { return n.name }
func (n *SimpleNode) GetAtts() map[string]string { return n.atts }
func (n *SimpleNode) GetExpensiveAtts() map[string]string { return make(map[string]string) }

type sureReader struct {
	cleanup func()
	file *bufio.Reader
	depth int
}

func (r *sureReader) Close() {
	r.cleanup()
}

func (r *sureReader) readFull(kind int) (Node, os.Error) {
	name, err := r.file.ReadString(' ')
	if err != nil {
		goto error
	}
	err = r.expect('[')
	if err != nil {
		goto error
	}
	atts := make(map[string]string)
	for {
		start, _, err := r.file.ReadRune()
		if err != nil {
			goto error
		}
		if start == ']' {
			break
		}
		key, err := r.file.ReadString(' ')
		if err != nil {
			goto error
		}
		key = string(start) + key
		val, err := r.file.ReadString(' ')
		if err != nil {
			goto error
		}
		atts[key] = val
	}

	err = r.expect('\n')
	if err != nil {
		goto error
	}

	return &SimpleNode{kind, name, atts}, nil

error:
	return nil, err
}

func (r *sureReader) expect(ch int) os.Error {
	code, _, err := r.file.ReadRune()
	if err != nil {
		return err
	}
	if code != ch {
		return os.NewError("Unexpected character")
	}
	return nil
}

func (r *sureReader) ReadNode() (Node, os.Error) {
	code, _, err := r.file.ReadRune()
	if err != nil {
		return nil, err
	}
	switch code {
	case 'd':
		r.depth++
		return r.readFull(ENTER)
	case 'f':
		return r.readFull(NODE)
	case '-':
		err := r.expect('\n')
		return &theMark, err
	case 'u':
		r.depth--
		err := r.expect('\n')
		return &theLeave, err
	}
	panic("Unreached")
}

func surefile(name string) (NodeReader, os.Error) {
	rawFile, err := os.Open(name + ".dat.gz", os.O_RDONLY, 0)
	if err != nil { return nil, err }
	cleanup := func() {
		rawFile.Close()
	}

	cFile, err := gzip.NewReader(rawFile)
	if err != nil {
		cleanup()
		return nil, err
	}

	file := bufio.NewReader(cFile)

	header := make([]byte, 16)
	_, err = io.ReadFull(file, header[:])
	if err != nil {
		cleanup()
		return nil, err
	}

	if string(header) != magic {
		fmt.Printf("%#v", header)
		cleanup()
		return nil, os.NewError("Invalid header on file")
	}

	return &sureReader{cleanup: cleanup, file: file}, nil
}
