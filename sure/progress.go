package sure

import (
	"fmt"
	"log"
	"time"
)

// A progress meter.  Records files visited, and number of bytes
// processed.  When given an estimate, prints a simple periodic report
// of how far along we think we are.
type Progress struct {
	next time.Time

	curFiles   uint64
	totalFiles uint64

	curBytes   uint64
	totalBytes uint64
}

func NewProgress(files, bytes uint64) Progress {
	return Progress{
		totalFiles: files,
		totalBytes: bytes,
		next:       time.Now().Add(5 * time.Second),
	}
}

// Update the progress meter, displaying it if enough time has
// elapsed.
func (p *Progress) Update(files, bytes uint64) {
	p.curFiles += files
	p.curBytes += bytes

	if time.Now().After(p.next) {
		p.Flush()
	}
}

// Flush the output, regardless of any update needed.
func (p *Progress) Flush() {
	log.Printf("%7d/%7d (%5.1f%%) files, %s/%s (%5.1f%%) bytes",
		p.curFiles, p.totalFiles,
		float64(p.curFiles)*100.0/float64(p.totalFiles),
		humanize(p.curBytes), humanize(p.totalBytes),
		float64(p.curBytes)*100.0/float64(p.totalBytes))

	p.next = time.Now().Add(5 * time.Second)
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
