// Package transfer owns the in-memory byte flow for active transfers.
package transfer

import (
	"errors"
	"io"
	"sync"
	"time"
)

// PipeCapacity is the maximum number of relay bytes kept in memory for one
// active file. Writers block when the buffer is full, applying backpressure to
// the HTTP upload that feeds them.
const PipeCapacity = 4 * 1024 * 1024

var ErrDeadlineExceeded = errors.New("relay deadline exceeded")

// Pipe is a bounded, in-memory rendezvous between one upload writer and one
// download reader. It implements neither network nor HTTP concerns.
type Pipe struct {
	mu sync.Mutex

	notEmpty *sync.Cond
	notFull  *sync.Cond

	buffer  []byte
	readAt  int
	writeAt int
	used    int

	readerClosed bool
	writerClosed bool
	terminalErr  error
	timeout      time.Duration
}

// Reader is the download side of a Pipe.
type Reader struct {
	pipe *Pipe
}

// Writer is the upload side of a Pipe.
type Writer struct {
	pipe *Pipe
}

// NewPipe creates a relay with a fixed 4 MiB buffer. Closing a Reader stops a
// blocked Writer with io.ErrClosedPipe. Closing a Writer lets the Reader drain
// any buffered bytes, then returns io.EOF.
func NewPipe() (*Reader, *Writer) {
	return NewPipeWithTimeout(30 * time.Second)
}

func NewPipeWithTimeout(timeout time.Duration) (*Reader, *Writer) {
	pipe := &Pipe{buffer: make([]byte, PipeCapacity), timeout: timeout}
	pipe.notEmpty = sync.NewCond(&pipe.mu)
	pipe.notFull = sync.NewCond(&pipe.mu)
	return &Reader{pipe: pipe}, &Writer{pipe: pipe}
}

// Read drains bytes from the relay buffer, blocking until bytes arrive or the
// upload side closes.
func (r *Reader) Read(destination []byte) (int, error) {
	if len(destination) == 0 {
		return 0, nil
	}

	p := r.pipe
	p.mu.Lock()
	defer p.mu.Unlock()
	stopDeadline := p.startDeadlineLocked()
	defer stopDeadline()

	for p.used == 0 && !p.writerClosed && !p.readerClosed && p.terminalErr == nil {
		p.notEmpty.Wait()
	}
	if p.terminalErr != nil {
		return 0, p.terminalErr
	}
	if p.readerClosed {
		return 0, io.ErrClosedPipe
	}
	if p.used == 0 {
		return 0, io.EOF
	}

	count := min(len(destination), p.used)
	first := min(count, len(p.buffer)-p.readAt)
	copy(destination, p.buffer[p.readAt:p.readAt+first])
	copy(destination[first:count], p.buffer[:count-first])
	p.readAt = (p.readAt + count) % len(p.buffer)
	p.used -= count
	p.notFull.Broadcast()
	return count, nil
}

// Close closes the download side and unblocks the upload side.
func (r *Reader) Close() error {
	r.abort(io.ErrClosedPipe)
	return nil
}

// Write places bytes in the relay buffer, blocking when it reaches capacity.
func (w *Writer) Write(source []byte) (int, error) {
	p := w.pipe
	p.mu.Lock()
	defer p.mu.Unlock()
	stopDeadline := p.startDeadlineLocked()
	defer stopDeadline()

	if p.terminalErr != nil {
		return 0, p.terminalErr
	}
	if p.writerClosed || p.readerClosed {
		return 0, io.ErrClosedPipe
	}

	written := 0
	for written < len(source) {
		for p.used == len(p.buffer) && !p.readerClosed && p.terminalErr == nil {
			p.notFull.Wait()
		}
		if p.terminalErr != nil {
			return written, p.terminalErr
		}
		if p.readerClosed {
			return written, io.ErrClosedPipe
		}

		count := min(len(source)-written, len(p.buffer)-p.used)
		first := min(count, len(p.buffer)-p.writeAt)
		copy(p.buffer[p.writeAt:p.writeAt+first], source[written:written+first])
		copy(p.buffer[:count-first], source[written+first:written+count])
		p.writeAt = (p.writeAt + count) % len(p.buffer)
		p.used += count
		written += count
		p.notEmpty.Broadcast()
	}
	return written, nil
}

// Close closes the upload side and unblocks the download side.
func (w *Writer) Close() error {
	p := w.pipe
	p.mu.Lock()
	if !p.writerClosed {
		p.writerClosed = true
		p.notEmpty.Broadcast()
		p.notFull.Broadcast()
	}
	p.mu.Unlock()
	return nil
}

// Abort tears down both endpoints immediately.
func (r *Reader) Abort(err error) { r.abort(err) }
func (w *Writer) Abort(err error) { w.pipe.abort(err) }

func (r *Reader) abort(err error) { r.pipe.abort(err) }

func (p *Pipe) abort(err error) {
	if err == nil {
		err = io.ErrClosedPipe
	}
	p.mu.Lock()
	if p.terminalErr == nil {
		p.terminalErr = err
	}
	p.readerClosed = true
	p.writerClosed = true
	p.notFull.Broadcast()
	p.notEmpty.Broadcast()
	p.mu.Unlock()
}

func (p *Pipe) startDeadlineLocked() func() {
	if p.timeout <= 0 {
		return func() {}
	}
	timer := time.AfterFunc(p.timeout, func() { p.abort(ErrDeadlineExceeded) })
	return func() { timer.Stop() }
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}
