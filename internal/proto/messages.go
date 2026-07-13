// Package proto defines the JSON frames exchanged over Befrest control sockets.
package proto

import "time"

const (
	MsgHello            = "hello"
	MsgSetName          = "set-name"
	MsgWelcome          = "welcome"
	MsgNeedName         = "need-name"
	MsgDevices          = "devices"
	MsgError            = "error"
	MsgOffer            = "offer"
	MsgOfferCreated     = "offer-created"
	MsgAccept           = "accept"
	MsgDecline          = "decline"
	MsgTransferAccepted = "transfer-accepted"
	MsgTransferDeclined = "transfer-declined"
	MsgFileReady        = "file-ready"
	MsgProgress         = "progress"
	MsgTransferDone     = "transfer-done"
)

// Device is the wire representation of a connected device.
type Device struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	RawName     string    `json:"rawName"`
	Kind        string    `json:"kind"`
	IsHost      bool      `json:"isHost"`
	ConnectedAt time.Time `json:"connectedAt"`
}

type Hello struct {
	Type      string `json:"type"`
	DeviceID  string `json:"deviceId,omitempty"`
	Name      string `json:"name,omitempty"`
	HostToken string `json:"hostToken,omitempty"`
}

type SetName struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type Welcome struct {
	Type     string `json:"type"`
	DeviceID string `json:"deviceId"`
	Self     Device `json:"self"`
	IsHost   bool   `json:"isHost"`
}

type NeedName struct {
	Type      string `json:"type"`
	Suggested string `json:"suggested"`
}

type Devices struct {
	Type    string   `json:"type"`
	Devices []Device `json:"devices"`
}

type Error struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type FileMeta struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Sent  int64  `json:"sent"`
}

type Transfer struct {
	ID         string     `json:"id"`
	SenderID   string     `json:"senderId"`
	ReceiverID string     `json:"receiverId"`
	Files      []FileMeta `json:"files"`
	State      string     `json:"state"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type OfferRequest struct {
	Type  string     `json:"type"`
	To    string     `json:"to"`
	Files []FileMeta `json:"files"`
}

type TransferID struct {
	Type       string `json:"type"`
	TransferID string `json:"transferId"`
}

type OfferCreated struct {
	Type     string   `json:"type"`
	Transfer Transfer `json:"transfer"`
}
type IncomingOffer struct {
	Type     string   `json:"type"`
	Transfer Transfer `json:"transfer"`
	From     Device   `json:"from"`
}
type FileReady struct {
	Type       string `json:"type"`
	TransferID string `json:"transferId"`
	Index      int    `json:"index"`
}
type Progress struct {
	Type       string `json:"type"`
	TransferID string `json:"transferId"`
	Index      int    `json:"index"`
	Sent       int64  `json:"sent"`
	Size       int64  `json:"size"`
	TotalSent  int64  `json:"totalSent"`
	TotalSize  int64  `json:"totalSize"`
}
