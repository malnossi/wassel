package discovery

import (
	"context"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	"localshare/internal/platform"

	"github.com/grandcat/zeroconf"
)

// MDNSService handles mDNS service registration and browsing with deduplicated entry processing.
type MDNSService struct {
	server         *zeroconf.Server
	resolver       *zeroconf.Resolver
	store          *PeerStore
	registeredName string
	regNameMu      sync.RWMutex
	onPeerFound    func(Peer)
}

// NewMDNSService creates a new mDNS service bound to the given peer store.
func NewMDNSService(store *PeerStore, onPeerFound func(Peer)) *MDNSService {
	return &MDNSService{
		store:       store,
		onPeerFound: onPeerFound,
	}
}

// Start initializes the mDNS resolver and begins passive browsing.
func (s *MDNSService) Start(ctx context.Context) error {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return err
	}
	s.resolver = resolver

	// Start passive browse loop
	go s.browseLoop(ctx)

	return nil
}

// RegisterService advertises our local instance via mDNS.
func (s *MDNSService) RegisterService(instanceName string, listenPort int) error {
	s.regNameMu.Lock()
	s.registeredName = instanceName
	s.regNameMu.Unlock()

	server, err := zeroconf.Register(
		instanceName,
		"_localshare._tcp",
		"local.",
		listenPort,
		[]string{"txtv=1"},
		nil,
	)
	if err != nil {
		return err
	}
	s.server = server
	return nil
}

// GetRegisteredName returns the currently registered mDNS instance name.
func (s *MDNSService) GetRegisteredName() string {
	s.regNameMu.RLock()
	defer s.regNameMu.RUnlock()
	return s.registeredName
}

// TriggerActiveQuery runs a short mDNS browse to actively query for peers.
func (s *MDNSService) TriggerActiveQuery(ctx context.Context, duration time.Duration) {
	browseCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry)
	go s.consumeEntries(browseCtx, entries)

	err := s.resolver.Browse(browseCtx, "_localshare._tcp", "local.", entries)
	if err != nil {
		log.Println("Error in active mDNS query:", err)
	}

	<-browseCtx.Done()
}

// Shutdown stops the mDNS server.
func (s *MDNSService) Shutdown() {
	if s.server != nil {
		s.server.Shutdown()
	}
}

// browseLoop runs the passive mDNS browse for the lifetime of the context.
func (s *MDNSService) browseLoop(ctx context.Context) {
	entries := make(chan *zeroconf.ServiceEntry)
	go s.consumeEntries(ctx, entries)

	err := s.resolver.Browse(ctx, "_localshare._tcp", "local.", entries)
	if err != nil {
		log.Println("Error starting passive mDNS Browse:", err)
	}
}

// consumeEntries processes mDNS entries from a channel — the SINGLE deduplicated
// entry processing method replacing the 3x copy-pasted logic from the old service.go.
func (s *MDNSService) consumeEntries(ctx context.Context, entries <-chan *zeroconf.ServiceEntry) {
	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-entries:
			if !ok {
				return
			}
			s.processEntry(entry)
		}
	}
}

// processEntry is the SINGLE method for mDNS entry processing.
// Previously this logic was copy-pasted 3 times in browseLoop, Reset, and triggerActiveQuery.
func (s *MDNSService) processEntry(entry *zeroconf.ServiceEntry) {
	// 1. Extract first IPv4 address
	var ip string
	if len(entry.AddrIPv4) > 0 {
		ip = entry.AddrIPv4[0].String()
	} else {
		return
	}

	// 2. Unescape DNS-SD instance name
	instanceName := unescapeInstanceName(entry.Instance)

	// 3. Filter self-discovery by IP (using cached lookup)
	if platform.IsLocalIP(ip) {
		return
	}

	// 4. Filter self-discovery by registered name
	s.regNameMu.RLock()
	regName := s.registeredName
	s.regNameMu.RUnlock()

	if regName != "" && instanceName == regName {
		return
	}

	// 5. Upsert into peer store
	peer := Peer{
		Hostname: instanceName,
		IP:       ip,
		Port:     entry.Port,
	}
	result := s.store.Upsert(peer)

	// 6. Notify if new or modified
	if result.WasModified && s.onPeerFound != nil {
		s.onPeerFound(result.Peer)
	}
}

// --- DNS-SD Name Unescaping ---

var decimalEscapeRegex = regexp.MustCompile(`\\([0-9]{3})`)
var spaceEscapeRegex = regexp.MustCompile(`\\(.)`)

// unescapeInstanceName converts DNS-SD decimal escape sequences (like \196\129) and
// space escapes back to their actual UTF-8 character representation.
func unescapeInstanceName(s string) string {
	// 1. Replace decimal escapes like \196 with their byte value
	res := decimalEscapeRegex.ReplaceAllStringFunc(s, func(match string) string {
		numStr := match[1:] // extract decimal digits, e.g. "196"
		num, err := strconv.Atoi(numStr)
		if err == nil && num >= 0 && num <= 255 {
			return string([]byte{byte(num)})
		}
		return match
	})

	// 2. Replace space and backslash escapes like \.
	res = spaceEscapeRegex.ReplaceAllString(res, "$1")

	return res
}
