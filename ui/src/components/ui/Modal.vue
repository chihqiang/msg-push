<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import { X } from '@lucide/vue'

const props = defineProps<{
  open: boolean
  title: string
  width?: string
}>()

const emit = defineEmits<{
  close: []
}>()

// 支持 Esc 关闭：弹窗打开时监听全局 keydown，其他弹窗关闭后移除
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}

watch(
  () => props.open,
  (open) => {
    if (open) window.addEventListener('keydown', onKeydown)
    else window.removeEventListener('keydown', onKeydown)
  }
)

onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        @click.self="emit('close')"
      >
        <div class="absolute inset-0 bg-slate-950/60" />
        <div
          class="relative flex max-h-[90vh] w-full flex-col overflow-hidden rounded-xl border border-border bg-card shadow-2xl"
          :style="{ maxWidth: width || '32rem' }"
        >
          <div class="flex items-center justify-between border-b border-border px-5 py-4">
            <h3 class="text-base font-semibold text-foreground">{{ title }}</h3>
            <button
              class="rounded-md p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
              @click="emit('close')"
            >
              <X class="h-4.5 w-4.5" />
            </button>
          </div>
          <div class="flex-1 overflow-y-auto px-5 py-4">
            <slot />
          </div>
          <div v-if="$slots.footer" class="border-t border-border px-5 py-3.5">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.15s ease;
}
.modal-enter-active .relative,
.modal-leave-active .relative {
  transition:
    transform 0.15s ease,
    opacity 0.15s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-from .relative,
.modal-leave-to .relative {
  opacity: 0;
  transform: translateY(8px) scale(0.98);
}
</style>
