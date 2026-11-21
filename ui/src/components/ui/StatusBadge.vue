<script setup lang="ts">
// 状态徽章：启用/禁用
withDefaults(
  defineProps<{
    value: boolean
    activeText?: string
    inactiveText?: string
  }>(),
  {
    activeText: '已启用',
    inactiveText: '已禁用',
  }
)
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium"
    :class="value ? 'bg-emerald-500/10 text-emerald-600' : 'bg-destructive/10 text-destructive'"
  >
    <span class="h-1.5 w-1.5 rounded-full" :class="value ? 'bg-emerald-500' : 'bg-destructive'" />
    {{ value ? activeText : inactiveText }}
  </span>
</template>
