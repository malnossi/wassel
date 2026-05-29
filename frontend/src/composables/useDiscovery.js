import { ref } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import * as App from '../../wailsjs/go/main/App'

// Module-level singleton state
const peers = ref([])

/**
 * Set up Wails event listeners for peer discovery/loss.
 */
const initDiscoveryListeners = () => {
  // Peer discovered
  EventsOn('peer:discovered', (peer) => {
    const idx = peers.value.findIndex(p => p.ip === peer.ip)
    if (idx === -1) {
      peers.value.push(peer)
    } else {
      peers.value[idx] = peer
    }
  })

  // Peer lost
  EventsOn('peer:lost', (peer) => {
    peers.value = peers.value.filter(p => p.ip !== peer.ip)
  })
}

/**
 * One-time fetch of current peers from backend (initial sync only).
 */
const refreshPeers = async () => {
  try {
    const list = await App.GetPeers()
    peers.value = list || []
  } catch (e) {
    console.error("Failed to fetch peers:", e)
  }
}

/**
 * Reset/clear the discovery list via the backend.
 */
const resetDiscovery = async () => {
  try {
    await App.ResetDiscovery()
  } catch (e) {
    console.error("Failed to reset discovery:", e)
  }
}

export function useDiscovery() {
  return {
    peers,
    initDiscoveryListeners,
    refreshPeers,
    resetDiscovery
  }
}
