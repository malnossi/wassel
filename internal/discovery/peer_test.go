package discovery

import (
	"testing"
	"time"
)

func TestPeerStoreUpsertAndGet(t *testing.T) {
	store := NewPeerStore()

	peer := Peer{
		Hostname: "TestDevice",
		IP:       "192.168.1.100",
		Port:     9998,
	}

	// 1. Initial Upsert
	res1 := store.Upsert(peer)
	if !res1.IsNew {
		t.Error("Expected IsNew to be true on initial insert")
	}
	if !res1.WasModified {
		t.Error("Expected WasModified to be true on initial insert")
	}

	// 2. Duplicate Upsert with identical data (should not trigger modification callback notification)
	res2 := store.Upsert(peer)
	if res2.IsNew {
		t.Error("Expected IsNew to be false on duplicate insert")
	}
	if res2.WasModified {
		t.Error("Expected WasModified to be false when data did not change")
	}

	// 3. Upsert with updated Hostname (should trigger modification callback)
	peer.Hostname = "TestDeviceUpdated"
	res3 := store.Upsert(peer)
	if res3.IsNew {
		t.Error("Expected IsNew to be false on update")
	}
	if !res3.WasModified {
		t.Error("Expected WasModified to be true when Hostname changed")
	}

	// 4. GetAll
	all := store.GetAll()
	if len(all) != 1 {
		t.Errorf("Expected 1 peer in store, got %d", len(all))
	}
	if all[0].Hostname != "TestDeviceUpdated" {
		t.Errorf("Expected Hostname 'TestDeviceUpdated', got %q", all[0].Hostname)
	}

	// 5. GetNames
	names := store.GetNames()
	if !names["TestDeviceUpdated"] {
		t.Error("Expected 'TestDeviceUpdated' to be in returned name map")
	}
}

func TestPeerStoreRemoveAndClear(t *testing.T) {
	store := NewPeerStore()

	peer1 := Peer{Hostname: "Device1", IP: "192.168.1.1"}
	peer2 := Peer{Hostname: "Device2", IP: "192.168.1.2"}

	store.Upsert(peer1)
	store.Upsert(peer2)

	// Remove peer1
	removed, ok := store.Remove("192.168.1.1")
	if !ok {
		t.Error("Expected Remove to return true for existing peer")
	}
	if removed.Hostname != "Device1" {
		t.Errorf("Expected removed peer Hostname to be 'Device1', got %q", removed.Hostname)
	}

	// Double remove
	_, ok2 := store.Remove("192.168.1.1")
	if ok2 {
		t.Error("Expected second Remove to return false")
	}

	// Clear remaining
	evicted := store.Clear()
	if len(evicted) != 1 || evicted[0].Hostname != "Device2" {
		t.Error("Expected Clear to evict remaining Device2 peer")
	}

	if len(store.GetAll()) != 0 {
		t.Error("Expected store to be empty after Clear")
	}
}

func TestPeerStoreSweepStale(t *testing.T) {
	store := NewPeerStore()

	peer1 := Peer{Hostname: "StaleDevice", IP: "192.168.1.1"}
	peer2 := Peer{Hostname: "FreshDevice", IP: "192.168.1.2"}

	store.Upsert(peer1)
	store.Upsert(peer2)

	// Artificially age peer1's LastSeen date
	store.mu.Lock()
	p1 := store.peers["192.168.1.1"]
	p1.LastSeen = time.Now().Add(-10 * time.Minute)
	store.peers["192.168.1.1"] = p1
	store.mu.Unlock()

	// Sweep peers older than 5 minutes (should only evict StaleDevice)
	evicted := store.SweepStale(5 * time.Minute)
	if len(evicted) != 1 || evicted[0].Hostname != "StaleDevice" {
		t.Errorf("Expected to sweep 'StaleDevice', swept %+v", evicted)
	}

	all := store.GetAll()
	if len(all) != 1 || all[0].Hostname != "FreshDevice" {
		t.Errorf("Expected 'FreshDevice' to remain in store, remaining: %+v", all)
	}
}
