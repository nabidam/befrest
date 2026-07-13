package server

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/nabidam/befrest/internal/presence"
	"github.com/nabidam/befrest/internal/proto"
	"github.com/nabidam/befrest/internal/transfer"
)

const (
	heartbeatInterval = 15 * time.Second
	heartbeatTimeout  = 15 * time.Second
)

// webSocketHub translates control-socket frames into presence registry calls.
type webSocketHub struct {
	registry         *presence.Registry
	transfers        *transfer.Manager
	mu               sync.RWMutex
	sockets          map[string]*websocket.Conn
	muting           atomic.Int32
	hostToken        string
	hostName         string
	invite           proto.InviteInfo
	interfaceChoices []proto.InterfaceChoice
	pickInterface    func(string) (proto.InviteInfo, error)
}

func newWebSocketHub(configs ...Config) *webSocketHub {
	config := Config{}
	if len(configs) > 0 {
		config = configs[0]
	}
	hub := &webSocketHub{
		sockets:          make(map[string]*websocket.Conn),
		hostToken:        config.HostToken,
		hostName:         config.HostName,
		invite:           config.Invite,
		interfaceChoices: append([]proto.InterfaceChoice(nil), config.InterfaceChoices...),
		pickInterface:    config.PickInterface,
	}
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var hello proto.Hello
	if err := wsjson.Read(ctx, conn, &hello); err != nil || hello.Type != proto.MsgHello {
		h.writeError(conn, "bad-request", "first message must be hello")
		return
	}

	name, kind := deviceSuggestion(request.UserAgent())
	var deviceID string
	isHost := h.consumeHostToken(hello.HostToken)
	if isHost {
		deviceID = h.joinID(conn, h.hostName, kind, true)
		if deviceID == "" {
			return
		}
	} else if strings.TrimSpace(hello.Name) == "" {
		h.write(conn, proto.NeedName{Type: proto.MsgNeedName, Suggested: name})
	} else {
		deviceID = h.joinID(conn, hello.Name, kind, false)
		if deviceID == "" {
			return
		}
	}

	go h.heartbeat(ctx, conn)

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
				deviceID = h.joinID(conn, setName.Name, kind, false)
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
			h.writeInvite(conn)
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
		case proto.MsgOfferCancel:
			var cancel proto.TransferID
			if err := json.Unmarshal(frame, &cancel); err != nil || h.transfers.CancelOffer(cancel.TransferID, deviceID) != nil {
				h.writeError(conn, "bad-request", "offer cannot be cancelled")
			}
		case proto.MsgTransferCancel:
			var cancel proto.TransferID
			if err := json.Unmarshal(frame, &cancel); err != nil || h.transfers.CancelTransfer(cancel.TransferID, deviceID) != nil {
				h.writeError(conn, "bad-request", "transfer cannot be cancelled")
			}
		case proto.MsgPickInterface:
			var pick proto.PickInterface
			if err := json.Unmarshal(frame, &pick); err != nil || !h.isHost(deviceID) || h.pickInterface == nil {
				h.writeError(conn, "bad-request", "interface cannot be selected")
				continue
			}
			invite, err := h.pickInterface(pick.InterfaceID)
			if err != nil {
				h.writeError(conn, "bad-request", "unknown network interface")
				continue
			}
			h.mu.Lock()
			h.invite = invite
			h.mu.Unlock()
			h.broadcastInvite()
		default:
			h.writeError(conn, "bad-request", "unsupported message type")
		}
	}

	if deviceID != "" {
		h.remove(deviceID)
	}
}

// heartbeat keeps liveness tied to the control socket even when neither side
// has application frames to exchange. Two unanswered pings close the socket.
func (h *webSocketHub) heartbeat(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	misses := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, heartbeatTimeout)
			err := conn.Ping(pingCtx)
			cancel()
			if err == nil {
				misses = 0
				continue
			}
			misses++
			if misses >= 2 {
				_ = conn.Close(websocket.StatusGoingAway, "heartbeat timed out")
				return
			}
		}
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
	case transfer.EventOfferCancelled:
		h.writeDevice(event.To, proto.OfferCancelled{Type: proto.MsgOfferCancelled, TransferID: event.TransferID, Reason: event.Reason})
	case transfer.EventFailed:
		h.writeDevice(event.To, proto.TransferFailed{Type: proto.MsgTransferFailed, TransferID: event.TransferID, Reason: event.Reason})
	}
}

func wireTransfer(value *transfer.Transfer) proto.Transfer {
	files := make([]proto.FileMeta, len(value.Files))
	for i, file := range value.Files {
		files[i] = proto.FileMeta{Index: file.Index, Name: file.Name, Size: file.Size, Sent: file.Sent}
	}
	return proto.Transfer{ID: value.ID, SenderID: value.SenderID, ReceiverID: value.ReceiverID, Files: files, State: string(value.State), CreatedAt: value.CreatedAt}
}

func (h *webSocketHub) joinID(conn *websocket.Conn, name, kind string, isHost bool) string {
	h.muting.Add(1)
	device, err := h.registry.Join(name, kind, isHost)
	h.muting.Add(-1)
	if err != nil {
		h.writeError(conn, "bad-request", "device name must not be empty")
		return ""
	}
	h.mu.Lock()
	h.sockets[device.ID] = conn
	h.mu.Unlock()
	h.write(conn, welcome(device))
	h.writeInvite(conn)
	h.writeInterfaceChoices(conn, device.ID)
	h.broadcastDevices(h.registry.Snapshot())
	return device.ID
}

func (h *webSocketHub) consumeHostToken(token string) bool {
	if token == "" {
		return false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.hostToken != token {
		return false
	}
	h.hostToken = ""
	return true
}

func (h *webSocketHub) writeInvite(conn *websocket.Conn) {
	h.mu.RLock()
	invite := h.invite
	h.mu.RUnlock()
	if invite.Port != 0 {
		h.write(conn, invite)
	}
}

func (h *webSocketHub) writeInterfaceChoices(conn *websocket.Conn, deviceID string) {
	if !h.isHost(deviceID) || len(h.interfaceChoices) == 0 {
		return
	}
	h.write(conn, proto.InterfaceChoices{Type: proto.MsgInterfaceChoices, Choices: h.interfaceChoices, Preselected: h.interfaceChoices[0].ID})
}

func (h *webSocketHub) broadcastInvite() {
	h.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(h.sockets))
	for _, conn := range h.sockets {
		connections = append(connections, conn)
	}
	invite := h.invite
	h.mu.RUnlock()
	for _, conn := range connections {
		h.write(conn, invite)
	}
}

func (h *webSocketHub) isHost(deviceID string) bool {
	for _, device := range h.registry.Snapshot() {
		if device.ID == deviceID {
			return device.IsHost
		}
	}
	return false
}

func (h *webSocketHub) remove(deviceID string) {
	h.mu.Lock()
	delete(h.sockets, deviceID)
	h.mu.Unlock()
	h.transfers.Disconnect(deviceID)
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
