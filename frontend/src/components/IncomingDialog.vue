<script setup>
import { useTransfers } from '../composables/useTransfers'
import { translateCityName } from '../utils/translate'

const { incomingRequest, acceptInbound, declineInbound } = useTransfers()
</script>

<template>
  <!-- Glassmorphic Permission Handshake Dialog -->
  <el-dialog
    v-model="incomingRequest"
    title="طلب ملف وارد"
    width="480px"
    align-center
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
    class="handshake-modal"
  >
    <div class="handshake-details" v-if="incomingRequest">
      <div class="sender-avatar">
        <el-icon class="pulse-icon"><User /></el-icon>
      </div>
      <h3>يريد <strong>{{ translateCityName(incomingRequest.peerName) }}</strong> مشاركة ملف معك</h3>
      <code class="sender-ip">{{ incomingRequest.peerIp }}</code>
      
      <div class="handshake-file-card glass-panel">
        <el-icon class="file-icon"><Document /></el-icon>
        <div class="file-meta">
          <h4>{{ incomingRequest.filename }}</h4>
          <span>{{ incomingRequest.formattedSize }}</span>
        </div>
      </div>
    </div>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="declineInbound(incomingRequest.id)" class="decline-btn">
          رفض
        </el-button>
        <el-button type="primary" @click="acceptInbound(incomingRequest.id)" class="accept-btn">
          قبول وحفظ
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
/* Handshake dialog overrides */
.handshake-details {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 12px;
  padding: 10px 0;
}

.sender-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: rgba(155, 81, 224, 0.1);
  border: 1px dashed rgba(155, 81, 224, 0.4);
  color: var(--color-purple);
  margin-bottom: 4px;
}

.pulse-icon {
  font-size: 26px;
  animation: pulse-ring 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse-ring {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: .3;
  }
}

.sender-ip {
  font-size: 13px;
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.03);
  padding: 4px 12px;
  border-radius: 100px;
  border: 1px solid var(--border-color);
}

.handshake-file-card {
  display: flex;
  align-items: center;
  padding: 16px 20px;
  width: 100%;
  gap: 16px;
  text-align: left;
  margin-top: 10px;
}

.file-icon {
  font-size: 26px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.file-meta {
  min-width: 0;
}

.file-meta h4 {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 2px;
}

.file-meta span {
  font-size: 11px;
  color: var(--text-secondary);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 10px;
}

.decline-btn {
  border-radius: 8px !important;
}

.accept-btn {
  background: var(--color-accent-purple) !important;
  border-radius: 8px !important;
}

.accept-btn:hover {
  box-shadow: 0 0 15px rgba(155, 81, 224, 0.4) !important;
}
</style>
