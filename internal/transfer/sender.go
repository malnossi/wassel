package transfer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SenderSession represents an active outbound file transfer.
// It dials the receiver, performs the handshake, and streams the file
// directly over the same TCP connection — no ephemeral HTTP server needed.
type SenderSession struct {
	ID         string
	FilePath   string
	FileSize   int64
	FileName   string
	Checksum   string // SHA-256 hex digest
	TargetIP   string
	TargetPort int
	OnProgress func(current int64, total int64)
	OnStatus   func(status string, err error)

	cancel context.CancelFunc
	conn   net.Conn
	mu     sync.Mutex
	done   bool
}

// NewSenderSession creates a sender session and pre-computes the file's SHA-256 checksum.
func NewSenderSession(id, filePath, targetIP string, targetPort int) (*SenderSession, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// Compute SHA-256 checksum for integrity verification
	checksum, err := computeFileChecksum(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute file checksum: %w", err)
	}

	return &SenderSession{
		ID:         id,
		FilePath:   filePath,
		FileSize:   fileInfo.Size(),
		FileName:   filepath.Base(filePath),
		Checksum:   checksum,
		TargetIP:   targetIP,
		TargetPort: targetPort,
	}, nil
}

// Start begins the transfer in a background goroutine.
func (s *SenderSession) Start(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)

	go func() {
		err := s.performTransfer(ctx)
		if err != nil {
			s.emitStatus(StatusFailed, err)
		}
	}()
}

func (s *SenderSession) performTransfer(ctx context.Context) error {
	// 1. Dial recipient control TCP port
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", s.TargetIP, s.TargetPort))
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	// Store conn for cancellation support
	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()

	defer func() {
		conn.Close()
		s.mu.Lock()
		s.conn = nil
		s.mu.Unlock()
	}()

	// 2. Send handshake metadata
	payload := HandshakePayload{
		ID:       s.ID,
		Filename: s.FileName,
		Size:     s.FileSize,
		Checksum: s.Checksum,
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := json.NewEncoder(conn).Encode(payload); err != nil {
		return fmt.Errorf("failed to send handshake metadata: %w", err)
	}

	// 3. Wait for approval response (up to 60s for user interaction)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	var resp HandshakeResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return fmt.Errorf("peer disconnected or failed to approve: %w", err)
	}

	if !resp.Accepted {
		s.emitStatus(StatusDeclined, nil)
		return nil
	}

	// 4. Stream file directly over the same TCP connection
	s.emitStatus(StatusTransferring, nil)

	// Clear all deadlines for the bulk transfer
	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})

	file, err := os.Open(s.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	pw := NewProgressWriter(conn, s.FileSize, 150*time.Millisecond, func(current, total int64) {
		if s.OnProgress != nil {
			s.OnProgress(current, total)
		}
	})

	_, err = io.Copy(pw, file)
	if err != nil {
		return fmt.Errorf("failed to stream file: %w", err)
	}

	s.emitStatus(StatusCompleted, nil)
	return nil
}

// Stop gracefully cancels the transfer and closes the connection.
func (s *SenderSession) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done {
		return
	}
	s.done = true

	if s.cancel != nil {
		s.cancel()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *SenderSession) emitStatus(status string, err error) {
	s.mu.Lock()
	if s.done {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	if s.OnStatus != nil {
		s.OnStatus(status, err)
	}
}

// computeFileChecksum returns the SHA-256 hex digest of a file.
func computeFileChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
