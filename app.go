package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"

	"localshare/internal/discovery"
	"localshare/internal/identity"
	"localshare/internal/platform"
	"localshare/internal/transfer"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ActiveHandshake holds a pending incoming transfer request awaiting user approval.
type ActiveHandshake struct {
	conn    net.Conn
	payload transfer.HandshakePayload
	peerIP  string
}

// App is the thin IPC adapter between the Wails frontend and the backend services.
// It delegates all business logic to the internal packages.
type App struct {
	ctx              context.Context
	discoveryManager *discovery.Manager
	transferManager  *transfer.Manager
	nameResolver     *identity.Resolver
	activeHandshakes map[string]*ActiveHandshake
	handshakeMu      sync.Mutex
	deviceName       string
	deviceNameMu     sync.RWMutex
}

// NewApp creates a new application instance.
func NewApp() *App {
	return &App{
		activeHandshakes: make(map[string]*ActiveHandshake),
		transferManager:  transfer.NewManager(),
		nameResolver:     identity.NewResolver(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 1. Initialize discovery manager (mDNS + subnet scanner)
	a.discoveryManager = discovery.NewManager(
		func(p discovery.Peer) {
			runtime.EventsEmit(a.ctx, "peer:discovered", p)
		},
		func(p discovery.Peer) {
			runtime.EventsEmit(a.ctx, "peer:lost", p)
		},
	)
	if err := a.discoveryManager.Start(ctx); err != nil {
		log.Println("Failed to start discovery manager:", err)
	}

	// 2. Resolve unique device name and register mDNS service
	go a.resolveNameAndRegister()

	// 3. Start TCP handshake listener
	_, err := transfer.NewHandshakeListener(ctx, transfer.HandshakeCallbacks{
		OnTransferRequest: a.handleIncomingTransfer,
		OnPingRequest:     a.handlePingRequest,
	})
	if err != nil {
		log.Println("Critical error: Failed to start handshake listener:", err)
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.discoveryManager != nil {
		a.discoveryManager.Shutdown()
	}
}

// --- Wails-Bound Methods ---

// SelectAndSendFile opens a file dialog and initiates a sending session to a peer IP.
func (a *App) SelectAndSendFile(peerIP string) string {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select File to Share",
	})
	if err != nil || filePath == "" {
		return ""
	}

	transferID := uuid.New().String()
	session, err := transfer.NewSenderSession(transferID, filePath, peerIP, 9998)
	if err != nil {
		log.Println("Failed to create sender session:", err)
		runtime.EventsEmit(a.ctx, "transfer:error", map[string]interface{}{
			"message": "Failed to prepare file: " + err.Error(),
		})
		return ""
	}

	// Dynamic IP routing discovery
	localIP, err := platform.GetOutboundIP(peerIP)
	if err != nil {
		log.Println("Failed to detect outbound IP:", err)
		runtime.EventsEmit(a.ctx, "transfer:error", map[string]interface{}{
			"message": "Cannot reach peer: " + err.Error(),
		})
		a.discoveryManager.RemovePeer(peerIP)
		return ""
	}

	peerName := a.discoveryManager.GetPeerName(peerIP)

	session.OnProgress = func(current, total int64) {
		percentage := float64(current) / float64(total) * 100
		runtime.EventsEmit(a.ctx, "transfer:progress", map[string]interface{}{
			"id":         transferID,
			"filename":   session.FileName,
			"percentage": percentage,
			"current":    current,
			"total":      total,
			"isIncoming": false,
		})
	}

	var transferMu sync.Mutex
	transferring := false

	session.OnStatus = func(status string, err error) {
		transferMu.Lock()
		if status == transfer.StatusTransferring {
			transferring = true
		}
		isTransferring := transferring
		transferMu.Unlock()

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		runtime.EventsEmit(a.ctx, "transfer:status", map[string]interface{}{
			"id":         transferID,
			"filename":   session.FileName,
			"size":       session.FileSize,
			"isIncoming": false,
			"status":     status,
			"error":      errMsg,
		})

		if status == transfer.StatusFailed && !isTransferring {
			a.discoveryManager.RemovePeer(peerIP)
		}

		if status == transfer.StatusCompleted || status == transfer.StatusFailed || status == transfer.StatusDeclined || status == transfer.StatusCancelled {
			a.transferManager.RemoveSender(transferID)
		}
	}

	a.transferManager.AddSender(transferID, session)

	// Notify frontend of initialization
	runtime.EventsEmit(a.ctx, "transfer:status", map[string]interface{}{
		"id":         transferID,
		"filename":   session.FileName,
		"size":       session.FileSize,
		"isIncoming": false,
		"status":     transfer.StatusConnecting,
		"peerName":   peerName,
	})

	err = session.Start(a.ctx, localIP)
	if err != nil {
		log.Println("Failed to start sender session:", err)
		runtime.EventsEmit(a.ctx, "transfer:error", map[string]interface{}{
			"message": "Failed to start transfer: " + err.Error(),
		})
		a.discoveryManager.RemovePeer(peerIP)
		return ""
	}

	return transferID
}

// AcceptTransfer approves an inbound handshake and opens a save dialog.
func (a *App) AcceptTransfer(id string) bool {
	// Atomically take ownership of the handshake — fixes race condition (Bug 1.1)
	a.handshakeMu.Lock()
	handshake, exists := a.activeHandshakes[id]
	if exists {
		delete(a.activeHandshakes, id) // Remove from map while holding lock
	}
	a.handshakeMu.Unlock()

	if !exists {
		return false
	}

	// 1. Open save dialog
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Received File",
		DefaultFilename: handshake.payload.Filename,
	})
	if err != nil || savePath == "" {
		// Decline if user cancels save dialog
		a.sendHandshakeResponse(handshake.conn, false)
		return false
	}

	// 2. Respond to sender TCP connection — check error BEFORE closing (Bug 1.6 fix)
	err = json.NewEncoder(handshake.conn).Encode(transfer.HandshakeResponse{Accepted: true})
	if err != nil {
		log.Println("Failed to send acceptance handshake:", err)
		handshake.conn.Close()
		return false
	}
	handshake.conn.Close()

	// 3. Initiate download stream with security validations
	session, err := transfer.NewReceiverSession(
		id,
		handshake.payload.DownloadURL,
		savePath,
		handshake.payload.Size,
		handshake.peerIP,          // For URL host validation
		handshake.payload.Checksum, // For integrity verification
	)
	if err != nil {
		log.Println("Failed to create receiver session:", err)
		runtime.EventsEmit(a.ctx, "transfer:error", map[string]interface{}{
			"message": "Security validation failed: " + err.Error(),
		})
		return false
	}

	session.OnProgress = func(current, total int64) {
		percentage := float64(current) / float64(total) * 100
		runtime.EventsEmit(a.ctx, "transfer:progress", map[string]interface{}{
			"id":         id,
			"filename":   handshake.payload.Filename,
			"percentage": percentage,
			"current":    current,
			"total":      total,
			"isIncoming": true,
		})
	}

	session.OnStatus = func(status string, err error) {
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		runtime.EventsEmit(a.ctx, "transfer:status", map[string]interface{}{
			"id":         id,
			"filename":   handshake.payload.Filename,
			"size":       handshake.payload.Size,
			"isIncoming": true,
			"status":     status,
			"error":      errMsg,
		})

		if status == transfer.StatusCompleted || status == transfer.StatusFailed || status == transfer.StatusCancelled {
			a.transferManager.RemoveReceiver(id)
		}
	}

	a.transferManager.AddReceiver(id, session)
	session.Start(a.ctx)
	return true
}

// DeclineTransfer rejects an inbound handshake.
func (a *App) DeclineTransfer(id string) bool {
	a.handshakeMu.Lock()
	handshake, exists := a.activeHandshakes[id]
	if exists {
		delete(a.activeHandshakes, id)
	}
	a.handshakeMu.Unlock()

	if !exists {
		return false
	}

	a.sendHandshakeResponse(handshake.conn, false)

	runtime.EventsEmit(a.ctx, "transfer:status", map[string]interface{}{
		"id":         id,
		"filename":   handshake.payload.Filename,
		"size":       handshake.payload.Size,
		"isIncoming": true,
		"status":     transfer.StatusDeclined,
	})

	return true
}

// CancelTransfer cancels an in-progress transfer (sender or receiver).
func (a *App) CancelTransfer(id string) bool {
	cancelled := a.transferManager.CancelTransfer(id)
	if cancelled {
		runtime.EventsEmit(a.ctx, "transfer:status", map[string]interface{}{
			"id":     id,
			"status": transfer.StatusCancelled,
		})
	}
	return cancelled
}

// GetPeers returns currently active network devices.
func (a *App) GetPeers() []discovery.Peer {
	if a.discoveryManager != nil {
		return a.discoveryManager.GetPeers()
	}
	return []discovery.Peer{}
}

// ResetDiscovery clears discovered peers and triggers fresh scanning.
func (a *App) ResetDiscovery() {
	if a.discoveryManager != nil {
		a.discoveryManager.Reset(a.ctx)
	}
}

// GetDeviceName returns the local device's display name.
func (a *App) GetDeviceName() string {
	a.deviceNameMu.RLock()
	defer a.deviceNameMu.RUnlock()
	return a.deviceName
}

// --- Internal Handlers ---

func (a *App) handleIncomingTransfer(conn net.Conn, payload transfer.HandshakePayload, peerIP string) {
	a.handshakeMu.Lock()
	a.activeHandshakes[payload.ID] = &ActiveHandshake{
		conn:    conn,
		payload: payload,
		peerIP:  peerIP,
	}
	a.handshakeMu.Unlock()

	runtime.EventsEmit(a.ctx, "transfer:request", map[string]interface{}{
		"id":       payload.ID,
		"filename": payload.Filename,
		"size":     payload.Size,
		"peerIp":   peerIP,
		"peerName": a.discoveryManager.GetPeerName(peerIP),
	})
}

func (a *App) handlePingRequest(conn net.Conn) {
	resp := transfer.PingResponse{
		DeviceName: a.GetDeviceName(),
	}
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	json.NewEncoder(conn).Encode(resp)
	conn.Close()
}

func (a *App) resolveNameAndRegister() {
	// Wait for initial mDNS sweep to populate peer names
	time.Sleep(1500 * time.Millisecond)

	takenNames := a.discoveryManager.GetPeerNames()
	chosenName := a.nameResolver.ResolveName(takenNames)

	a.deviceNameMu.Lock()
	a.deviceName = chosenName
	a.deviceNameMu.Unlock()

	log.Printf("Selected local device name: %s\n", chosenName)

	// Register mDNS service
	err := a.discoveryManager.RegisterService(chosenName, 9998)
	if err != nil {
		log.Println("Failed to register local mDNS service:", err)
	}

	// Emit device name to frontend
	runtime.EventsEmit(a.ctx, "device:name", chosenName)
}

func (a *App) sendHandshakeResponse(conn net.Conn, accepted bool) {
	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	json.NewEncoder(conn).Encode(transfer.HandshakeResponse{Accepted: accepted})
	conn.Close()
}
