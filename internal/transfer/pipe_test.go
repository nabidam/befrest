package transfer

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"runtime"
	"testing"
	"time"
)

func TestPipeRelaysLargeStreamWithBoundedHeap(t *testing.T) {
	const size = 100 * 1024 * 1024
	source := deterministicBytes(size)
	wantHash := sha256.Sum256(source)

	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	reader, writer := NewPipe()
	var gotHash [sha256.Size]byte
	var readErr error
	readerDone := make(chan struct{})
	go func() {
		hash := sha256.New()
		_, readErr = io.Copy(hash, reader)
		copy(gotHash[:], hash.Sum(nil))
		close(readerDone)
	}()

	if _, err := io.Copy(writer, bytes.NewReader(source)); err != nil {
		t.Fatalf("copy into pipe: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	<-readerDone
	if readErr != nil {
		t.Fatalf("copy from pipe: %v", readErr)
	}

	if gotHash != wantHash {
		t.Fatal("relayed stream hash differs from source")
	}
	runtime.ReadMemStats(&after)
	if growth := after.HeapAlloc - before.HeapAlloc; growth >= 16*1024*1024 {
		t.Fatalf("heap grew by %d bytes, want less than 16 MiB", growth)
	}
}

func TestPipeAppliesBackpressureUntilReaderDrains(t *testing.T) {
	reader, writer := NewPipe()
	writeDone := make(chan error, 1)
	go func() {
		_, err := writer.Write(bytes.Repeat([]byte{1}, PipeCapacity+1))
		writeDone <- err
	}()

	select {
	case err := <-writeDone:
		t.Fatalf("write completed before reader drained: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	if _, err := io.CopyN(io.Discard, reader, 1); err != nil {
		t.Fatalf("drain one byte: %v", err)
	}
	select {
	case err := <-writeDone:
		if err != nil {
			t.Fatalf("write after drain: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("write remained blocked after reader drained")
	}
}

func TestPipeHandlesZeroBytesAndCloseTeardown(t *testing.T) {
	t.Run("zero byte stream", func(t *testing.T) {
		reader, writer := NewPipe()
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}
		if _, err := reader.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
			t.Fatalf("read after empty stream = %v, want EOF", err)
		}
	})

	t.Run("reader close unblocks writer", func(t *testing.T) {
		reader, writer := NewPipe()
		if _, err := writer.Write(bytes.Repeat([]byte{1}, PipeCapacity)); err != nil {
			t.Fatalf("fill pipe: %v", err)
		}
		result := make(chan error, 1)
		go func() {
			_, err := writer.Write([]byte{1})
			result <- err
		}()
		if err := reader.Close(); err != nil {
			t.Fatalf("close reader: %v", err)
		}
		if err := <-result; !errors.Is(err, io.ErrClosedPipe) {
			t.Fatalf("blocked writer error = %v, want io.ErrClosedPipe", err)
		}
	})

	t.Run("writer close unblocks reader", func(t *testing.T) {
		reader, writer := NewPipe()
		result := make(chan error, 1)
		go func() {
			_, err := reader.Read(make([]byte, 1))
			result <- err
		}()
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}
		if err := <-result; !errors.Is(err, io.EOF) {
			t.Fatalf("blocked reader error = %v, want EOF", err)
		}
	})
}

func deterministicBytes(size int) []byte {
	data := make([]byte, size)
	var state uint64 = 1
	for index := range data {
		state = state*6364136223846793005 + 1
		data[index] = byte(state >> 56)
	}
	return data
}

var _ io.ReadCloser = (*Reader)(nil)
var _ io.WriteCloser = (*Writer)(nil)
