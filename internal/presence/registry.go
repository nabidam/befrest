// Package presence manages the hub's in-memory connected-device registry.
package presence

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	ErrInvalidName    = errors.New("invalid device name")
	ErrDeviceNotFound = errors.New("device not found")
)

// Device is a live device. A device exists exactly while its control socket is open.
type Device struct {
	ID          string
	Name        string
	RawName     string
	Kind        string
	IsHost      bool
	ConnectedAt time.Time
}

// Notifier receives a complete, independent snapshot after every registry change.
type Notifier func([]Device)

// Registry stores connected devices and maintains unique display names.
type Registry struct {
	mu         sync.Mutex
	devices    map[string]*Device
	namesInUse map[string]string
	notify     Notifier
}

// NewRegistry creates an empty registry. A nil notifier is allowed.
func NewRegistry(notify Notifier) *Registry {
	return &Registry{
		devices:    make(map[string]*Device),
		namesInUse: make(map[string]string),
		notify:     notify,
	}
}

// Join adds a newly connected device and returns its deduplicated identity.
func (r *Registry) Join(requestedName, kind string, isHost bool) (*Device, error) {
	rawName, err := validateName(requestedName)
	if err != nil {
		return nil, err
	}

	id, err := newUUIDv4()
	if err != nil {
		return nil, fmt.Errorf("mint device id: %w", err)
	}

	r.mu.Lock()
	device := &Device{
		ID:          id,
		Name:        r.deduplicateLocked(rawName, ""),
		RawName:     rawName,
		Kind:        kind,
		IsHost:      isHost,
		ConnectedAt: time.Now().UTC(),
	}
	r.devices[id] = device
	r.namesInUse[device.Name] = id
	snapshot := r.snapshotLocked()
	result := cloneDevice(device)
	r.mu.Unlock()

	r.fanout(snapshot)
	return &result, nil
}

// Rename changes a device's requested name and returns its deduplicated identity.
func (r *Registry) Rename(id, requestedName string) (*Device, error) {
	rawName, err := validateName(requestedName)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	device, ok := r.devices[id]
	if !ok {
		r.mu.Unlock()
		return nil, ErrDeviceNotFound
	}
	delete(r.namesInUse, device.Name)
	device.Name = r.deduplicateLocked(rawName, id)
	device.RawName = rawName
	r.namesInUse[device.Name] = id
	snapshot := r.snapshotLocked()
	result := cloneDevice(device)
	r.mu.Unlock()

	r.fanout(snapshot)
	return &result, nil
}

// Leave removes a disconnected device. It reports whether the device was present.
func (r *Registry) Leave(id string) bool {
	r.mu.Lock()
	device, ok := r.devices[id]
	if !ok {
		r.mu.Unlock()
		return false
	}
	delete(r.devices, id)
	delete(r.namesInUse, device.Name)
	snapshot := r.snapshotLocked()
	r.mu.Unlock()

	r.fanout(snapshot)
	return true
}

// Snapshot returns an independent, deterministic view of all connected devices.
func (r *Registry) Snapshot() []Device {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.snapshotLocked()
}

func (r *Registry) fanout(snapshot []Device) {
	if r.notify != nil {
		r.notify(snapshot)
	}
}

func (r *Registry) deduplicateLocked(rawName, selfID string) string {
	for number := 1; ; number++ {
		suffix := ""
		if number > 1 {
			suffix = fmt.Sprintf(" (%d)", number)
		}
		candidate := clampRunes(rawName, 32-utf8.RuneCountInString(suffix)) + suffix
		owner, inUse := r.namesInUse[candidate]
		if !inUse || owner == selfID {
			return candidate
		}
	}
}

func (r *Registry) snapshotLocked() []Device {
	snapshot := make([]Device, 0, len(r.devices))
	for _, device := range r.devices {
		snapshot = append(snapshot, cloneDevice(device))
	}
	sort.Slice(snapshot, func(i, j int) bool {
		if snapshot[i].ConnectedAt.Equal(snapshot[j].ConnectedAt) {
			return snapshot[i].ID < snapshot[j].ID
		}
		return snapshot[i].ConnectedAt.Before(snapshot[j].ConnectedAt)
	})
	return snapshot
}

func cloneDevice(device *Device) Device { return *device }

func validateName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", ErrInvalidName
	}
	return clampRunes(trimmed, 32), nil
}

func clampRunes(value string, maximum int) string {
	if utf8.RuneCountInString(value) <= maximum {
		return value
	}
	return string([]rune(value)[:maximum])
}

func newUUIDv4() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16]), nil
}
