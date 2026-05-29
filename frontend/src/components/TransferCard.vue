<script setup>
import { formatSpeed } from '../utils/format'
import { translateCityName } from '../utils/translate'

const props = defineProps({
  transfer: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['cancel'])

// Helper to determine status tags/colors
const getStatusType = (status) => {
  switch (status) {
    case 'completed': return 'success'
    case 'failed': return 'danger'
    case 'declined': return 'info'
    case 'transferring': return 'warning'
    case 'connecting': return 'primary'
    default: return 'info'
  }
}

// Helper to get status descriptions in Arabic
const getStatusLabel = (status) => {
  switch (status) {
    case 'completed': return 'اكتمل النقل'
    case 'failed': return 'فشل النقل'
    case 'declined': return 'تم الرفض'
    case 'transferring': return 'جاري النقل...'
    case 'connecting': return 'جاري الاتصال...'
    case 'cancelled': return 'تم الإلغاء'
    default: return status
  }
}

const isActive = (status) => status === 'transferring' || status === 'connecting'
</script>

<template>
  <div 
    class="transfer-card glass-panel"
    :class="transfer.isIncoming ? 'glow-card-purple' : 'glow-card-cyan'"
  >
    <div class="transfer-header">
      <div class="file-info">
        <el-icon class="file-icon"><Document /></el-icon>
        <div class="file-meta">
          <h4 :title="transfer.filename">{{ transfer.filename }}</h4>
          <span>{{ transfer.formattedSize }} &bull; {{ transfer.isIncoming ? 'وارد من' : 'صادر إلى' }} <strong>{{ translateCityName(transfer.peerName) }}</strong></span>
        </div>
      </div>
      <div class="transfer-actions">
        <el-button
          v-if="isActive(transfer.status)"
          type="danger"
          size="small"
          class="cancel-btn"
          @click="emit('cancel', transfer.id)"
        >
          إلغاء
        </el-button>
        <el-tag :type="getStatusType(transfer.status)" effect="dark" size="small" class="status-tag">
          {{ getStatusLabel(transfer.status) }}
        </el-tag>
      </div>
    </div>

    <!-- Status Message / Speed & Bytes -->
    <div class="transfer-body">
      <div v-if="transfer.status === 'transferring'" class="transfer-stats">
        <span class="speed-stat">{{ formatSpeed(transfer.speed) }}</span>
        <span class="progress-stat">{{ transfer.formattedProgress }}</span>
      </div>
      <div v-else-if="transfer.status === 'failed'" class="error-msg">
        <el-icon><CircleClose /></el-icon>
        <span>{{ transfer.error || 'فشل النقل' }}</span>
      </div>
      <div v-else-if="transfer.status === 'declined'" class="error-msg text-muted">
        <span>تم رفض النقل من قِبل الطرف الآخر</span>
      </div>

      <!-- Progress Bar -->
      <el-progress 
        v-if="transfer.status === 'transferring' || transfer.status === 'completed'" 
        :percentage="Math.round(transfer.percentage)" 
        :stroke-width="6" 
        :color="transfer.isIncoming ? '#9b51e0' : '#00f2fe'"
        :show-text="false"
        class="custom-progress"
      />
    </div>
  </div>
</template>

<style scoped>
.transfer-card {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.transfer-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
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
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-meta span {
  font-size: 11px;
  color: var(--text-secondary);
}

.transfer-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.cancel-btn {
  border-radius: 6px !important;
}

.status-tag {
  flex-shrink: 0;
}

.transfer-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.transfer-stats {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.speed-stat {
  font-weight: 600;
  color: var(--color-cyan);
}

.progress-stat {
  color: var(--text-secondary);
}

.error-msg {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--color-danger);
}

.custom-progress {
  margin-top: 2px;
}
</style>
