package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/nabidam/befrest/internal/proto"
	"github.com/nabidam/befrest/internal/transfer"
)

func TestTransferHTTPRelayLifecycle(t *testing.T) {
	handler, err := New(fstest.MapFS{"dist/index.html": &fstest.MapFile{Data: []byte("Befrest")}})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	sender := dialWS(t, server.URL, "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	defer sender.CloseNow()
	writeFrame(t, sender, proto.Hello{Type: proto.MsgHello, Name: "Sender"})
	senderWelcome := readTransferFrame(t, sender)
	_ = readTransferFrame(t, sender)
	receiver := dialWS(t, server.URL, "Mozilla/5.0 (iPhone)")
	defer receiver.CloseNow()
	writeFrame(t, receiver, proto.Hello{Type: proto.MsgHello, Name: "Receiver"})
	receiverWelcome := readTransferFrame(t, receiver)
	_ = readTransferFrame(t, receiver)
	_ = readTransferFrame(t, sender)

	writeFrame(t, sender, proto.OfferRequest{Type: proto.MsgOffer, To: receiverWelcome.Self.ID, Files: []proto.FileMeta{{Name: "payload.bin", Size: 100 * 1024 * 1024}}})
	created := readTransferFrame(t, sender)
	if created.Type != proto.MsgOfferCreated {
		t.Fatalf("offer-created frame = %#v", created)
	}
	incoming := readTransferFrame(t, receiver)
	if incoming.Type != proto.MsgOffer || incoming.Transfer.SenderID != senderWelcome.Self.ID {
		t.Fatalf("incoming offer = %#v", incoming)
	}
	writeFrame(t, receiver, proto.TransferID{Type: proto.MsgAccept, TransferID: created.Transfer.ID})
	if accepted := readTransferFrame(t, sender); accepted.Type != proto.MsgTransferAccepted {
		t.Fatalf("accepted frame = %#v", accepted)
	}

	payload := serverBytes(100 * 1024 * 1024)
	uploadDone := make(chan *http.Response, 1)
	go func() {
		request, requestErr := http.NewRequest(http.MethodPost, server.URL+"/api/transfers/"+created.Transfer.ID+"/files/0", bytes.NewReader(payload))
		if requestErr != nil {
			t.Errorf("new upload request: %v", requestErr)
			return
		}
		response, requestErr := http.DefaultClient.Do(request)
		if requestErr != nil {
			t.Errorf("upload: %v", requestErr)
			return
		}
		uploadDone <- response
	}()
	ready := readTransferFrame(t, receiver)
	if ready.Type != proto.MsgFileReady {
		t.Fatalf("file-ready frame = %#v", ready)
	}
	download, err := http.Get(server.URL + "/api/transfers/" + created.Transfer.ID + "/files/0")
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	received, err := io.ReadAll(download.Body)
	download.Body.Close()
	if err != nil {
		t.Fatalf("read download: %v", err)
	}
	if download.StatusCode != http.StatusOK || download.Header.Get("Content-Disposition") != `attachment; filename="payload.bin"` {
		t.Fatalf("download response = %d, %#v", download.StatusCode, download.Header)
	}
	if sha256.Sum256(payload) != sha256.Sum256(received) {
		t.Fatal("downloaded content differs from uploaded content")
	}
	response := <-uploadDone
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d", response.StatusCode)
	}

	progress := readThroughDone(t, sender)
	if progress.Sent != progress.Size || progress.Size != int64(len(payload)) {
		t.Fatalf("final progress = %#v", progress)
	}
	if progress := readThroughDone(t, receiver); progress.Sent != int64(len(payload)) {
		t.Fatalf("receiver final progress = %#v", progress)
	}
}

func TestTransferHTTPValidation(t *testing.T) {
	handler, err := New(fstest.MapFS{"dist/index.html": &fstest.MapFile{Data: []byte("Befrest")}})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, method, path string
		body               io.Reader
		want               int
	}{
		{"unknown", http.MethodPost, "/api/transfers/missing/files/0", bytes.NewReader(nil), http.StatusNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, test.body)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status = %d, want %d", response.Code, test.want)
			}
		})
	}
	hub := newWebSocketHub()
	transfer, err := hub.transfers.Offer("sender", "receiver", []transfer.FileMeta{{Name: "file", Size: 1}})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/transfers/"+transfer.ID+"/files/0", http.NoBody)
	request.ContentLength = -1
	response := httptest.NewRecorder()
	hub.serveFiles(response, request)
	if response.Code != http.StatusLengthRequired {
		t.Fatalf("missing Content-Length status = %d, want 411", response.Code)
	}
	request = httptest.NewRequest(http.MethodPost, "/api/transfers/"+transfer.ID+"/files/0", bytes.NewReader([]byte("x")))
	response = httptest.NewRecorder()
	hub.serveFiles(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("POST before accept status = %d, want 409", response.Code)
	}
	if _, err := hub.transfers.Accept(transfer.ID, "receiver"); err != nil {
		t.Fatal(err)
	}
	request = httptest.NewRequest(http.MethodPost, "/api/transfers/"+transfer.ID+"/files/0", bytes.NewReader([]byte("xx")))
	response = httptest.NewRecorder()
	hub.serveFiles(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("wrong Content-Length status = %d, want 400", response.Code)
	}
}

func readThroughDone(t *testing.T, conn *websocket.Conn) transferFrame {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	var latestProgress transferFrame
	for time.Now().Before(deadline) {
		frame := readTransferFrame(t, conn)
		if frame.Type == proto.MsgProgress {
			latestProgress = frame
		}
		if frame.Type == proto.MsgTransferDone {
			return latestProgress
		}
	}
	t.Fatal("did not receive transfer-done")
	return transferFrame{}
}

type transferFrame struct {
	Type       string         `json:"type"`
	Self       proto.Device   `json:"self"`
	TransferID string         `json:"transferId"`
	Transfer   proto.Transfer `json:"transfer"`
	Index      int            `json:"index"`
	Sent       int64          `json:"sent"`
	Size       int64          `json:"size"`
}

func readTransferFrame(t *testing.T, conn *websocket.Conn) transferFrame {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var frame transferFrame
	if err := wsjson.Read(ctx, conn, &frame); err != nil {
		t.Fatalf("read frame: %v", err)
	}
	return frame
}

func serverBytes(size int) []byte {
	value := make([]byte, size)
	for i := range value {
		value[i] = byte(i*17 + i/127)
	}
	return value
}

var _ = context.Background
