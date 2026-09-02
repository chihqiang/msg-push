<script setup lang="ts">
// 新手引导向导：选择消息类型 → 逐步展示配置链路 → 跳转对应页面
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowRight, Check, Sparkles, FlaskConical } from '@lucide/vue'
import Modal from '@/components/ui/Modal.vue'
import { guideConfigs, testGuide, type GuideConfig, type SendSample } from './guide-data'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const router = useRouter()

// 当前选中的引导配置（null 表示在类型选择页）
const selected = ref<GuideConfig | null>(null)
// 当前步骤索引
const stepIndex = ref(0)
// 是否已显示测试模式快捷入口
const showTestTip = ref(false)

const currentStep = computed(() => selected.value?.steps[stepIndex.value])
const isLast = computed(() => selected.value ? stepIndex.value === selected.value.steps.length - 1 : false)
// 是否测试模式引导（testGuide.type === 'test'）
const isTestGuide = computed(() => selected.value?.type === 'test')

// 打开时重置状态
watch(
  () => props.open,
  (open) => {
    if (open) {
      selected.value = null
      stepIndex.value = 0
      showTestTip.value = false
    }
  }
)

function choose(cfg: GuideConfig) {
  selected.value = cfg
  stepIndex.value = 0
}

function goStep(delta: number) {
  const next = stepIndex.value + delta
  if (next < 0 || !selected.value || next >= selected.value.steps.length) return
  stepIndex.value = next
}

/** 跳转到目标页面（可选 query 参数，页面据此自动打开新建弹窗等） */
function jumpTo(target?: string, query?: Record<string, string>) {
  if (!target) return
  emit('close')
  if (query && Object.keys(query).length > 0) {
    // 附加唯一 _t 参数：确保 query 对象每次都变化，触发目标页面的 watch(route.query)。
    // 否则在已带 ?action=create 的页面上重复跳转会因 query 未变而无法再次弹窗。
    router.push({ path: target, query: { ...query, _t: String(Date.now()) } })
  } else {
    router.push(target)
  }
}

/** 跳转到当前步骤的目标页面 */
function jump() {
  const step = currentStep.value
  jumpTo(step?.target, step?.query)
}

/** 点击操作提示（如「点击新建应用」）→ 跳转并自动打开对应弹窗 */
function jumpAction() {
  const step = currentStep.value
  jumpTo(step?.target, step?.query)
}

/** 格式化请求体：JSON 美化缩进 2 空格 */
function formatBody(body: Record<string, unknown>): string {
  try {
    return JSON.stringify(body, null, 2)
  } catch {
    return String(body)
  }
}

/** 将发送示例拼成终端样式多行文本 */
function sampleLines(sample: SendSample | undefined): string[] {
  if (!sample) return []
  const lines: string[] = [`${sample.method} ${sample.path}`]
  for (const h of sample.headers ?? []) lines.push(h)
  lines.push('')
  lines.push(formatBody(sample.body))
  return lines
}

function close() {
  emit('close')
}
</script>

<template>
  <Modal :open="open" :title="selected ? `${selected.label}配置引导` : '新手引导'" width="42rem" @close="close">
    <!-- 类型选择 -->
    <div v-if="!selected" class="space-y-3">
      <div class="flex items-start gap-3 rounded-lg border border-primary/20 bg-primary/5 px-4 py-3">
        <Sparkles class="mt-0.5 h-5 w-5 shrink-0 text-primary" />
        <div>
          <p class="text-sm font-medium text-foreground">选择你想接入的消息类型</p>
          <p class="mt-0.5 text-xs text-muted-foreground">每种类型都有完整的配置链路引导，按步骤操作即可完成接入。</p>
        </div>
      </div>

      <button
        v-for="cfg in guideConfigs"
        :key="cfg.type"
        class="flex w-full items-center gap-3 rounded-lg border border-border bg-card px-4 py-3 text-left transition-colors hover:border-primary/40 hover:bg-muted"
        @click="choose(cfg)"
      >
        <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg" :class="cfg.color">
          <component :is="cfg.icon" class="h-5 w-5" />
        </span>
        <span class="min-w-0 flex-1">
          <span class="block text-sm font-medium text-foreground">{{ cfg.label }}</span>
          <span class="mt-0.5 block truncate text-xs text-muted-foreground">{{ cfg.desc }}</span>
        </span>
        <span class="shrink-0 text-xs text-muted-foreground">{{ cfg.steps.length }} 步</span>
        <ArrowRight class="h-4 w-4 shrink-0 text-muted-foreground" />
      </button>

      <!-- 测试模式快捷入口 -->
      <div class="flex items-start gap-3 rounded-lg border border-dashed border-amber-300 bg-amber-50 px-4 py-3">
        <FlaskConical class="mt-0.5 h-5 w-5 shrink-0 text-amber-600" />
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-foreground">想先快速体验？试试测试模式</p>
          <p class="mt-0.5 text-xs text-muted-foreground">{{ testGuide.desc }}，无需真实服务商。</p>
          <button class="mt-2 text-xs font-medium text-amber-700 hover:underline" @click="choose(testGuide)">
            查看测试模式引导 →
          </button>
        </div>
      </div>
    </div>

    <!-- 步骤引导 -->
    <div v-else class="flex gap-5">
      <!-- 左侧步骤列表 -->
      <div class="w-44 shrink-0 space-y-1">
        <button
          v-for="(s, i) in selected.steps"
          :key="i"
          class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-xs transition-colors"
          :class="i === stepIndex ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground hover:bg-muted'"
          @click="stepIndex = i"
        >
          <span
            class="flex h-4.5 w-4.5 shrink-0 items-center justify-center rounded-full text-[10px]"
            :class="i < stepIndex ? 'bg-primary text-primary-foreground' : i === stepIndex ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'"
          >
            {{ i < stepIndex ? '' : i + 1 }}
          </span>
          <span class="truncate">{{ s.title }}</span>
        </button>
      </div>

      <!-- 右侧步骤内容 -->
      <div class="min-w-0 flex-1">
        <div class="flex items-start gap-3 rounded-lg border border-border bg-muted/40 px-4 py-3">
          <span class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-medium text-primary-foreground">
            {{ stepIndex + 1 }}
          </span>
          <div class="min-w-0">
            <p class="text-sm font-semibold text-foreground">{{ currentStep?.title }}</p>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">{{ currentStep?.desc }}</p>
            <button
              v-if="currentStep?.action"
              class="mt-1.5 inline-flex items-center gap-1 rounded-md bg-primary/10 px-2 py-1 text-xs font-medium text-primary transition-colors hover:bg-primary/20"
              @click="jumpAction"
            >
              {{ currentStep.action }}
              <ArrowRight class="h-3 w-3" />
            </button>
          </div>
        </div>

        <!-- 测试模式提示：测试引导直接显示，真实类型引导在用户点「测试模式」快捷入口后显示小提示 -->
        <div v-if="isTestGuide || showTestTip" class="mt-3">
          <div class="flex items-start gap-2 rounded-md border border-amber-200 bg-amber-50 px-3 py-2.5">
            <FlaskConical class="mt-0.5 h-4 w-4 shrink-0 text-amber-600" />
            <p class="text-xs leading-relaxed text-amber-800">{{ isTestGuide ? (selected?.testTip ?? '') : '小提示：' + (selected?.testTip ?? '') }}</p>
          </div>
        </div>

        <!-- 发送示例：终端样式（请求行 + 头 + 格式化 JSON body） -->
        <div v-if="isLast && selected.sendSample" class="mt-3">
          <p class="mb-1 text-xs font-medium text-muted-foreground">发送接口示例</p>
          <div class="overflow-x-auto rounded-md bg-slate-900 px-3 py-2.5 font-mono text-[11px] leading-relaxed text-slate-100">
            <div v-for="(line, i) in sampleLines(selected.sendSample)" :key="i" class="whitespace-pre">{{ line }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部操作 -->
    <template #footer>
      <div class="flex items-center justify-between gap-2">
        <div class="flex items-center gap-2">
          <button v-if="selected" class="rounded-md px-3 py-1.5 text-xs text-muted-foreground transition-colors hover:bg-muted" @click="selected = null">
            ← 返回类型选择
          </button>
        </div>

        <div class="flex items-center gap-2">
          <button class="h-8 rounded-md border border-border px-3 text-xs text-muted-foreground transition-colors hover:bg-muted" @click="close">
            关闭
          </button>
          <button
            v-if="selected && !isLast"
            class="inline-flex h-8 items-center gap-1 rounded-md border border-border px-3 text-xs font-medium text-foreground transition-colors hover:bg-muted"
            @click="goStep(1)"
          >
            下一步 <ArrowRight class="h-3.5 w-3.5" />
          </button>
          <button
            v-if="selected && isLast"
            class="inline-flex h-8 items-center gap-1 rounded-md bg-primary px-3 text-xs font-medium text-primary-foreground transition-opacity hover:opacity-90"
            @click="jump"
          >
            <Check class="h-3.5 w-3.5" /> 去 {{ currentStep?.target === '/tasks' ? '查看任务' : currentStep?.action ? '配置' : '完成' }}
          </button>
        </div>
      </div>
    </template>
  </Modal>
</template>
