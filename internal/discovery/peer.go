package discovery

import (
	"sync"
	"time"
)

// Peer represents a discovered device on the local network.
type Peer struct {
	Hostname string    `json:"hostname"`
	IP       string    `json:"ip"`
	Port     int       `json:"port"`
	LastSeen time.Time `json:"-"`
}

// PeerStore is a thread-safe container for discovered peers, keyed by IP address.
type PeerStore struct {
	mu    sync.Mutex
	peers map[string]Peer
}

// NewPeerStore creates an empty peer store.
func NewPeerStore() *PeerStore {
	return &PeerStore{
		peers: make(map[string]Peer),
	}
}

// UpsertResult indicates what happened when upserting a peer.
type UpsertResult struct {
	Peer        Peer
	IsNew       bool
	WasModified bool
}

// Upsert adds or updates a peer. Returns whether the frontend should be notified.
func (s *PeerStore) Upsert(peer Peer) UpsertResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.peers[peer.IP]
	shouldNotify := !exists || existing.Hostname != peer.Hostname || existing.Port != peer.Port

	peer.LastSeen = time.Now()
	s.peers[peer.IP] = peer

	return UpsertResult{
		Peer:        peer,
		IsNew:       !exists,
		WasModified: shouldNotify,
	}
}

// Remove deletes a peer by IP. Returns the removed peer and whether it existed.
func (s *PeerStore) Remove(ip string) (Peer, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peer, exists := s.peers[ip]
	if exists {
		delete(s.peers, ip)
	}
	return peer, exists
}

// GetAll returns a snapshot of all current peers.
func (s *PeerStore) GetAll() []Peer {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := make([]Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		list = append(list, peer)
	}
	return list
}

// GetNames returns a set of all current peer hostnames (for name collision avoidance).
func (s *PeerStore) GetNames() map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	names := make(map[string]bool, len(s.peers))
	for _, peer := range s.peers {
		names[peer.Hostname] = true
	}
	return names
}

// Clear removes all peers and returns the evicted list.
func (s *PeerStore) Clear() []Peer {
	s.mu.Lock()
	defer s.mu.Unlock()

	evicted := make([]Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		evicted = append(evicted, peer)
	}
	s.peers = make(map[string]Peer)
	return evicted
}

// SweepStale removes peers not seen within the given TTL. Returns evicted peers.
func (s *PeerStore) SweepStale(ttl time.Duration) []Peer {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var stale []Peer
	for ip, peer := range s.peers {
		if now.Sub(peer.LastSeen) > ttl {
			stale = append(stale, peer)
			delete(s.peers, ip)
		}
	}
	return stale
}
