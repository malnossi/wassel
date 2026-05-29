import { ref } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import * as App from '../../wailsjs/go/main/App'
import { formatBytes } from '../utils/format'

// Module-level singleton state
const transfers = ref([])
const incomingRequest = ref(null)

/**
 * Set up Wails event listeners for transfer lifecycle.
 */
const initTransferListeners = () => {
  // Incoming transfer request handshake
  EventsOn('transfer:request', (payload) => {
    incomingRequest.value = {
      id: payload.id,
      filename: payload.filename,
      size: payload.size,
      formattedSize: formatBytes(payload.size),
      peerIp: payload.peerIp,
      peerName: payload.peerName
    }
  })

  // Progress updates
  EventsOn('transfer:progress', (data) => {
    const idx = transfers.value.findIndex(t => t.id === data.id)
    if (idx !== -1) {
      const t = transfers.value[idx]
      // Calculate instantaneous speed
      const now = Date.now()
      const elapsedSecs = (now - t.lastUpdateTime) / 1000

      if (elapsedSecs > 0.05) { // Throttle speed calculations
        const bytesSentSinceLast = data.current - t.lastBytesTransferred
        t.speed = bytesSentSinceLast / elapsedSecs
        t.lastBytesTransferred = data.current
        t.lastUpdateTime = now
      }

      t.percentage = data.percentage
      t.current = data.current
      t.formattedProgress = `${formatBytes(data.current)} of ${formatBytes(data.total)}`

      transfers.value[idx] = { ...t }
    }
  })

  // Transfer status transitions
  EventsOn('transfer:status', (data) => {
    const idx = transfers.value.findIndex(t => t.id === data.id)
    const now = Date.now()

    if (idx === -1) {
      // Create new item in queue
      transfers.value.unshift({
        id: data.id,
        filename: data.filename,
        size: data.size,
        formattedSize: formatBytes(data.size),
        isIncoming: data.isIncoming,
        status: data.status,
        percentage: 0,
        current: 0,
        speed: 0,
        peerName: data.peerName || 'Unknown Device',
        error: data.error || '',
        lastBytesTransferred: 0,
        lastUpdateTime: now
      })
    } else {
      // Update existing item status
      const t = transfers.value[idx]
      t.status = data.status
      t.error = data.error || ''
      if (data.status === 'completed') {
        t.percentage = 100
        t.speed = 0
      } else if (data.status === 'failed' || data.status === 'declined') {
        t.speed = 0
      }
      transfers.value[idx] = { ...t }
    }
  })
}

/**
 * Send file to a peer.
 */
const triggerSendFile = async (peerIP) => {
  try {
    await App.SelectAndSendFile(peerIP)
  } catch (e) {
    console.error("Failed to initiate file send:", e)
  }
}

/**
 * Accept incoming transfer.
 */
const acceptInbound = async (id) => {
  if (!incomingRequest.value) return
  try {
    incomingRequest.value = null
    await App.AcceptTransfer(id)
  } catch (e) {
    console.error("Failed to accept transfer:", e)
  }
}

/**
 * Decline incoming transfer.
 */
const declineInbound = async (id) => {
  if (!incomingRequest.value) return
  try {
    incomingRequest.value = null
    await App.DeclineTransfer(id)
  } catch (e) {
    console.error("Failed to decline transfer:", e)
  }
}

/**
 * Cancel an active transfer (NEW).
 */
const cancelTransfer = async (id) => {
  try {
    await App.CancelTransfer(id)
  } catch (e) {
    console.error("Failed to cancel transfer:", e)
  }
}

/**
 * Remove completed/failed/declined transfers from the list (NEW).
 */
const clearCompleted = () => {
  transfers.value = transfers.value.filter(
    t => t.status !== 'completed' && t.status !== 'failed' && t.status !== 'declined'
  )
}

export function useTransfers() {
  return {
    transfers,
    incomingRequest,
    initTransferListeners,
    triggerSendFile,
    acceptInbound,
    declineInbound,
    cancelTransfer,
    clearCompleted
  }
}
