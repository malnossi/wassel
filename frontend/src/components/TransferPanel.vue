<script setup>
import { computed } from 'vue'
import TransferCard from './TransferCard.vue'
import { useTransfers } from '../composables/useTransfers'

const { transfers, cancelTransfer, clearCompleted } = useTransfers()

const hasCompleted = computed(() =>
  transfers.value.some(t => t.status === 'completed' || t.status === 'failed' || t.status === 'declined')
)
</script>

<template>
  <section class="section-right">
    <div class="section-card glass-panel flex-col">
      <div class="card-header flex-header">
        <div>
          <h2>قائمة عمليات النقل</h2>
          <p>عمليات النقل النشطة وسجل الملفات</p>
        </div>
        <el-button
          v-if="hasCompleted"
          type="default"
          size="small"
          class="clear-btn"
          @click="clearCompleted"
        >
          مسح المكتملة
        </el-button>
      </div>

      <!-- Empty state queue -->
      <div v-if="transfers.length === 0" class="empty-queue-container">
        <el-icon class="empty-icon"><FolderOpened /></el-icon>
        <p class="empty-label">لا توجد عمليات نقل نشطة</p>
        <p class="empty-sublabel">ستظهر الملفات المرسلة والتنزيلات الواردة هنا</p>
      </div>

      <!-- Transfer Ledger List -->
      <div v-else class="transfer-list scrollable-area">
        <TransferCard 
          v-for="t in transfers" 
          :key="t.id" 
          :transfer="t"
          @cancel="cancelTransfer"
        />
      </div>
    </div>
  </section>
</template>

<style scoped>
.section-right {
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

.clear-btn {
  border-radius: 8px !important;
}

/* Empty Queue */
.empty-queue-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  gap: 16px;
  color: var(--text-muted);
}

.empty-icon {
  font-size: 40px;
}

.empty-label {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-secondary);
}

.empty-sublabel {
  font-size: 12px;
  text-align: center;
  max-width: 280px;
}

/* Scrollable Areas */
.scrollable-area {
  flex: 1;
  overflow-y: auto;
  padding-right: 4px;
}

/* Transfers Queue */
.transfer-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>
