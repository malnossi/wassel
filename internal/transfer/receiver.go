package transfer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// ReceiverSession represents an active inbound file transfer.
// It reads file data directly from the sender's TCP connection
// (the same connection used for the handshake).
type ReceiverSession struct {
	ID         string
	Conn       net.Conn // TCP connection from the sender
	FileSize   int64
	SavePath   string
	Checksum   string // Expected SHA-256 hex digest from the handshake payload
	OnProgress func(current, total int64)
	OnStatus   func(status string, err error)

	cancel context.CancelFunc
	mu     sync.Mutex
	done   bool
}

// NewReceiverSession creates a receiver session that reads file data directly
// from the sender's TCP connection (the same connection used for the handshake).
func NewReceiverSession(id string, conn net.Conn, savePath string, fileSize int64, checksum string) *ReceiverSession {
	return &ReceiverSession{
		ID:       id,
		Conn:     conn,
		SavePath: savePath,
		FileSize: fileSize,
		Checksum: checksum,
	}
}

// Start begins the download in a background goroutine.
func (r *ReceiverSession) Start(ctx context.Context) {
	ctx, r.cancel = context.WithCancel(ctx)

	go func() {
		r.emitStatus(StatusTransferring, nil)

		err := r.download(ctx)
		if err != nil {
			r.emitStatus(StatusFailed, err)
			// Clean up partially downloaded file
			os.Remove(r.SavePath)
			return
		}

		r.emitStatus(StatusCompleted, nil)
	}()
}

// Stop cancels the download by closing the connection.
func (r *ReceiverSession) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.done {
		return
	}
	r.done = true

	if r.cancel != nil {
		r.cancel()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}

func (r *ReceiverSession) download(ctx context.Context) error {
	defer r.Conn.Close()

	// Clear any lingering deadlines from the handshake phase
	r.Conn.SetReadDeadline(time.Time{})
	r.Conn.SetWriteDeadline(time.Time{})

	outFile, err := os.Create(r.SavePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// SECURITY: Limit the download to the declared file size + small margin
	// Prevents a malicious sender from streaming unlimited data
	maxBytes := r.FileSize + 1024 // 1KB margin
	limitedReader := io.LimitReader(r.Conn, maxBytes)

	// Use TeeReader to compute SHA-256 while downloading (O(1) memory)
	hasher := sha256.New()
	teeReader := io.TeeReader(limitedReader, hasher)

	pw := NewProgressWriter(outFile, r.FileSize, 150*time.Millisecond, func(current, total int64) {
		if r.OnProgress != nil {
			r.OnProgress(current, total)
		}
	})

	written, err := io.Copy(pw, teeReader)
	if err != nil {
		return err
	}

	// Validate received byte count matches expected file size
	if written != r.FileSize {
		return fmt.Errorf("size mismatch: expected %d bytes, received %d bytes", r.FileSize, written)
	}

	// INTEGRITY: Verify SHA-256 checksum if the sender provided one
	if r.Checksum != "" {
		computedHash := hex.EncodeToString(hasher.Sum(nil))
		if computedHash != r.Checksum {
			return fmt.Errorf("integrity check failed: expected checksum %s, got %s", r.Checksum, computedHash)
		}
	}

	return nil
}

func (r *ReceiverSession) emitStatus(status string, err error) {
	r.mu.Lock()
	if r.done {
		r.mu.Unlock()
		return
	}
	r.mu.Unlock()

	if r.OnStatus != nil {
		r.OnStatus(status, err)
	}
}
