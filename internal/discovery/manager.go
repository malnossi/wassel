package discovery

import (
	"context"
	"log"
	"time"
)

// Manager orchestrates mDNS browsing, subnet scanning, and peer lifecycle management.
type Manager struct {
	store       *PeerStore
	mdns        *MDNSService
	scanner     *Scanner
	onPeerFound func(Peer)
	onPeerLost  func(Peer)
}

// NewManager creates a discovery manager that coordinates mDNS and subnet scanning.
func NewManager(onPeerFound func(Peer), onPeerLost func(Peer)) *Manager {
	store := NewPeerStore()

	m := &Manager{
		store:       store,
		onPeerFound: onPeerFound,
		onPeerLost:  onPeerLost,
	}

	m.mdns = NewMDNSService(store, onPeerFound)
	m.scanner = NewScanner(store, onPeerFound)

	return m
}

// Start boots up all discovery mechanisms.
func (m *Manager) Start(ctx context.Context) error {
	// 1. Start mDNS passive browsing
	err := m.mdns.Start(ctx)
	if err != nil {
		log.Println("Failed to start mDNS service:", err)
		// Continue — subnet scanner will serve as fallback
	}

	// 2. Start background active mDNS query scheduler
	go m.startBackgroundScheduler(ctx)

	// 3. Start background subnet scanner
	go m.scanner.StartBackgroundScanner(ctx)

	return err
}

// RegisterService registers our local mDNS service instance.
func (m *Manager) RegisterService(name string, port int) error {
	return m.mdns.RegisterService(name, port)
}

// GetRegisteredName returns the registered mDNS instance name.
func (m *Manager) GetRegisteredName() string {
	return m.mdns.GetRegisteredName()
}

// GetPeers returns all currently discovered peers.
func (m *Manager) GetPeers() []Peer {
	return m.store.GetAll()
}

// GetPeerNames returns the set of discovered peer hostnames.
func (m *Manager) GetPeerNames() map[string]bool {
	return m.store.GetNames()
}

// InjectPeer manually adds a peer (e.g. from subnet scanner).
func (m *Manager) InjectPeer(peer Peer) {
	result := m.store.Upsert(peer)
	if result.WasModified && m.onPeerFound != nil {
		m.onPeerFound(result.Peer)
	}
}

// RemovePeer removes a peer by IP (e.g. when a connection fails).
func (m *Manager) RemovePeer(ip string) {
	peer, existed := m.store.Remove(ip)
	if existed && m.onPeerLost != nil {
		m.onPeerLost(peer)
	}
}

// GetPeerName returns the hostname for a given IP, or the IP itself if not found.
func (m *Manager) GetPeerName(ip string) string {
	peers := m.store.GetAll()
	for _, p := range peers {
		if p.IP == ip {
			return p.Hostname
		}
	}
	return ip
}

// Reset clears all peers and triggers a fresh mDNS browse + subnet scan.
func (m *Manager) Reset(ctx context.Context) {
	evicted := m.store.Clear()
	if m.onPeerLost != nil {
		for _, peer := range evicted {
			m.onPeerLost(peer)
		}
	}

	// Trigger fresh discovery
	go m.mdns.TriggerActiveQuery(ctx, 6*time.Second)
	go m.scanner.ScanSubnet(ctx)
}

// Shutdown stops all discovery services.
func (m *Manager) Shutdown() {
	m.mdns.Shutdown()
}

// startBackgroundScheduler runs periodic active mDNS queries and stale peer sweeps.
func (m *Manager) startBackgroundScheduler(ctx context.Context) {
	queryTicker := time.NewTicker(10 * time.Second)
	defer queryTicker.Stop()

	sweepTicker := time.NewTicker(10 * time.Second)
	defer sweepTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-queryTicker.C:
			go m.mdns.TriggerActiveQuery(ctx, 3*time.Second)
		case <-sweepTicker.C:
			m.sweepStalePeers()
		}
	}
}

func (m *Manager) sweepStalePeers() {
	stale := m.store.SweepStale(120 * time.Second)
	if m.onPeerLost != nil {
		for _, peer := range stale {
			log.Printf("Background sweeper evicted stale peer: %s (%s)\n", peer.Hostname, peer.IP)
			m.onPeerLost(peer)
		}
	}
}
