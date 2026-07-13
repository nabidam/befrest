package transfer

import (
	"bytes"
	"crypto/sha256"
	"io"
	"testing"
	"time"
)

func TestManagerRelaysAndCompletesTransfer(t *testing.T) {
	events := make(chan Event, 512)
	manager := NewManager(func(event Event) { events <- event })
	transfer, err := manager.Offer("sender", "receiver", []FileMeta{{Name: "payload.bin", Size: 100 * 1024 * 1024}})
	if err != nil {
		t.Fatalf("Offer() error = %v", err)
	}
	if _, err := manager.Accept(transfer.ID, "receiver"); err != nil {
		t.Fatalf("Accept() error = %v", err)
	}

	payload := managerBytes(100 * 1024 * 1024)
	uploaded := make(chan error, 1)
	go func() { uploaded <- manager.Upload(transfer.ID, 0, bytes.NewReader(payload)) }()
	waitForEvent(t, events, EventFileReady)
	var received bytes.Buffer
	file, err := manager.Download(transfer.ID, 0, &received)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	if err := <-uploaded; err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if file.Name != "payload.bin" || file.Size != int64(len(payload)) {
		t.Fatalf("downloaded metadata = %#v", file)
	}
	if sha256.Sum256(payload) != sha256.Sum256(received.Bytes()) {
		t.Fatal("relayed bytes differ from upload")
	}

	var final Progress
	done := map[string]bool{}
	for len(events) > 0 {
		event := <-events
		if event.Type == EventProgress && event.Sent == int64(len(payload)) {
			final = Progress{sent: event.Sent, size: event.Size}
		}
		if event.Type == EventDone {
			done[event.To] = true
		}
	}
	if final.sent != int64(len(payload)) || final.size != int64(len(payload)) {
		t.Fatalf("final progress = %#v", final)
	}
	if !done["sender"] || !done["receiver"] {
		t.Fatalf("done recipients = %#v", done)
	}
}

func TestManagerZeroByteAndTerminalCleanup(t *testing.T) {
	now := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	events := make(chan Event, 8)
	manager := NewManagerWithClock(func(event Event) { events <- event }, func() time.Time { return now })
	transfer, err := manager.Offer("sender", "receiver", []FileMeta{{Name: "empty.txt", Size: 0}})
	if err != nil {
		t.Fatalf("Offer() error = %v", err)
	}
	if _, err := manager.Accept(transfer.ID, "receiver"); err != nil {
		t.Fatalf("Accept() error = %v", err)
	}
	uploaded := make(chan error, 1)
	go func() { uploaded <- manager.Upload(transfer.ID, 0, bytes.NewReader(nil)) }()
	waitForEvent(t, events, EventFileReady)
	var received bytes.Buffer
	if _, err := manager.Download(transfer.ID, 0, &received); err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	if err := <-uploaded; err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	now = now.Add(time.Minute)
	manager.Cleanup()
	if _, _, err := manager.File(transfer.ID, 0); err != ErrNotFound {
		t.Fatalf("File() after retention = %v, want ErrNotFound", err)
	}
}

func TestManagerRejectsInvalidStates(t *testing.T) {
	manager := NewManager(nil)
	transfer, err := manager.Offer("sender", "receiver", []FileMeta{{Name: "file", Size: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Upload(transfer.ID, 0, bytes.NewReader([]byte("x"))); err != ErrWrongState {
		t.Fatalf("Upload before accept = %v, want ErrWrongState", err)
	}
	if _, _, err := manager.File("missing", 0); err != ErrNotFound {
		t.Fatalf("unknown transfer = %v, want ErrNotFound", err)
	}
	if _, err := manager.Accept(transfer.ID, "other"); err != ErrNotReceiver {
		t.Fatalf("Accept wrong receiver = %v, want ErrNotReceiver", err)
	}
}

type Progress struct{ sent, size int64 }

func waitForEvent(t *testing.T, events <-chan Event, eventType string) {
	t.Helper()
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	for {
		select {
		case event := <-events:
			if event.Type == eventType {
				return
			}
		case <-deadline.C:
			t.Fatalf("timed out waiting for %s", eventType)
		}
	}
}

func managerBytes(size int) []byte {
	value := make([]byte, size)
	for i := range value {
		value[i] = byte(i*31 + i/251)
	}
	return value
}

var _ io.Reader = (*bytes.Reader)(nil)
