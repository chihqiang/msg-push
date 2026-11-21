<script setup lang="ts">
import { computed, useSlots } from 'vue'
import { Loader2 } from '@lucide/vue'

// 列配置：通过插槽 `#cell-{key}` 自定义单元格，未提供插槽时按 row[key] 渲染
export interface TableColumn {
  key: string
  label: string
  width?: string
  align?: 'left' | 'center' | 'right'
  class?: string
}

const slots = useSlots()

const props = withDefaults(
  defineProps<{
    columns: TableColumn[]
    data: unknown[]
    loading?: boolean
    total?: number
    page?: number
    pageSize?: number
    rowKey?: string
    emptyText?: string
  }>(),
  {
    loading: false,
    total: 0,
    page: 1,
    pageSize: 10,
    rowKey: 'id',
    emptyText: '暂无数据',
  }
)

const emit = defineEmits<{
  'update:page': [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

function cellValue(row: unknown, key: string): unknown {
  return (row as Record<string, unknown>)[key]
}

function alignClass(align?: string) {
  if (align === 'right') return 'text-right'
  if (align === 'center') return 'text-center'
  return 'text-left'
}
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-border bg-card">
    <div class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-border bg-muted/50 text-left text-xs font-medium text-muted-foreground">
            <th
              v-for="col in columns"
              :key="col.key"
              class="px-4 py-3"
              :class="[alignClass(col.align), col.class]"
              :style="col.width ? `width: ${col.width}` : undefined"
            >
              {{ col.label }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="columns.length" class="px-4 py-12 text-center text-muted-foreground">
              <Loader2 class="mx-auto h-5 w-5 animate-spin" />
            </td>
          </tr>
          <tr v-else-if="data.length === 0">
            <td :colspan="columns.length" class="px-4 py-12 text-center text-muted-foreground">
              {{ emptyText }}
            </td>
          </tr>
          <tr
            v-for="row in data"
            :key="(row as Record<string, unknown>)[rowKey] as string | number"
            class="border-b border-border/60 transition-colors last:border-0 hover:bg-muted/30"
          >
            <td
              v-for="col in columns"
              :key="col.key"
              class="px-4 py-3"
              :class="[alignClass(col.align), col.class]"
            >
              <slot
                v-if="slots[`cell-${col.key}`]"
                :name="`cell-${col.key}`"
                :row="row"
                :value="cellValue(row, col.key)"
              />
              <template v-else>{{ cellValue(row, col.key) ?? '—' }}</template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div class="flex items-center justify-between border-t border-border px-4 py-3">
      <p class="text-xs text-muted-foreground">
        共 {{ total }} 条 · 第 {{ page }} / {{ totalPages }} 页
      </p>
      <div class="flex items-center gap-2">
        <button
          class="h-8 rounded-md border border-border px-3 text-sm text-muted-foreground transition-colors hover:bg-muted disabled:opacity-40"
          :disabled="page <= 1 || loading"
          @click="emit('update:page', page - 1)"
        >
          上一页
        </button>
        <button
          class="h-8 rounded-md border border-border px-3 text-sm text-muted-foreground transition-colors hover:bg-muted disabled:opacity-40"
          :disabled="page >= totalPages || loading"
          @click="emit('update:page', page + 1)"
        >
          下一页
        </button>
      </div>
    </div>
  </div>
</template>
