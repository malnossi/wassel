<script setup>
import { onMounted } from 'vue'
import AppHeader from './components/AppHeader.vue'
import PeerPanel from './components/PeerPanel.vue'
import TransferPanel from './components/TransferPanel.vue'
import IncomingDialog from './components/IncomingDialog.vue'
import { useDiscovery } from './composables/useDiscovery'
import { useTransfers } from './composables/useTransfers'
import { useDevice } from './composables/useDevice'

const { initDiscoveryListeners, refreshPeers } = useDiscovery()
const { initTransferListeners, triggerSendFile } = useTransfers()
const { initDeviceListeners, refreshDeviceName } = useDevice()

onMounted(() => {
  // Initialize all event listeners
  initDiscoveryListeners()
  initTransferListeners()
  initDeviceListeners()

  // One-time sync from backend (no polling)
  refreshPeers()
  refreshDeviceName()
})
</script>

<template>
  <div class="dashboard-container" dir="rtl">
    <AppHeader />

    <main class="dashboard-body">
      <PeerPanel :triggerSendFile="triggerSendFile" />
      <TransferPanel />
    </main>

    <IncomingDialog />
  </div>
</template>

<style scoped>
.dashboard-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  padding: 20px;
  gap: 20px;
}

.dashboard-body {
  display: flex;
  flex: 1;
  gap: 20px;
  min-height: 0; /* Important for flex child scroll */
}
</style>
