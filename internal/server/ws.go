package server

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/nabidam/befrest/internal/presence"
	"github.com/nabidam/befrest/internal/proto"
	"github.com/nabidam/befrest/internal/transfer"
)

// webSocketHub translates control-socket frames into presence registry calls.
type webSocketHub struct {
	registry  *presence.Registry
	transfers *transfer.Manager
	mu        sync.RWMutex
	sockets   map[string]*websocket.Conn
	muting    atomic.Int32
}

func newWebSocketHub() *webSocketHub {
	hub := &webSocketHub{sockets: make(map[string]*websocket.Conn)}
	hub.registry = presence.NewRegistry(hub.broadcastDevices)
	hub.transfers = transfer.NewManager(hub.notifyTransfer)
	return hub
}

func (h *webSocketHub) serveWS(writer http.ResponseWriter, request *http.Request) {
	if !strings.EqualFold(request.Header.Get("Upgrade"), "websocket") {
		http.Error(writer, "websocket upgrade required", http.StatusBadRequest)
		return
	}
	conn, err := websocket.Accept(writer, request, nil)
	if err != nil {
		return
	}
	defer conn.CloseNow()

	ctx := context.Background()
	var hello proto.Hello
	if err := wsjson.Read(ctx, conn, &hello); err != nil || hello.Type != proto.MsgHello {
		h.writeError(conn, "bad-request", "first message must be hello")
		return
	}

	name, kind := deviceSuggestion(request.UserAgent())
	var deviceID string
	if strings.TrimSpace(hello.Name) == "" {
		h.write(conn, proto.NeedName{Type: proto.MsgNeedName, Suggested: name})
	} else {
		deviceID = h.joinID(conn, hello.Name, kind)
		if deviceID == "" {
			return
		}
	}

	for {
		var frame json.RawMessage
		if err := wsjson.Read(ctx, conn, &frame); err != nil {
			break
		}
		var envelope struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(frame, &envelope); err != nil {
			h.writeError(conn, "bad-request", "invalid message")
			continue
		}
		switch envelope.Type {
		case proto.MsgSetName:
			var setName proto.SetName
			if err := json.Unmarshal(frame, &setName); err != nil {
				h.writeError(conn, "bad-request", "invalid set-name message")
				continue
			}
			if deviceID == "" {
				deviceID = h.joinID(conn, setName.Name, kind)
				if deviceID == "" {
					return
				}
				continue
			}
			h.muting.Add(1)
			device, err := h.registry.Rename(deviceID, setName.Name)
			h.muting.Add(-1)
			if err != nil {
				h.writeError(conn, "bad-request", "device name must not be empty")
				continue
			}
			h.write(conn, welcome(device))
			h.broadcastDevices(h.registry.Snapshot())
		case proto.MsgHello:
			h.writeError(conn, "bad-request", "hello is only allowed once")
		case proto.MsgOffer:
			if deviceID == "" {
				h.writeError(conn, "bad-request", "set a device name before offering files")
				continue
			}
			var offer proto.OfferRequest
			if err := json.Unmarshal(frame, &offer); err != nil {
				h.writeError(conn, "bad-request", "invalid offer message")
				continue
			}
			if !h.connected(offer.To) {
				h.writeError(conn, "target-gone", "target device is no longer connected")
				continue
			}
			files := make([]transfer.FileMeta, len(offer.Files))
			for i, file := range offer.Files {
				files[i] = transfer.FileMeta{Name: file.Name, Size: file.Size}
			}
			created, err := h.transfers.Offer(deviceID, offer.To, files)
			if err != nil {
				h.writeError(conn, "bad-request", "offer must contain named files with non-negative sizes")
				continue
			}
			h.write(conn, proto.OfferCreated{Type: proto.MsgOfferCreated, Transfer: wireTransfer(created)})
			h.writeDevice(offer.To, proto.IncomingOffer{Type: proto.MsgOffer, Transfer: wireTransfer(created), From: h.device(deviceID)})
		case proto.MsgAccept:
			var accept proto.TransferID
			if err := json.Unmarshal(frame, &accept); err != nil {
				h.writeError(conn, "bad-request", "invalid accept message")
				continue
			}
			if _, err := h.transfers.Accept(accept.TransferID, deviceID); err != nil {
				h.writeError(conn, "bad-request", "transfer cannot be accepted")
				continue
			}
		case proto.MsgDecline:
			var decline proto.TransferID
			if err := json.Unmarshal(frame, &decline); err != nil {
				h.writeError(conn, "bad-request", "invalid decline message")
				continue
			}
			if _, err := h.transfers.Decline(decline.TransferID, deviceID); err != nil {
				h.writeError(conn, "bad-request", "transfer cannot be declined")
				continue
			}
		default:
			h.writeError(conn, "bad-request", "unsupported message type")
		}
	}

	if deviceID != "" {
		h.remove(deviceID)
	}
}

func (h *webSocketHub) connected(deviceID string) bool {
	h.mu.RLock()
	_, ok := h.sockets[deviceID]
	h.mu.RUnlock()
	return ok
}

func (h *webSocketHub) device(deviceID string) proto.Device {
	for _, device := range h.registry.Snapshot() {
		if device.ID == deviceID {
			return proto.Device{ID: device.ID, Name: device.Name, RawName: device.RawName, Kind: device.Kind, IsHost: device.IsHost, ConnectedAt: device.ConnectedAt}
		}
	}
	return proto.Device{}
}

func (h *webSocketHub) writeDevice(deviceID string, message any) {
	h.mu.RLock()
	conn := h.sockets[deviceID]
	h.mu.RUnlock()
	if conn != nil {
		h.write(conn, message)
	}
}

func (h *webSocketHub) notifyTransfer(event transfer.Event) {
	switch event.Type {
	case transfer.EventAccepted:
		h.writeDevice(event.To, proto.TransferID{Type: proto.MsgTransferAccepted, TransferID: event.TransferID})
	case transfer.EventDeclined:
		h.writeDevice(event.To, proto.TransferID{Type: proto.MsgTransferDeclined, TransferID: event.TransferID})
	case transfer.EventFileReady:
		h.writeDevice(event.To, proto.FileReady{Type: proto.MsgFileReady, TransferID: event.TransferID, Index: event.Index})
	case transfer.EventProgress:
		h.writeDevice(event.To, proto.Progress{Type: proto.MsgProgress, TransferID: event.TransferID, Index: event.Index, Sent: event.Sent, Size: event.Size, TotalSent: event.TotalSent, TotalSize: event.TotalSize})
	case transfer.EventDone:
		h.writeDevice(event.To, proto.TransferID{Type: proto.MsgTransferDone, TransferID: event.TransferID})
	}
}

func wireTransfer(value *transfer.Transfer) proto.Transfer {
	files := make([]proto.FileMeta, len(value.Files))
	for i, file := range value.Files {
		files[i] = proto.FileMeta{Index: file.Index, Name: file.Name, Size: file.Size, Sent: file.Sent}
	}
	return proto.Transfer{ID: value.ID, SenderID: value.SenderID, ReceiverID: value.ReceiverID, Files: files, State: string(value.State), CreatedAt: value.CreatedAt}
}

func (h *webSocketHub) joinID(conn *websocket.Conn, name, kind string) string {
	h.muting.Add(1)
	device, err := h.registry.Join(name, kind, false)
	h.muting.Add(-1)
	if err != nil {
		h.writeError(conn, "bad-request", "device name must not be empty")
		return ""
	}
	h.mu.Lock()
	h.sockets[device.ID] = conn
	h.mu.Unlock()
	h.write(conn, welcome(device))
	h.broadcastDevices(h.registry.Snapshot())
	return device.ID
}

func (h *webSocketHub) remove(deviceID string) {
	h.mu.Lock()
	delete(h.sockets, deviceID)
	h.mu.Unlock()
	h.registry.Leave(deviceID)
}

func (h *webSocketHub) broadcastDevices(snapshot []presence.Device) {
	if h.muting.Load() != 0 {
		return
	}
	devices := make([]proto.Device, len(snapshot))
	for i, device := range snapshot {
		devices[i] = proto.Device{
			ID: device.ID, Name: device.Name, RawName: device.RawName, Kind: device.Kind,
			IsHost: device.IsHost, ConnectedAt: device.ConnectedAt,
		}
	}

	h.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(h.sockets))
	for _, conn := range h.sockets {
		connections = append(connections, conn)
	}
	h.mu.RUnlock()
	for _, conn := range connections {
		h.write(conn, proto.Devices{Type: proto.MsgDevices, Devices: devices})
	}
}

func (h *webSocketHub) write(conn *websocket.Conn, message any) {
	_ = wsjson.Write(context.Background(), conn, message)
}

func (h *webSocketHub) writeError(conn *websocket.Conn, code, message string) {
	h.write(conn, proto.Error{Type: proto.MsgError, Code: code, Message: message})
}

func welcome(device *presence.Device) proto.Welcome {
	return proto.Welcome{
		Type:     proto.MsgWelcome,
		DeviceID: device.ID,
		Self:     proto.Device{ID: device.ID, Name: device.Name, RawName: device.RawName, Kind: device.Kind, IsHost: device.IsHost, ConnectedAt: device.ConnectedAt},
		IsHost:   device.IsHost,
	}
}

var androidModel = regexp.MustCompile(`Android [^;()]+;\s*([^;)]+)`) // Model is the segment after the Android version.

func deviceSuggestion(userAgent string) (suggestion, kind string) {
	if match := androidModel.FindStringSubmatch(userAgent); len(match) == 2 {
		return strings.TrimSpace(match[1]), "mobile"
	}
	switch {
	case strings.Contains(userAgent, "iPad"):
		return "iPad", "mobile"
	case strings.Contains(userAgent, "iPhone"):
		return "iPhone", "mobile"
	case strings.Contains(userAgent, "Windows"):
		return "Windows PC", "desktop"
	case strings.Contains(userAgent, "Mac OS X"):
		return "Mac", "desktop"
	case strings.Contains(userAgent, "Linux"):
		return "Linux PC", "desktop"
	default:
		return "Desktop", "desktop"
	}
}
