package transfer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SenderSession represents an active outbound file transfer.
type SenderSession struct {
	ID         string
	FilePath   string
	FileSize   int64
	FileName   string
	Checksum   string // SHA-256 hex digest
	TargetIP   string
	TargetPort int
	HTTPServer *http.Server
	Listener   net.Listener
	OnProgress func(current int64, total int64)
	OnStatus   func(status string, err error)

	cancel context.CancelFunc
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

// Start spins up the ephemeral HTTP server and executes the handshake.
func (s *SenderSession) Start(ctx context.Context, localIP string) error {
	ctx, s.cancel = context.WithCancel(ctx)

	// 1. Start ephemeral HTTP server on :0 (random open port)
	var err error
	s.Listener, err = net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	port := s.Listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/download/%s", s.ID), func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open(s.FilePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			s.emitStatus(StatusFailed, err)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Length", fmt.Sprintf("%d", s.FileSize))
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", s.FileName))
		w.Header().Set("Content-Type", "application/octet-stream")

		pw := NewProgressWriter(w, s.FileSize, 150*time.Millisecond, func(current, total int64) {
			if s.OnProgress != nil {
				s.OnProgress(current, total)
			}
		})

		_, err = io.Copy(pw, file)
		if err != nil {
			s.emitStatus(StatusFailed, err)
			return
		}

		s.emitStatus(StatusCompleted, nil)

		// Self-shutdown server after serving, with delay for TCP buffer flush
		go func() {
			time.Sleep(1 * time.Second)
			s.Stop()
		}()
	})

	s.HTTPServer = &http.Server{Handler: mux}

	go func() {
		if err := s.HTTPServer.Serve(s.Listener); err != nil && err != http.ErrServerClosed {
			s.emitStatus(StatusFailed, err)
		}
	}()

	// 2. Perform the TCP Handshake in a background routine
	go func() {
		err := s.performHandshake(ctx, localIP, port)
		if err != nil {
			s.Stop()
			s.emitStatus(StatusFailed, err)
		}
	}()

	return nil
}

func (s *SenderSession) performHandshake(ctx context.Context, localIP string, httpPort int) error {
	downloadURL := fmt.Sprintf("http://%s:%d/download/%s", localIP, httpPort, s.ID)

	// Dial recipient control TCP port with context awareness
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", s.TargetIP, s.TargetPort))
	if err != nil {
		return fmt.Errorf("failed to connect to peer control channel: %w", err)
	}
	defer conn.Close()

	payload := HandshakePayload{
		ID:          s.ID,
		Filename:    s.FileName,
		Size:        s.FileSize,
		DownloadURL: downloadURL,
		Checksum:    s.Checksum,
	}

	// Write Handshake Payload JSON
	if err := json.NewEncoder(conn).Encode(payload); err != nil {
		return fmt.Errorf("failed to send handshake metadata: %w", err)
	}

	// Wait up to 60s for approval response
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	var resp HandshakeResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return fmt.Errorf("peer disconnected or failed to approve: %w", err)
	}

	if !resp.Accepted {
		s.emitStatus(StatusDeclined, nil)
		return nil
	}

	s.emitStatus(StatusTransferring, nil)
	return nil
}

// Stop gracefully shuts down the HTTP server and cancels the session.
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
	if s.HTTPServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s.HTTPServer.Shutdown(ctx)
	}
	if s.Listener != nil {
		s.Listener.Close()
	}
}

func (s *SenderSession) emitStatus(status string, err error) {
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
