package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"localshare/internal/transfer"
)

// Scanner performs TCP-based subnet scanning as a fallback discovery mechanism
// for environments where mDNS is unreliable.
type Scanner struct {
	store       *PeerStore
	onPeerFound func(Peer)
}

// NewScanner creates a subnet scanner bound to the given peer store.
func NewScanner(store *PeerStore, onPeerFound func(Peer)) *Scanner {
	return &Scanner{
		store:       store,
		onPeerFound: onPeerFound,
	}
}

// ScanSubnet scans the local /24 subnet(s) for LocalShare instances.
// Uses a bounded worker pool (max 64 concurrent connections) instead of unbounded goroutines.
func (s *Scanner) ScanSubnet(ctx context.Context) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	var ips []string
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue
		}

		ip := ipNet.IP.To4()
		for i := 1; i <= 254; i++ {
			if byte(i) == ip[3] {
				continue
			}
			ips = append(ips, fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], i))
		}
	}

	if len(ips) == 0 {
		return
	}

	// Bounded worker pool — max 64 concurrent TCP connections
	const maxWorkers = 64
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, ip := range ips {
		select {
		case <-ctx.Done():
			break
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(targetIP string) {
			defer wg.Done()
			defer func() { <-sem }()

			s.probeHost(ctx, targetIP)
		}(ip)
	}

	wg.Wait()
}

// probeHost connects to a single host on port 9998, sends a ping, and processes the response.
func (s *Scanner) probeHost(ctx context.Context, targetIP string) {
	dialer := net.Dialer{Timeout: 1200 * time.Millisecond}
	conn, err := dialer.DialContext(ctx, "tcp", targetIP+":9998")
	if err != nil {
		return
	}
	defer conn.Close()

	payload := transfer.HandshakePayload{
		ID: "ping",
	}
	conn.SetWriteDeadline(time.Now().Add(800 * time.Millisecond))
	if err := json.NewEncoder(conn).Encode(payload); err != nil {
		return
	}

	conn.SetReadDeadline(time.Now().Add(1200 * time.Millisecond))
	var resp transfer.PingResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return
	}

	if resp.DeviceName != "" {
		peer := Peer{
			Hostname: resp.DeviceName,
			IP:       targetIP,
			Port:     9998,
		}
		result := s.store.Upsert(peer)
		if result.WasModified && s.onPeerFound != nil {
			s.onPeerFound(result.Peer)
		}
	}
}

// StartBackgroundScanner runs periodic subnet scans every 12 seconds.
func (s *Scanner) StartBackgroundScanner(ctx context.Context) {
	ticker := time.NewTicker(12 * time.Second)
	defer ticker.Stop()

	// Initial sweep on startup
	go s.ScanSubnet(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go s.ScanSubnet(ctx)
		}
	}
}
