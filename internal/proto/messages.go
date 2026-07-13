// Package proto defines the JSON frames exchanged over Befrest control sockets.
package proto

import "time"

const (
	MsgHello    = "hello"
	MsgSetName  = "set-name"
	MsgWelcome  = "welcome"
	MsgNeedName = "need-name"
	MsgDevices  = "devices"
	MsgError    = "error"
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
