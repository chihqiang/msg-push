<script setup lang="ts">
import { CheckCircle2, XCircle, Info } from '@lucide/vue'
import { useToastStore } from '@/stores/toast'

const toast = useToastStore()

const icons = {
  success: CheckCircle2,
  error: XCircle,
  info: Info,
} as const

const colors = {
  success: 'text-emerald-500',
  error: 'text-destructive',
  info: 'text-cyan-500',
} as const
</script>

<template>
  <Teleport to="body">
    <div class="fixed right-4 top-4 z-[60] flex w-80 flex-col gap-2">
      <TransitionGroup
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="translate-x-4 opacity-0"
        enter-to-class="translate-x-0 opacity-100"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="translate-x-0 opacity-100"
        leave-to-class="translate-x-4 opacity-0"
      >
        <div
          v-for="t in toast.toasts"
          :key="t.id"
          class="flex items-start gap-3 rounded-lg border border-border bg-card px-4 py-3 shadow-lg"
        >
          <component :is="icons[t.type]" class="mt-0.5 h-5 w-5 shrink-0" :class="colors[t.type]" />
          <p class="flex-1 text-sm text-foreground">{{ t.message }}</p>
          <button class="text-muted-foreground hover:text-foreground" @click="toast.remove(t.id)">✕</button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
