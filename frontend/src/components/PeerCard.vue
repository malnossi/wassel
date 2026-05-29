<script setup>
import { translateCityName } from '../utils/translate'

defineProps({
  peer: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['send-file'])
</script>

<template>
  <div class="peer-card glass-panel glow-card-cyan">
    <div class="peer-avatar">
      <el-icon><Monitor /></el-icon>
    </div>
    <div class="peer-details">
      <h3>{{ translateCityName(peer.hostname) }}</h3>
      <code>{{ peer.ip }}</code>
    </div>
    <el-button 
      type="primary" 
      class="send-btn" 
      @click="emit('send-file', peer.ip)"
    >
      <template #icon>
        <el-icon><Upload /></el-icon>
      </template>
      إرسال ملف
    </el-button>
  </div>
</template>

<style scoped>
.peer-card {
  display: flex;
  align-items: center;
  padding: 16px 20px;
  gap: 16px;
}

.peer-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border-radius: 12px;
  background: rgba(0, 242, 254, 0.1);
  border: 1px solid rgba(0, 242, 254, 0.2);
  color: var(--color-cyan);
  font-size: 20px;
}

.peer-details {
  flex: 1;
}

.peer-details h3 {
  font-size: 15px;
  margin-bottom: 2px;
}

.peer-details code {
  font-size: 12px;
  color: var(--text-secondary);
}

.send-btn {
  border-radius: 8px !important;
}
</style>
