<script setup>
import { useDevice } from '../composables/useDevice'
import { translateCityName } from '../utils/translate'

const { deviceName } = useDevice()
</script>

<template>
  <header class="app-header glass-panel">
    <div class="brand">
      <div class="brand-logo">
        <el-icon class="logo-icon"><Connection /></el-icon>
      </div>
      <div class="brand-info">
        <h1>وا<span class="text-gradient-cyan">صل</span></h1>
        <p>مشاركة آمنة للملفات عبر الشبكة المحلية</p>
      </div>
    </div>
    
    <div class="sys-status">
      <div class="status-indicator">
        <span class="pulse-dot"></span>
        <span class="status-text">اسم الجهاز: <strong>{{ translateCityName(deviceName) }}</strong></span>
      </div>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  flex-shrink: 0;
}

.brand {
  display: flex;
  align-items: center;
  gap: 16px;
}

.brand-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: var(--color-accent);
  box-shadow: 0 0 15px rgba(0, 242, 254, 0.2);
}

.logo-icon {
  font-size: 24px;
  color: #080c14;
}

.brand-info h1 {
  font-size: 20px;
  margin-bottom: 2px;
}

.brand-info p {
  font-size: 12px;
  color: var(--text-secondary);
}

.sys-status {
  display: flex;
  align-items: center;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.2);
  padding: 6px 14px;
  border-radius: 100px;
}

.pulse-dot {
  width: 8px;
  height: 8px;
  background-color: var(--color-success);
  border-radius: 50%;
  box-shadow: 0 0 8px var(--color-success);
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

.status-text {
  font-size: 12px;
  font-weight: 500;
  color: var(--color-success);
}
</style>
