package transfer

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"time"
)

// HandshakeCallbacks defines the interface for handling incoming handshake events.
type HandshakeCallbacks struct {
	// OnTransferRequest is called when a valid file transfer handshake arrives.
	OnTransferRequest func(conn net.Conn, payload HandshakePayload, peerIP string)
	// OnPingRequest is called when a TCP discovery ping arrives.
	OnPingRequest func(conn net.Conn)
}

// HandshakeListener manages the TCP control channel on port 9998.
type HandshakeListener struct {
	listener  net.Listener
	callbacks HandshakeCallbacks
}

// NewHandshakeListener creates and starts a TCP listener for incoming handshakes.
// Retries binding up to 5 times to handle address-in-use on rapid restarts.
func NewHandshakeListener(ctx context.Context, callbacks HandshakeCallbacks) (*HandshakeListener, error) {
	var listener net.Listener
	var err error

	for i := 0; i < 5; i++ {
		listener, err = net.Listen("tcp", ":9998")
		if err == nil {
			break
		}
		log.Printf("Failed to bind TCP control channel on :9998 (attempt %d/5): %v\n", i+1, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	if err != nil {
		return nil, err
	}

	hl := &HandshakeListener{
		listener:  listener,
		callbacks: callbacks,
	}

	// Close listener when context is cancelled
	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	// Accept loop
	go hl.acceptLoop(ctx)

	return hl, nil
}

func (hl *HandshakeListener) acceptLoop(ctx context.Context) {
	for {
		conn, err := hl.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Println("Error accepting handshake connection:", err)
				continue
			}
		}
		go hl.handleConnection(conn)
	}
}

func (hl *HandshakeListener) handleConnection(conn net.Conn) {
	// Set read timeout for safety
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// SECURITY: Limit incoming JSON payload to 4KB to prevent memory exhaustion
	limitedReader := io.LimitReader(conn, 4096)

	var payload HandshakePayload
	if err := json.NewDecoder(limitedReader).Decode(&payload); err != nil {
		log.Println("Failed to decode incoming handshake:", err)
		conn.Close()
		return
	}

	// Remove read deadline so we can hold connection during UI approval
	conn.SetReadDeadline(time.Time{})

	// Route based on payload type
	if payload.ID == "ping" {
		if hl.callbacks.OnPingRequest != nil {
			hl.callbacks.OnPingRequest(conn)
		} else {
			conn.Close()
		}
		return
	}

	// Transfer request
	remoteIP := conn.RemoteAddr().(*net.TCPAddr).IP.String()
	if hl.callbacks.OnTransferRequest != nil {
		hl.callbacks.OnTransferRequest(conn, payload, remoteIP)
	} else {
		conn.Close()
	}
}
