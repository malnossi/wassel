package transfer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// ReceiverSession represents an active inbound file transfer.
type ReceiverSession struct {
	ID             string
	DownloadURL    string
	FileSize       int64
	SavePath       string
	ExpectedPeerIP string // The peer's IP from the TCP handshake — used for URL validation
	Checksum       string // Expected SHA-256 hex digest from the handshake payload
	OnProgress     func(current, total int64)
	OnStatus       func(status string, err error)

	cancel context.CancelFunc
}

// NewReceiverSession creates a receiver session with download URL validation.
func NewReceiverSession(id, downloadURL, savePath string, fileSize int64, expectedPeerIP, checksum string) (*ReceiverSession, error) {
	// SECURITY: Validate that the download URL host matches the sender's IP
	parsed, err := url.Parse(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("invalid download URL: %w", err)
	}

	urlHost := parsed.Hostname()
	if urlHost != expectedPeerIP {
		return nil, fmt.Errorf("download URL host %q does not match peer IP %q — possible URL injection attack", urlHost, expectedPeerIP)
	}

	// Only allow http scheme for local network transfers
	if parsed.Scheme != "http" {
		return nil, fmt.Errorf("download URL uses disallowed scheme %q — only http is permitted", parsed.Scheme)
	}

	return &ReceiverSession{
		ID:             id,
		DownloadURL:    downloadURL,
		SavePath:       savePath,
		FileSize:       fileSize,
		ExpectedPeerIP: expectedPeerIP,
		Checksum:       checksum,
	}, nil
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

// Stop cancels the download.
func (r *ReceiverSession) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *ReceiverSession) download(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", r.DownloadURL, nil)
	if err != nil {
		return err
	}

	// Use a custom client with timeouts instead of http.DefaultClient
	client := &http.Client{
		Timeout: 0, // No overall timeout — large files can take a long time
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ResponseHeaderTimeout: 15 * time.Second,
			IdleConnTimeout:       30 * time.Second,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http error: status code %d", resp.StatusCode)
	}

	outFile, err := os.Create(r.SavePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// SECURITY: Limit the download to the declared file size + small margin
	// Prevents a malicious sender from streaming unlimited data
	maxBytes := r.FileSize + 1024 // 1KB margin for HTTP overhead
	limitedBody := io.LimitReader(resp.Body, maxBytes)

	// Use TeeReader to compute SHA-256 while downloading (O(1) memory)
	hasher := sha256.New()
	teeReader := io.TeeReader(limitedBody, hasher)

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
	if r.OnStatus != nil {
		r.OnStatus(status, err)
	}
}
