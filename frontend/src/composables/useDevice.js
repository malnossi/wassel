import { ref } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import * as App from '../../wailsjs/go/main/App'

// Module-level singleton state
const deviceName = ref('Registering...')

/**
 * Set up Wails event listener for device name updates.
 */
const initDeviceListeners = () => {
  EventsOn('device:name', (name) => {
    deviceName.value = name
  })
}

/**
 * One-time fetch of the device name from the backend.
 */
const refreshDeviceName = async () => {
  try {
    const name = await App.GetDeviceName()
    if (name) {
      deviceName.value = name
    }
  } catch (e) {
    console.error("Failed to fetch device name:", e)
  }
}

export function useDevice() {
  return {
    deviceName,
    initDeviceListeners,
    refreshDeviceName
  }
}
