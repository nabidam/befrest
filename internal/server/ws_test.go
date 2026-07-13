package server

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/nabidam/befrest/internal/proto"
)

func TestWebSocketPresenceHandshakeAndFanout(t *testing.T) {
	handler, err := New(fstest.MapFS{"dist/index.html": &fstest.MapFile{Data: []byte("Befrest")}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	android := dialWS(t, server.URL, "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit")
	defer android.CloseNow()
	writeFrame(t, android, proto.Hello{Type: proto.MsgHello})
	needName := readFrame(t, android)
	if needName.Type != proto.MsgNeedName || needName.Suggested != "Pixel 8" {
		t.Fatalf("need-name = %#v, want Pixel 8 suggestion", needName)
	}
	writeFrame(t, android, proto.SetName{Type: proto.MsgSetName, Name: "Pixel 8"})
	welcomeOne := readFrame(t, android)
	if welcomeOne.Type != proto.MsgWelcome || welcomeOne.Self.Kind != "mobile" {
		t.Fatalf("welcome = %#v, want mobile welcome", welcomeOne)
	}
	assertDevices(t, readFrame(t, android), "Pixel 8")

	desktop := dialWS(t, server.URL, "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	defer desktop.CloseNow()
	writeFrame(t, desktop, proto.Hello{Type: proto.MsgHello, Name: "Pixel 8"})
	welcomeTwo := readFrame(t, desktop)
	if welcomeTwo.Type != proto.MsgWelcome || welcomeTwo.Self.Name != "Pixel 8 (2)" || welcomeTwo.Self.Kind != "desktop" {
		t.Fatalf("returning welcome = %#v", welcomeTwo)
	}
	assertDevices(t, readFrame(t, desktop), "Pixel 8", "Pixel 8 (2)")
	assertDevices(t, readFrame(t, android), "Pixel 8", "Pixel 8 (2)")

	if err := desktop.Close(websocket.StatusNormalClosure, "done"); err != nil {
		t.Fatalf("close desktop: %v", err)
	}
	assertDevices(t, readFrame(t, android), "Pixel 8")
}

func TestWebSocketRejectsBadUpgrade(t *testing.T) {
	handler, err := New(fstest.MapFS{"dist/index.html": &fstest.MapFile{Data: []byte("Befrest")}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/ws", nil))
	if response.Code != 400 {
		t.Fatalf("GET /ws without upgrade = %d, want 400", response.Code)
	}
}

func TestWebSocketHostTokenJoinsHostAndSendsInviteInfo(t *testing.T) {
	handler, err := New(fstest.MapFS{"dist/index.html": &fstest.MapFile{Data: []byte("Befrest")}}, Config{
		HostToken: "one-time-token",
		HostName:  "Befrest Host",
		Invite: proto.InviteInfo{
			Type: proto.MsgInviteInfo,
			URLs: proto.InviteURLs{
				MDNS: "http://befrest.local:5311",
				IP:   "http://192.168.1.10:5311",
			},
			Port: 5311,
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	host := dialWS(t, server.URL, "Mozilla/5.0 (X11; Linux x86_64)")
	defer host.CloseNow()
	writeFrame(t, host, proto.Hello{Type: proto.MsgHello, HostToken: "one-time-token"})
	welcome := readFrame(t, host)
	if welcome.Type != proto.MsgWelcome || !welcome.Self.IsHost || welcome.Self.Name != "Befrest Host" {
		t.Fatalf("host welcome = %#v", welcome)
	}
	invite := readFrame(t, host)
	if invite.Type != proto.MsgInviteInfo || invite.Port != 5311 || invite.URLs.IP != "http://192.168.1.10:5311" || invite.URLs.MDNS != "http://befrest.local:5311" {
		t.Fatalf("invite-info = %#v", invite)
	}
	assertDevices(t, readFrame(t, host), "Befrest Host")

	second := dialWS(t, server.URL, "Mozilla/5.0")
	defer second.CloseNow()
	writeFrame(t, second, proto.Hello{Type: proto.MsgHello, HostToken: "one-time-token"})
	if frame := readFrame(t, second); frame.Type != proto.MsgNeedName {
		t.Fatalf("reused host token frame = %#v, want need-name", frame)
	}
}

func TestDeviceSuggestionClassifiesSupportedUserAgents(t *testing.T) {
	tests := []struct {
		name, userAgent, suggestion, kind string
	}{
		{"android model", "Mozilla/5.0 (Linux; Android 14; Pixel 8)", "Pixel 8", "mobile"},
		{"iphone", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)", "iPhone", "mobile"},
		{"ipad", "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X)", "iPad", "mobile"},
		{"windows", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "Windows PC", "desktop"},
		{"mac", "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0)", "Mac", "desktop"},
		{"linux", "Mozilla/5.0 (X11; Linux x86_64)", "Linux PC", "desktop"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			suggestion, kind := deviceSuggestion(test.userAgent)
			if suggestion != test.suggestion || kind != test.kind {
				t.Fatalf("deviceSuggestion() = (%q, %q), want (%q, %q)", suggestion, kind, test.suggestion, test.kind)
			}
		})
	}
}

func dialWS(t *testing.T, serverURL, userAgent string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws"
	conn, response, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: map[string][]string{"User-Agent": {userAgent}}})
	if err != nil {
		t.Fatalf("Dial(%s) error = %v (response: %#v)", wsURL, err, response)
	}
	return conn
}

func writeFrame(t *testing.T, conn *websocket.Conn, frame any) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := wsjson.Write(ctx, conn, frame); err != nil {
		t.Fatalf("write frame: %v", err)
	}
}

type receivedFrame struct {
	Type      string           `json:"type"`
	Suggested string           `json:"suggested"`
	Self      proto.Device     `json:"self"`
	Devices   []proto.Device   `json:"devices"`
	URLs      proto.InviteURLs `json:"urls"`
	Port      int              `json:"port"`
}

func readFrame(t *testing.T, conn *websocket.Conn) receivedFrame {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var frame receivedFrame
	if err := wsjson.Read(ctx, conn, &frame); err != nil {
		t.Fatalf("read frame: %v", err)
	}
	return frame
}

func assertDevices(t *testing.T, frame receivedFrame, names ...string) {
	t.Helper()
	if frame.Type != proto.MsgDevices || len(frame.Devices) != len(names) {
		t.Fatalf("devices = %#v, want %v", frame, names)
	}
	for i, name := range names {
		if frame.Devices[i].Name != name {
			t.Fatalf("devices[%d] = %q, want %q", i, frame.Devices[i].Name, name)
		}
	}
}
