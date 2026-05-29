<script setup>
import PeerCard from './PeerCard.vue'
import { useDiscovery } from '../composables/useDiscovery'

defineProps({
  triggerSendFile: {
    type: Function,
    required: true
  }
})

const { peers, resetDiscovery } = useDiscovery()
</script>

<template>
  <section class="section-left">
    <div class="section-card glass-panel flex-col">
      <div class="card-header flex-header">
        <div>
          <h2>الأجهزة المكتشفة</h2>
          <p>اختر جهازاً محلياً لمشاركة الملفات فورياً</p>
        </div>
        <el-button 
          type="default" 
          circle
          class="refresh-btn"
          @click="resetDiscovery"
          title="إعادة تعيين البحث"
        >
          <el-icon><Refresh /></el-icon>
        </el-button>
      </div>

      <!-- Radar Hub (Rendered when no peers are active) -->
      <div v-if="peers.length === 0" class="radar-container">
        <div class="radar-hub">
          <div class="radar-wave radar-wave-1"></div>
          <div class="radar-wave radar-wave-2"></div>
          <div class="radar-wave radar-wave-3"></div>
          <el-icon class="radar-icon"><Search /></el-icon>
        </div>
        <p class="radar-label">جاري البحث في الشبكة المحلية...</p>
        <p class="radar-sublabel">تأكد من تشغيل تطبيق LocalShare على الأجهزة الأخرى</p>
      </div>

      <!-- Peers Grid -->
      <div v-else class="peers-list scrollable-area">
        <PeerCard 
          v-for="peer in peers" 
          :key="peer.ip" 
          :peer="peer"
          @send-file="triggerSendFile"
        />
      </div>
    </div>
  </section>
</template>

<style scoped>
.section-left {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.section-card {
  flex: 1;
  padding: 24px;
  min-height: 0;
}

.flex-col {
  display: flex;
  flex-direction: column;
}

.card-header {
  margin-bottom: 20px;
  flex-shrink: 0;
}

.card-header h2 {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 4px;
}

.card-header p {
  font-size: 13px;
  color: var(--text-secondary);
}

.flex-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.refresh-btn {
  background: rgba(255, 255, 255, 0.04) !important;
  border: 1px solid var(--border-color) !important;
  color: var(--text-secondary) !important;
  transition: var(--transition-smooth) !important;
}

.refresh-btn:hover {
  background: rgba(255, 255, 255, 0.1) !important;
  color: var(--color-cyan) !important;
  border-color: var(--border-color-hover) !important;
  box-shadow: 0 0 10px rgba(0, 242, 254, 0.15) !important;
}

/* Radar Animation container */
.radar-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  gap: 16px;
  position: relative;
}

.radar-hub {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 140px;
  height: 140px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px dashed rgba(0, 242, 254, 0.2);
  border-radius: 50%;
}

.radar-icon {
  font-size: 32px;
  color: var(--color-cyan);
  filter: drop-shadow(0 0 8px rgba(0, 242, 254, 0.4));
}

.radar-label {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  margin-top: 8px;
}

.radar-sublabel {
  font-size: 12px;
  color: var(--text-muted);
}

/* Scrollable Areas */
.scrollable-area {
  flex: 1;
  overflow-y: auto;
  padding-right: 4px;
}

/* Peers list */
.peers-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>
