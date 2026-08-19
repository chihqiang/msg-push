<script setup lang="ts">
// 应用配额使用弹窗
import Modal from '@/components/ui/Modal.vue'
import type { AppQuotaUsage } from '@/types'

defineProps<{
  open: boolean
  data: AppQuotaUsage | null
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <Modal :open="open" title="今日配额使用" @close="emit('close')">
    <div v-if="data" class="space-y-4">
      <div class="flex items-center justify-between">
        <span class="text-sm text-muted-foreground">今日已用</span>
        <span class="text-xl font-semibold text-foreground">{{ data.today_used }}</span>
      </div>
      <div class="flex items-center justify-between">
        <span class="text-sm text-muted-foreground">每日配额</span>
        <span class="text-xl font-semibold text-foreground">{{ data.daily_quota === 0 ? '不限' : data.daily_quota }}</span>
      </div>
      <div class="flex items-center justify-between">
        <span class="text-sm text-muted-foreground">剩余</span>
        <span class="text-xl font-semibold text-foreground">{{ data.remaining }}</span>
      </div>
      <div>
        <div class="mb-1.5 flex items-center justify-between text-xs text-muted-foreground">
          <span>使用率</span>
          <span>{{ data.usage_percentage.toFixed(1) }}%</span>
        </div>
        <div class="h-2 overflow-hidden rounded-full bg-muted">
          <div
            class="h-full rounded-full transition-all"
            :class="data.usage_percentage > 80 ? 'bg-destructive' : 'bg-primary'"
            :style="{ width: `${Math.min(100, data.usage_percentage)}%` }"
          />
        </div>
      </div>
    </div>
  </Modal>
</template>
