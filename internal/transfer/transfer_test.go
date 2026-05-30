package transfer

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestProgressWriterThrottling(t *testing.T) {
	var buf bytes.Buffer
	progressCount := 0
	lastProgressVal := int64(0)

	pw := NewProgressWriter(&buf, 100, 100*time.Millisecond, func(current, total int64) {
		progressCount++
		lastProgressVal = current
	})

	// 1. Initial write
	_, err := pw.Write([]byte("abcd")) // 4 bytes
	if err != nil {
		t.Fatal(err)
	}

	// Should fire immediately on first write
	if progressCount != 1 {
		t.Errorf("Expected progress callback to fire immediately on first write, got %d", progressCount)
	}
	if lastProgressVal != 4 {
		t.Errorf("Expected progress value to be 4, got %d", lastProgressVal)
	}

	// 2. Immediate write (throttled)
	_, _ = pw.Write([]byte("efgh")) // 4 more bytes (total 8)
	if progressCount != 1 {
		t.Errorf("Expected progress callback to be throttled, got %d", progressCount)
	}

	// 3. Write to 100% completion (should bypass throttle and notify final size)
	_, _ = pw.Write([]byte(string(make([]byte, 92)))) // 92 more bytes (total 100)
	if progressCount != 2 {
		t.Errorf("Expected progress callback to fire on 100%% completion regardless of throttle, got %d", progressCount)
	}
	if lastProgressVal != 100 {
		t.Errorf("Expected final progress value to be 100, got %d", lastProgressVal)
	}
}

func TestReceiverSessionCreation(t *testing.T) {
	// Create a pipe to simulate a TCP connection
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Test session creation with valid params
	session := NewReceiverSession(
		"transfer-123",
		server,
		"/tmp/testfile",
		1000,
		"some-checksum",
	)
	if session == nil {
		t.Fatal("Expected session to be non-nil")
	}
	if session.ID != "transfer-123" {
		t.Errorf("Expected ID to be 'transfer-123', got '%s'", session.ID)
	}
	if session.FileSize != 1000 {
		t.Errorf("Expected FileSize to be 1000, got %d", session.FileSize)
	}
	if session.Checksum != "some-checksum" {
		t.Errorf("Expected Checksum to be 'some-checksum', got '%s'", session.Checksum)
	}
}

func TestTransferManager(t *testing.T) {
	m := NewManager()

	// Register sender
	sender := &SenderSession{ID: "sender-1", FilePath: "a.txt"}
	m.AddSender("sender-1", sender)

	s, ok := m.GetSender("sender-1")
	if !ok || s.FilePath != "a.txt" {
		t.Error("Failed to retrieve registered sender session")
	}

	// Register receiver
	receiver := &ReceiverSession{ID: "receiver-1", SavePath: "b.txt"}
	m.AddReceiver("receiver-1", receiver)

	r, ok := m.GetReceiver("receiver-1")
	if !ok || r.SavePath != "b.txt" {
		t.Error("Failed to retrieve registered receiver session")
	}

	// Cancel active receiver
	cancelled := m.CancelTransfer("receiver-1")
	if !cancelled {
		t.Error("Expected CancelTransfer to return true")
	}

	_, ok = m.GetReceiver("receiver-1")
	if ok {
		t.Error("Expected receiver session to be removed from manager after cancellation")
	}

	// Cleanup sender
	m.RemoveSender("sender-1")
	_, ok = m.GetSender("sender-1")
	if ok {
		t.Error("Expected sender session to be removed from manager after RemoveSender")
	}
}
