package platform

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// cachedLocalIPs holds a thread-safe cache of the machine's local IP addresses.
// Avoids calling net.InterfaceAddrs() on every mDNS entry in hot loops.
type cachedLocalIPs struct {
	mu      sync.RWMutex
	ips     map[string]bool
	updated time.Time
	ttl     time.Duration
}

var localIPCache = &cachedLocalIPs{
	ips: make(map[string]bool),
	ttl: 5 * time.Second,
}

func (c *cachedLocalIPs) refresh() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	ips := make(map[string]bool, len(addrs))
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			ips[ipNet.IP.String()] = true
		}
	}
	c.mu.Lock()
	c.ips = ips
	c.updated = time.Now()
	c.mu.Unlock()
}

// IsLocalIP checks whether the given IP belongs to this machine.
// Uses a cache with a 5-second TTL to avoid repeated syscalls in hot loops.
func IsLocalIP(ip string) bool {
	localIPCache.mu.RLock()
	needsRefresh := time.Since(localIPCache.updated) > localIPCache.ttl
	localIPCache.mu.RUnlock()

	if needsRefresh {
		localIPCache.refresh()
	}

	localIPCache.mu.RLock()
	defer localIPCache.mu.RUnlock()
	return localIPCache.ips[ip]
}

// GetOutboundIP determines the local IP address that would be used to reach the given target IP.
// Falls back to scanning active network interfaces if UDP dial fails.
func GetOutboundIP(targetIP string) (string, error) {
	conn, err := net.Dial("udp", targetIP+":80")
	if err != nil {
		// Fallback to searching active interfaces
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}
		return "", fmt.Errorf("no active network interfaces")
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
