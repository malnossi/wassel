package transfer

import (
	"sync"
)

// Manager tracks active sender and receiver sessions with thread-safe access.
type Manager struct {
	senders   map[string]*SenderSession
	senderMu  sync.Mutex
	receivers map[string]*ReceiverSession
	receiverMu sync.Mutex
}

// NewManager creates an empty transfer manager.
func NewManager() *Manager {
	return &Manager{
		senders:   make(map[string]*SenderSession),
		receivers: make(map[string]*ReceiverSession),
	}
}

// AddSender registers a sender session.
func (m *Manager) AddSender(id string, session *SenderSession) {
	m.senderMu.Lock()
	defer m.senderMu.Unlock()
	m.senders[id] = session
}

// RemoveSender removes a sender session.
func (m *Manager) RemoveSender(id string) {
	m.senderMu.Lock()
	defer m.senderMu.Unlock()
	delete(m.senders, id)
}

// GetSender retrieves a sender session by ID.
func (m *Manager) GetSender(id string) (*SenderSession, bool) {
	m.senderMu.Lock()
	defer m.senderMu.Unlock()
	s, ok := m.senders[id]
	return s, ok
}

// AddReceiver registers a receiver session.
func (m *Manager) AddReceiver(id string, session *ReceiverSession) {
	m.receiverMu.Lock()
	defer m.receiverMu.Unlock()
	m.receivers[id] = session
}

// RemoveReceiver removes a receiver session.
func (m *Manager) RemoveReceiver(id string) {
	m.receiverMu.Lock()
	defer m.receiverMu.Unlock()
	delete(m.receivers, id)
}

// GetReceiver retrieves a receiver session by ID.
func (m *Manager) GetReceiver(id string) (*ReceiverSession, bool) {
	m.receiverMu.Lock()
	defer m.receiverMu.Unlock()
	r, ok := m.receivers[id]
	return r, ok
}

// CancelTransfer stops and removes a transfer by ID (sender or receiver).
// Returns true if a transfer was found and cancelled.
func (m *Manager) CancelTransfer(id string) bool {
	m.senderMu.Lock()
	if sender, ok := m.senders[id]; ok {
		delete(m.senders, id)
		m.senderMu.Unlock()
		sender.Stop()
		return true
	}
	m.senderMu.Unlock()

	m.receiverMu.Lock()
	if receiver, ok := m.receivers[id]; ok {
		delete(m.receivers, id)
		m.receiverMu.Unlock()
		receiver.Stop()
		return true
	}
	m.receiverMu.Unlock()

	return false
}
