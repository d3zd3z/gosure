// Package status provides a writer for various types of status and is
// responsible for making sure all of this output is presented cleanly
// on Stdout.  It also captures output from the standard logger and
// makes sure its output doesn't interfere with other output.
package status

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Manager is the primary status manager.  It implements io.Writer and
// can function similar to os.Stdout for this purpose.  It also
// captures the standard logger and outputs that with the regular
// output.
type Manager struct {
	meter  []byte     // Text of current meter, if one.
	dangle bool       // Is the last line dangling (no newline)?
	lock   sync.Mutex // To sync 'lines' access.
}

// Meter is a special writer wrapped around a manager.  Writes to the
// meter should be one or more lines of text.  Subsequent writes that
// did not have another type of write to the underlying Manager will
// be preceeded by cursor movement and clearning to overwrite the
// meter with an updated value.
type Meter struct {
	m     *Manager      // The Manager to write to.
	delay time.Duration // How frequently to update the meter.
	next  time.Time     // When we can next write.
}

// NewManager creates a new manager, capturing the current logger.
func NewManager() *Manager {
	var m Manager
	log.SetOutput(&m)
	return &m
}

// Write outputs data through to stdout, without corrupting the meter.
func (m *Manager) Write(p []byte) (n int, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(p) > 0 {
		err = m.clear()
		if err != nil {
			return 0, err
		}
		m.dangle = p[len(p)-1] != '\n'
		n, err = os.Stdout.Write(p)
		if err != nil {
			return 0, err
		}
		err = m.redraw()
	}

	return n, err
}

// Printf convenience.
func (m *Manager) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(m, format, a...)
}

// Close closes the meter.  The meter is left in place, and the
// standard logger is restored.
func (m *Manager) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Update the meter to the latest values.
	m.clear()
	m.redraw()

	log.SetOutput(os.Stdout)
	m.meter = nil

	return nil
}

// Meter returns a wrapper meter associated with the given text.  The
// duration is the minimal update frequency of the meter.
func (m *Manager) Meter(delay time.Duration) *Meter {
	return &Meter{
		m:     m,
		delay: delay,
		next:  time.Now(),
	}
}

// clear clears any written lines.
func (m *Manager) clear() error {
	for _, ch := range m.meter {
		if ch == '\n' {
			_, err := os.Stdout.WriteString("\x1b[1A\x1b[2K")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// redraw draws the current meter.
func (m *Manager) redraw() error {
	_, err := os.Stdout.Write(m.meter)
	return err
}

// Write sets a progress meter.  The write should consist of one or
// more lines of text followed by '\n' (the entire write should end
// with '\n'.  The lines should be short enough to not wrap the user's
// terminal.  If other writes are interleaved, this text will be
// cleared, and rewritten after that other text is printed.
func (me *Meter) Write(p []byte) (n int, err error) {
	me.m.lock.Lock()
	defer me.m.lock.Unlock()

	// Make a copy of the meter, so we don't worry about it
	// changing by the user.
	var newMeter []byte
	if p != nil {
		newMeter = make([]byte, len(p))
		copy(newMeter, p)
	}

	now := time.Now()
	if now.Before(me.next) {
		me.m.meter = newMeter
		// Not enough time, don't show the meter.
		return len(p), nil
	}

	me.next = now.Add(me.delay)

	err = me.m.clear()
	if err != nil {
		return
	}
	me.m.meter = newMeter

	// The return isn't quite right here, but I'm not sure we're
	// going to handle errors writing to Stdout anyway.
	return len(p), me.m.redraw()
}

// Printf for convenience.
func (me *Meter) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(me, format, a...)
}

// Flush the meter.
func (me *Meter) Flush() error {
	me.m.lock.Lock()
	defer me.m.lock.Unlock()

	err := me.m.clear()
	if err != nil {
		return err
	}

	return me.m.redraw()
}

// Close finishes the progress meter.  Its output is flushed, and new
// printing will occur after the meter.
func (me *Meter) Close() error {
	me.m.lock.Lock()
	defer me.m.lock.Unlock()

	err := me.m.clear()
	err2 := me.m.redraw()

	if err == nil {
		err = err2
	}

	me.m.meter = nil
	return err
}
