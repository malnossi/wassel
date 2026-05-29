package platform

import (
	"net"
	"testing"
	"time"
)

func TestIsLocalIP(t *testing.T) {
	// Loopback IP should not be marked as a local interface IP based on our custom logic
	// which filters out IsLoopback() in localIPCache.refresh().
	if IsLocalIP("127.0.0.1") {
		t.Error("Expected IsLocalIP(127.0.0.1) to be false (filtered loopback)")
	}

	// Retrieve real interface addresses to find a valid non-loopback local IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Skip("Skipping local IP verification due to interface lookup error:", err)
	}

	var targetIP string
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			targetIP = ipNet.IP.String()
			break
		}
	}

	if targetIP != "" {
		if !IsLocalIP(targetIP) {
			t.Errorf("Expected IsLocalIP(%s) to be true for active local interface", targetIP)
		}
	}
}

func TestLocalIPCacheTTL(t *testing.T) {
	localIPCache.mu.Lock()
	localIPCache.ips = map[string]bool{"192.168.1.50": true}
	localIPCache.updated = time.Now().Add(-10 * time.Second) // Force expired
	localIPCache.mu.Unlock()

	// This call should trigger a refresh because of the expired TTL
	IsLocalIP("127.0.0.1")

	localIPCache.mu.RLock()
	defer localIPCache.mu.RUnlock()
	if time.Since(localIPCache.updated) > 1*time.Second {
		t.Error("Expected local IP cache to have been refreshed after expired TTL")
	}
}

func TestGetOutboundIP(t *testing.T) {
	// Should be able to detect outbound IP for local interfaces or resolve fallback
	ip, err := GetOutboundIP("8.8.8.8")
	if err != nil {
		// If offline or no route, it fallback scans active interfaces
		if ip == "" {
			t.Error("Expected non-empty IP returned even under offline fallback mode")
		}
	} else {
		if net.ParseIP(ip) == nil {
			t.Errorf("Expected a valid parsed IP, got %q", ip)
		}
	}
}
