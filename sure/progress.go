package sure

import (
	"fmt"
	"io"
	"sync"
)

// A progress meter.  Records files visited, and number of bytes
// processed.  When given an estimate, prints a simple periodic report
// of how far along we think we are.
type Progress struct {
	wr         io.Writer
	curFiles   uint64
	totalFiles uint64

	curBytes   uint64
	totalBytes uint64

	lock sync.Mutex
}

func NewProgress(files, bytes uint64, wr io.Writer) Progress {
	return Progress{
		totalFiles: files,
		totalBytes: bytes,
		wr:         wr,
	}
}

// Update the progress meter, displaying it if enough time has
// elapsed.
func (p *Progress) Update(files, bytes uint64) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.curFiles += files
	p.curBytes += bytes
	p.flush()
}

// Flush the output, regardless of any update needed.
func (p *Progress) Flush() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.flush()
}

// flush does the actual flush, assuming the lock is already taken.
func (p *Progress) flush() {
	fmt.Fprintf(p.wr, "%7d/%7d (%5.1f%%) files, %s/%s (%5.1f%%) bytes\n",
		p.curFiles, p.totalFiles,
		float64(p.curFiles)*100.0/float64(p.totalFiles),
		humanize(p.curBytes), humanize(p.totalBytes),
		float64(p.curBytes)*100.0/float64(p.totalBytes))
}

// Print a size in a more human-friendly format.
func humanize(value uint64) string {
	v := float64(value)
	unit := 0

	for v > 1024.0 {
		v /= 1024.0
		unit += 1
	}

	var str string
	if v < 10.0 {
		str = "%6.3f%s"
	} else if v < 100.0 {
		str = "%6.2f%s"
	} else {
		str = "%6.1f%s"
	}

	return fmt.Sprintf(str, v, units[unit])
}

var units []string = []string{
	"B  ", "Kib", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB",
}
