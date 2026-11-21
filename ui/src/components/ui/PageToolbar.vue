<script setup lang="ts">
// 页面工具栏：标题 + 右侧操作区
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <div class="mb-6 flex flex-wrap items-center justify-between gap-3">
    <div>
      <h2 class="text-xl font-semibold text-foreground">{{ title }}</h2>
      <p v-if="description" class="mt-0.5 text-sm text-muted-foreground">{{ description }}</p>
    </div>
    <div class="flex items-center gap-2.5">
      <slot />
    </div>
  </div>
</template>
