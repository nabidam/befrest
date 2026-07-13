// Package transfer owns the lifecycle and byte flow of active transfers.
package transfer

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

var (
	ErrNotFound    = errors.New("transfer not found")
	ErrWrongState  = errors.New("transfer in wrong state")
	ErrNotReceiver = errors.New("transfer receiver mismatch")
	ErrInvalidFile = errors.New("invalid transfer file")
)

type TransferState string

const (
	StateOffered   TransferState = "offered"
	StateAccepted  TransferState = "accepted"
	StateStreaming TransferState = "streaming"
	StateDone      TransferState = "done"
	StateDeclined  TransferState = "declined"
)

type FileMeta struct {
	Index int
	Name  string
	Size  int64
	Sent  int64
}
type Transfer struct {
	ID, SenderID, ReceiverID string
	Files                    []FileMeta
	State                    TransferState
	CreatedAt                time.Time
}

// Event is emitted for transfer state changes and progress. Server translates it
// into proto frames so this package stays independent of WebSocket details.
type Event struct {
	Type                             string
	To                               string
	TransferID                       string
	Index                            int
	Sent, Size, TotalSent, TotalSize int64
}

const (
	EventAccepted  = "accepted"
	EventDeclined  = "declined"
	EventFileReady = "file-ready"
	EventProgress  = "progress"
	EventDone      = "done"
)

type activeFile struct {
	reader                         *Reader
	writer                         *Writer
	uploadStarted, downloadStarted bool
}

type Manager struct {
	mu         sync.Mutex
	transfers  map[string]*Transfer
	files      map[string]map[int]*activeFile
	notify     func(Event)
	now        func() time.Time
	terminalAt map[string]time.Time
}

func NewManager(notify func(Event)) *Manager { return NewManagerWithClock(notify, time.Now) }
func NewManagerWithClock(notify func(Event), now func() time.Time) *Manager {
	return &Manager{transfers: make(map[string]*Transfer), files: make(map[string]map[int]*activeFile), terminalAt: make(map[string]time.Time), notify: notify, now: now}
}

func (m *Manager) Offer(senderID, receiverID string, files []FileMeta) (*Transfer, error) {
	if senderID == "" || receiverID == "" || senderID == receiverID || len(files) == 0 {
		return nil, ErrInvalidFile
	}
	copyFiles := make([]FileMeta, len(files))
	var total int64
	for i, file := range files {
		if file.Name == "" || file.Size < 0 {
			return nil, ErrInvalidFile
		}
		file.Index = i
		file.Sent = 0
		copyFiles[i] = file
		total += file.Size
		if total < 0 {
			return nil, ErrInvalidFile
		}
	}
	id, err := newUUID()
	if err != nil {
		return nil, fmt.Errorf("mint transfer id: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupLocked()
	transfer := &Transfer{ID: id, SenderID: senderID, ReceiverID: receiverID, Files: copyFiles, State: StateOffered, CreatedAt: m.now().UTC()}
	m.transfers[id] = transfer
	m.files[id] = make(map[int]*activeFile)
	return cloneTransfer(transfer), nil
}

func (m *Manager) Accept(id, receiverID string) (*Transfer, error) {
	m.mu.Lock()
	transfer, err := m.forReceiverLocked(id, receiverID)
	if err == nil && transfer.State == StateOffered {
		transfer.State = StateAccepted
	} else if err == nil {
		err = ErrWrongState
	}
	result := cloneTransfer(transfer)
	m.mu.Unlock()
	if err == nil {
		m.emit(Event{Type: EventAccepted, To: result.SenderID, TransferID: result.ID})
	}
	return result, err
}
func (m *Manager) Decline(id, receiverID string) (*Transfer, error) {
	m.mu.Lock()
	transfer, err := m.forReceiverLocked(id, receiverID)
	if err == nil && transfer.State == StateOffered {
		transfer.State = StateDeclined
		m.terminalAt[id] = m.now().Add(time.Minute)
	} else if err == nil {
		err = ErrWrongState
	}
	result := cloneTransfer(transfer)
	m.mu.Unlock()
	if err == nil {
		m.emit(Event{Type: EventDeclined, To: result.SenderID, TransferID: result.ID})
	}
	return result, err
}

// Upload copies an accepted file into its bounded relay. It blocks naturally
// until the receiver attaches through Download.
func (m *Manager) Upload(id string, index int, source io.Reader) error {
	m.mu.Lock()
	transfer, file, active, err := m.beginUploadLocked(id, index)
	m.mu.Unlock()
	if err != nil {
		return err
	}
	m.emit(Event{Type: EventFileReady, To: transfer.ReceiverID, TransferID: id, Index: index})
	buffer := make([]byte, 128*1024)
	var lastSent int64
	lastAt := m.now()
	for {
		count, readErr := source.Read(buffer)
		if count > 0 {
			if _, writeErr := active.writer.Write(buffer[:count]); writeErr != nil {
				_ = active.writer.Close()
				return writeErr
			}
			m.mu.Lock()
			file.Sent += int64(count)
			sent, totalSent, totalSize := file.Sent, totalSent(transfer), totalSize(transfer)
			m.mu.Unlock()
			now := m.now()
			if sent == file.Size || sent-lastSent >= max64(1, file.Size/100) || now.Sub(lastAt) >= 500*time.Millisecond {
				m.emit(Event{Type: EventProgress, To: transfer.SenderID, TransferID: id, Index: index, Sent: sent, Size: file.Size, TotalSent: totalSent, TotalSize: totalSize})
				m.emit(Event{Type: EventProgress, To: transfer.ReceiverID, TransferID: id, Index: index, Sent: sent, Size: file.Size, TotalSent: totalSent, TotalSize: totalSize})
				lastSent, lastAt = sent, now
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = active.writer.Close()
			return readErr
		}
	}
	if file.Sent != file.Size {
		_ = active.writer.Close()
		return fmt.Errorf("uploaded %d bytes, want %d: %w", file.Sent, file.Size, ErrInvalidFile)
	}
	_ = active.writer.Close()
	m.mu.Lock()
	if allComplete(transfer) {
		transfer.State = StateDone
		m.terminalAt[id] = m.now().Add(time.Minute)
	}
	done := transfer.State == StateDone
	sender, receiver := transfer.SenderID, transfer.ReceiverID
	m.mu.Unlock()
	if done {
		m.emit(Event{Type: EventDone, To: sender, TransferID: id})
		m.emit(Event{Type: EventDone, To: receiver, TransferID: id})
	}
	return nil
}

func (m *Manager) Download(id string, index int, destination io.Writer) (*FileMeta, error) {
	m.mu.Lock()
	transfer := m.transfers[id]
	if transfer == nil {
		m.mu.Unlock()
		return nil, ErrNotFound
	}
	active := m.files[id][index]
	if active == nil || active.downloadStarted {
		m.mu.Unlock()
		return nil, ErrWrongState
	}
	active.downloadStarted = true
	file := &transfer.Files[index]
	result := *file
	m.mu.Unlock()
	_, err := io.Copy(destination, active.reader)
	_ = active.reader.Close()
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Ready reports whether the receiver can attach to a file relay without
// consuming its one-shot download slot.
func (m *Manager) Ready(id string, index int) (*FileMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	transfer := m.transfers[id]
	if transfer == nil {
		return nil, ErrNotFound
	}
	if index < 0 || index >= len(transfer.Files) {
		return nil, ErrNotFound
	}
	active := m.files[id][index]
	if active == nil || active.downloadStarted {
		return nil, ErrWrongState
	}
	file := transfer.Files[index]
	return &file, nil
}

func (m *Manager) File(id string, index int) (*Transfer, *FileMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	transfer := m.transfers[id]
	if transfer == nil {
		return nil, nil, ErrNotFound
	}
	if index < 0 || index >= len(transfer.Files) {
		return nil, nil, ErrNotFound
	}
	result := cloneTransfer(transfer)
	file := result.Files[index]
	return result, &file, nil
}
func (m *Manager) Cleanup() { m.mu.Lock(); m.cleanupLocked(); m.mu.Unlock() }
func (m *Manager) cleanupLocked() {
	now := m.now()
	for id, due := range m.terminalAt {
		if !now.Before(due) {
			delete(m.transfers, id)
			delete(m.files, id)
			delete(m.terminalAt, id)
		}
	}
}
func (m *Manager) beginUploadLocked(id string, index int) (*Transfer, *FileMeta, *activeFile, error) {
	m.cleanupLocked()
	transfer := m.transfers[id]
	if transfer == nil {
		return nil, nil, nil, ErrNotFound
	}
	if transfer.State != StateAccepted && transfer.State != StateStreaming {
		return nil, nil, nil, ErrWrongState
	}
	if index < 0 || index >= len(transfer.Files) {
		return nil, nil, nil, ErrNotFound
	}
	active := m.files[id][index]
	if active != nil && active.uploadStarted {
		return nil, nil, nil, ErrWrongState
	}
	reader, writer := NewPipe()
	active = &activeFile{reader: reader, writer: writer, uploadStarted: true}
	m.files[id][index] = active
	transfer.State = StateStreaming
	return transfer, &transfer.Files[index], active, nil
}
func (m *Manager) forReceiverLocked(id, receiverID string) (*Transfer, error) {
	m.cleanupLocked()
	transfer := m.transfers[id]
	if transfer == nil {
		return nil, ErrNotFound
	}
	if transfer.ReceiverID != receiverID {
		return nil, ErrNotReceiver
	}
	return transfer, nil
}
func (m *Manager) emit(event Event) {
	if m.notify != nil {
		m.notify(event)
	}
}
func cloneTransfer(value *Transfer) *Transfer {
	if value == nil {
		return nil
	}
	clone := *value
	clone.Files = append([]FileMeta(nil), value.Files...)
	return &clone
}
func allComplete(transfer *Transfer) bool {
	for _, file := range transfer.Files {
		if file.Sent != file.Size {
			return false
		}
	}
	return true
}
func totalSent(transfer *Transfer) int64 {
	var total int64
	for _, file := range transfer.Files {
		total += file.Sent
	}
	return total
}
func totalSize(transfer *Transfer) int64 {
	var total int64
	for _, file := range transfer.Files {
		total += file.Size
	}
	return total
}
func max64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
