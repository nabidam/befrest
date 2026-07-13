package transfer

import (
	"bytes"
	"crypto/sha256"
	"errors"
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

func TestManagerCancelsOffersAndStreamsWithSingleVerdicts(t *testing.T) {
	tests := []struct {
		name, action, wantType, wantReason string
	}{
		{"sender withdraws offer", "offer-cancel", EventOfferCancelled, "sender-cancelled"},
		{"sender cancels stream", "sender-cancel", EventFailed, "cancelled-by-sender"},
		{"receiver cancels stream", "receiver-cancel", EventFailed, "cancelled-by-receiver"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events := make(chan Event, 8)
			manager := NewManager(func(event Event) { events <- event })
			transfer, err := manager.Offer("sender", "receiver", []FileMeta{{Name: "file", Size: 1}})
			if err != nil {
				t.Fatal(err)
			}
			switch test.action {
			case "offer-cancel":
				err = manager.CancelOffer(transfer.ID, "sender")
			case "sender-cancel":
				_, err = manager.Accept(transfer.ID, "receiver")
				if err == nil {
					err = manager.CancelTransfer(transfer.ID, "sender")
				}
			case "receiver-cancel":
				_, err = manager.Accept(transfer.ID, "receiver")
				if err == nil {
					err = manager.CancelTransfer(transfer.ID, "receiver")
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			got := drainEvents(events)
			matches := 0
			for _, event := range got {
				if event.Type == test.wantType && event.Reason == test.wantReason {
					matches++
				}
			}
			want := 1
			if test.wantType == EventFailed {
				want = 2
			}
			if matches != want {
				t.Fatalf("events = %#v, got %d %s verdicts, want %d", got, matches, test.wantReason, want)
			}
		})
	}
}

func TestManagerDisconnectSweepAndRelayDeadline(t *testing.T) {
	t.Run("disconnect reasons cover both roles", func(t *testing.T) {
		for _, test := range []struct{ disconnected, reason string }{{"sender", "sender-disconnected"}, {"receiver", "receiver-disconnected"}} {
			t.Run(test.disconnected, func(t *testing.T) {
				events := make(chan Event, 8)
				manager := NewManager(func(event Event) { events <- event })
				transfer, err := manager.Offer("sender", "receiver", []FileMeta{{Name: "file", Size: 1}})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := manager.Accept(transfer.ID, "receiver"); err != nil {
					t.Fatal(err)
				}
				manager.Disconnect(test.disconnected)
				got := drainEvents(events)
				if countFailure(got, test.reason) != 2 {
					t.Fatalf("events = %#v", got)
				}
			})
		}
	})
	t.Run("stalled relay fails once with stream error", func(t *testing.T) {
		events := make(chan Event, 8)
		manager := NewManagerWithClockAndTimeout(func(event Event) { events <- event }, time.Now, 20*time.Millisecond)
		transfer, err := manager.Offer("sender", "receiver", []FileMeta{{Name: "file", Size: 1}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := manager.Accept(transfer.ID, "receiver"); err != nil {
			t.Fatal(err)
		}
		stalled := newStalledReader()
		uploaded := make(chan error, 1)
		go func() { uploaded <- manager.Upload(transfer.ID, 0, stalled) }()
		waitForEvent(t, events, EventFileReady)
		downloaded := make(chan error, 1)
		go func() { _, err := manager.Download(transfer.ID, 0, io.Discard); downloaded <- err }()
		if err := <-uploaded; !errors.Is(err, ErrDeadlineExceeded) {
			t.Fatalf("Upload() = %v, want deadline", err)
		}
		if err := <-downloaded; !errors.Is(err, ErrDeadlineExceeded) {
			t.Fatalf("Download() = %v, want deadline", err)
		}
		if got := countFailure(drainEvents(events), "stream-error"); got != 2 {
			t.Fatalf("stream-error verdicts = %d, want 2", got)
		}
	})
}

func drainEvents(events <-chan Event) []Event {
	var result []Event
	for len(events) > 0 {
		result = append(result, <-events)
	}
	return result
}

func countFailure(events []Event, reason string) int {
	count := 0
	for _, event := range events {
		if event.Type == EventFailed && event.Reason == reason {
			count++
		}
	}
	return count
}

type stalledReader struct{ closed chan struct{} }

func newStalledReader() *stalledReader            { return &stalledReader{closed: make(chan struct{})} }
func (r *stalledReader) Read([]byte) (int, error) { <-r.closed; return 0, io.ErrClosedPipe }
func (r *stalledReader) Close() error {
	select {
	case <-r.closed:
	default:
		close(r.closed)
	}
	return nil
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
