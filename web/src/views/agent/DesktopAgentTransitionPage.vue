<template>
  <main class="desktop-agent-transition" aria-labelledby="desktop-agent-title">
    <div class="desktop-agent-transition__ambient desktop-agent-transition__ambient--cyan" aria-hidden="true"></div>
    <div class="desktop-agent-transition__ambient desktop-agent-transition__ambient--orange" aria-hidden="true"></div>

    <section class="desktop-agent-shell">
      <header class="desktop-agent-header">
        <div class="desktop-agent-header__left">
          <button type="button" class="desktop-agent-header__back" aria-label="返回首页" @click="goHome">
            <ArrowLeft aria-hidden="true" />
          </button>

          <span class="desktop-agent-header__mark" aria-hidden="true">
            <Cpu />
            <i></i>
          </span>

          <div class="desktop-agent-header__copy">
            <h1 id="desktop-agent-title">丝路 Agent 正在生成出海营销方案</h1>
            <p>正在理解商品信息、目标市场、合规边界与营销目标。</p>
          </div>
        </div>

        <span class="desktop-agent-status" :class="statusClass">
          <CircleCheck v-if="phase === 'completed'" aria-hidden="true" />
          <span v-else class="desktop-agent-spinner" aria-hidden="true"></span>
          <span>{{ statusLabel }}</span>
        </span>
      </header>

      <section class="desktop-user-prompt" aria-label="用户输入摘要">
        <span class="desktop-user-prompt__icon" aria-hidden="true">
          <MagicStick />
        </span>
        <div class="desktop-user-prompt__copy">
          <span>用户输入</span>
          <p>
            <template v-for="(part, index) in userPromptParts" :key="`${part.text}-${index}`">
              <strong v-if="part.tone" :class="`is-${part.tone}`">{{ part.text }}</strong>
              <span v-else>{{ part.text }}</span>
            </template>
          </p>
        </div>
      </section>

      <section class="desktop-agent-main-grid">
        <article class="desktop-summary-card" aria-label="分析摘要">
          <span class="desktop-summary-card__top-line" aria-hidden="true"></span>
          <div class="desktop-panel-heading">
            <div class="desktop-panel-heading__title">
              <span class="desktop-panel-heading__icon" aria-hidden="true">
                <MagicStick />
              </span>
            </div>
            <span class="desktop-panel-pill" :class="{ 'is-completed': analysisDone }">
              <CircleCheck v-if="analysisDone" aria-hidden="true" />
              <span v-else class="desktop-agent-spinner" aria-hidden="true"></span>
              {{ analysisDone ? '分析完成' : '分析中' }}
            </span>
          </div>

          <p v-if="fallbackNotice" class="desktop-fallback-notice">{{ fallbackNotice }}</p>

          <div class="desktop-summary-body" aria-live="polite">
            <template v-if="summaryParagraphSegments.length">
              <p
                v-for="(paragraph, paragraphIndex) in summaryParagraphSegments"
                :key="paragraph.key"
                :class="{ 'is-latest': paragraphIndex === summaryParagraphSegments.length - 1 && !analysisDone }"
              >
                <template v-for="(part, partIndex) in paragraph.parts" :key="`${paragraph.key}-${part.text}-${partIndex}`">
                  <strong v-if="part.tone" :class="`is-${part.tone}`">{{ part.text }}</strong>
                  <span v-else>{{ part.text }}</span>
                </template>
                <span v-if="paragraphIndex === summaryParagraphSegments.length - 1 && !analysisDone" class="desktop-type-cursor" aria-hidden="true"></span>
              </p>
            </template>
            <div v-else class="desktop-summary-placeholder">
              <span class="desktop-dot-loader" aria-hidden="true"><i></i><i></i><i></i></span>
            </div>
          </div>

          <div v-if="analysisDone && phase !== 'completed'" class="desktop-analysis-done" aria-live="polite">
            <p>分析摘要已完成，正在编排 Agent 任务链……</p>
            <span>
              <span class="desktop-agent-spinner" aria-hidden="true"></span>
              任务链生成中
            </span>
          </div>
        </article>

        <aside class="desktop-recognized-panel" aria-labelledby="desktop-recognized-title">
          <div class="desktop-recognized-panel__head">
            <div>
              <h2 id="desktop-recognized-title">
                <CollectionTag aria-hidden="true" />
                <span>实时识别信息</span>
              </h2>
            </div>
            <span class="desktop-recognized-panel__badge" :class="{ 'is-empty': visibleRecognizedItems.length === 0 }">
              <CircleCheck v-if="visibleRecognizedItems.length" aria-hidden="true" />
              <span v-else class="desktop-agent-spinner" aria-hidden="true"></span>
              {{ visibleRecognizedItems.length ? '已识别' : '抽取中' }}
            </span>
          </div>

          <TransitionGroup name="desktop-recognized-list" tag="div" class="desktop-recognized-list">
            <article
              v-for="item in visibleRecognizedItems"
              :key="item.label"
              class="desktop-recognized-item"
              :class="`is-${item.tone}`"
            >
              <span class="desktop-recognized-item__icon" aria-hidden="true">
                <component :is="item.icon" />
              </span>
              <span class="desktop-recognized-item__copy">
                <small>{{ item.label }}</small>
                <strong>{{ item.value }}</strong>
              </span>
              <CircleCheck class="desktop-recognized-item__check" aria-hidden="true" />
            </article>
          </TransitionGroup>

          <div v-if="!visibleRecognizedItems.length" class="desktop-recognized-empty">
            <span class="desktop-dot-loader" aria-hidden="true"><i></i><i></i><i></i></span>
            <span>等待分析摘要产生可抽取信息</span>
          </div>
        </aside>
      </section>

      <Transition name="desktop-chain-in">
        <section v-if="showTaskChain" class="desktop-task-chain-card" aria-labelledby="desktop-task-title">
          <div class="desktop-task-chain-card__header">
            <div class="desktop-task-chain-card__title">
              <span aria-hidden="true">
                <Lightning />
              </span>
              <h2 id="desktop-task-title">Agent 自动任务链</h2>
            </div>
            <p>
              进度
              <strong>{{ completedTaskCount }}/{{ taskSteps.length }}</strong>
            </p>
          </div>

          <div class="desktop-task-chain" aria-label="Agent 自动任务链完成进度">
            <span class="desktop-task-chain__track" aria-hidden="true"></span>
            <span class="desktop-task-chain__progress" :style="{ width: taskProgressWidth }" aria-hidden="true"></span>

            <ol class="desktop-task-chain__list">
              <li
                v-for="task in taskSteps"
                :key="task.step"
                class="desktop-task-chain__item"
                :class="taskClass(task.step)"
              >
                <span class="desktop-task-chain__node" aria-hidden="true">
                  <CircleCheck v-if="task.step <= completedTaskCount" />
                  <span v-else>{{ task.step }}</span>
                </span>
                <strong>{{ task.name }}</strong>
              </li>
            </ol>
          </div>
        </section>
      </Transition>

      <Transition name="desktop-complete-in">
        <section v-if="phase === 'completed'" class="desktop-complete-notice" aria-live="polite">
          <span class="desktop-complete-notice__icon" aria-hidden="true">
            <CircleCheck />
          </span>
          <span class="desktop-complete-notice__copy">
            <strong>所有任务节点已完成</strong>
            <small>正在进入生成结果页……</small>
          </span>
          <span class="desktop-complete-notice__status">
            <span class="desktop-dot-loader is-green" aria-hidden="true"><i></i><i></i><i></i></span>
            自动跳转中
          </span>
        </section>
      </Transition>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { Component } from 'vue'
import type { AgentInput, AgentResult } from '@/types/agent'
import { agentAPI } from '@/api/agent'
import { readAgentInput, readAgentUserInput, saveAgentResult } from '@/utils/agentStorage'
import {
  ArrowLeft,
  CircleCheck,
  CollectionTag,
  Cpu,
  Goods,
  Lightning,
  Location,
  MagicStick,
  Platform,
  PriceTag,
  User
} from '@element-plus/icons-vue'

type Phase = 'analyzing' | 'arranging' | 'completed'
type Tone = 'cyan' | 'violet' | 'orange' | 'blue'

interface RecognizedInfo {
  product: string
  category: string
  market: string
  platform: string
  audience: string
  sellingPoints: string
}

interface RecognizedItem {
  label: string
  value: string
  tone: Tone
  icon: Component
}

interface TaskStep {
  step: number
  name: string
  description: string
}

interface TextPart {
  text: string
  tone?: Tone
}

const DEFAULT_USER_INPUT = '请补充商品、目标市场、内容平台、目标用户和核心卖点。'
const BLOCKED_SUMMARY_WORDS = ['思维链', 'chain-of-thought', 'reasoning_content', '内部推理', '模型推理过程', '内部思考']

const router = useRouter()
const phase = ref<Phase>('analyzing')
const summaryText = ref('')
const recognizedInfo = ref<RecognizedInfo | null>(null)
const visibleInfoCount = ref(0)
const analysisDone = ref(false)
const showTaskChain = ref(false)
const completedTaskCount = ref(0)
const fallbackNotice = ref('')
const generationReady = ref(false)
const allDone = ref(false)
const hasRedirected = ref(false)
const storedAgentInput = ref<AgentInput | null>(readAgentInput())

let abortController: AbortController | null = null
let redirectTimer: number | undefined
let resultFallbackTimer: number | undefined
const timers: number[] = []

const taskSteps: TaskStep[] = [
  { step: 1, name: '商品理解', description: '确认商品类目、卖点与使用场景' },
  { step: 2, name: '合规风险识别', description: '匹配目标市场规则与广告敏感表达' },
  { step: 3, name: '本地化方向', description: '根据已明确的目标市场生成内容方向' },
  { step: 4, name: '短视频脚本', description: '生成开头、中段、结尾三段式脚本' },
  { step: 5, name: '数字人方案', description: '推荐数字人形象、口播语气与字幕语言' },
  { step: 6, name: '投放优化', description: '规划平台、内容方向与关键指标' }
]

const userInputText = computed(() => {
  return storedAgentInput.value?.rawPrompt?.trim() || readAgentUserInput().trim() || DEFAULT_USER_INPUT
})

const currentInfo = computed(() => recognizedInfo.value || inferRecognizedInfo(storedAgentInput.value, userInputText.value))

const recognizedItems = computed<RecognizedItem[]>(() => [
  { label: '商品', value: currentInfo.value.product, tone: 'cyan', icon: Goods },
  { label: '商品类目', value: currentInfo.value.category, tone: 'cyan', icon: CollectionTag },
  { label: '目标市场', value: currentInfo.value.market, tone: 'blue', icon: Location },
  { label: '目标平台', value: currentInfo.value.platform, tone: 'violet', icon: Platform },
  { label: '目标人群', value: currentInfo.value.audience, tone: 'violet', icon: User },
  { label: '核心卖点', value: currentInfo.value.sellingPoints, tone: 'orange', icon: PriceTag }
])

const visibleRecognizedItems = computed(() => {
  return recognizedItems.value.slice(0, Math.min(visibleInfoCount.value, recognizedItems.value.length))
})

const statusLabel = computed(() => {
  if (phase.value === 'completed') return '分析完成'
  if (analysisDone.value || phase.value === 'arranging') return '任务编排中'
  return '分析中'
})

const statusClass = computed(() => ({
  'is-running': phase.value === 'analyzing',
  'is-arranging': analysisDone.value && phase.value !== 'completed',
  'is-completed': phase.value === 'completed'
}))

const summaryParagraphSegments = computed(() => {
  return summaryText.value
    .split(/\n+/)
    .map((item) => item.trim())
    .filter(Boolean)
    .map((text, index) => ({
      key: `${index}-${text.slice(0, 12)}`,
      parts: segmentText(text)
    }))
})

const userPromptParts = computed(() => {
  return segmentText(userInputText.value)
})

const taskProgressWidth = computed(() => {
  if (completedTaskCount.value <= 0) return '0%'
  return `${Math.min(100, ((completedTaskCount.value - 1) / (taskSteps.length - 1)) * 100)}%`
})

const schedule = (callback: () => void, delay: number) => {
  const timer = window.setTimeout(callback, delay)
  timers.push(timer)
  return timer
}

const wait = (delay: number) => new Promise<void>((resolve) => {
  schedule(resolve, delay)
})

const clampStep = (value: number) => {
  return Math.max(0, Math.min(taskSteps.length, value))
}

const taskClass = (step: number) => {
  const currentStep = clampStep(completedTaskCount.value + 1)
  return {
    'is-completed': step <= completedTaskCount.value,
    'is-current': phase.value === 'arranging' && step === currentStep,
    'is-pending': step > currentStep
  }
}

const goHome = () => {
  router.push('/')
}

const isUsableAgentImage = (value: string) => {
  return value.trim().toLowerCase().startsWith('data:image/')
}

const buildAnalyzePayload = () => {
  const input = storedAgentInput.value
  return {
    scene: 'desktop_transition',
    userInput: userInputText.value,
    productName: input?.productName || '',
    category: input?.category || '',
    targetMarket: input?.targetMarket || '',
    targetPlatform: input?.targetPlatform || '',
    targetAudience: input?.targetAudience || '',
    coreSellingPoints: input?.coreSellingPoints || [],
    materialSpec: input?.materialSpec || '',
    usageScenario: input?.usageScenario || '',
    imageDataUrl: isUsableAgentImage(input?.imageDataUrl || '') ? input?.imageDataUrl || '' : ''
  }
}

const startRecognizedReveal = () => {
  recognizedInfo.value = currentInfo.value
  const delays = [760, 1400, 2100, 2850, 3600, 4350]
  delays.forEach((delay, index) => {
    schedule(() => {
      if (allDone.value) return
      visibleInfoCount.value = Math.max(visibleInfoCount.value, index + 1)
    }, delay)
  })
}

const startResultGeneration = async () => {
  const input = storedAgentInput.value || buildFallbackInput()
  resultFallbackTimer = schedule(() => {
    if (generationReady.value) return
    saveAgentResult(buildFallbackResult(input, currentRecognizedInfo()))
    generationReady.value = true
    tryRedirect()
  }, 18000)

  try {
    const result = await agentAPI.generate(input)
    saveAgentResult(result)
    generationReady.value = true
    tryRedirect()
  } catch {
    if (!generationReady.value) {
      saveAgentResult(buildFallbackResult(input, currentRecognizedInfo()))
      generationReady.value = true
      tryRedirect()
    }
  }
}

const startStreamingAnalysis = async () => {
  abortController = new AbortController()
  try {
    const response = await fetch('/api/v1/agent/analyze', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(buildAnalyzePayload()),
      signal: abortController.signal
    })
    if (!response.ok || !response.body) {
      throw new Error('stream unavailable')
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    while (true) {
      const { value, done } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      const blocks = buffer.split(/\n\n/)
      buffer = blocks.pop() || ''
      blocks.forEach(handleSSEBlock)
    }
    if (buffer.trim()) {
      handleSSEBlock(buffer)
    }
    if (!allDone.value) {
      await runTaskFallback()
    }
  } catch {
    await runLocalFallbackFlow()
  }
}

const handleSSEBlock = (block: string) => {
  const lines = block.split(/\r?\n/)
  let event = 'message'
  const dataLines: string[] = []
  lines.forEach((line) => {
    if (line.startsWith('event:')) {
      event = line.slice(6).trim()
    } else if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).trim())
    }
  })
  if (!dataLines.length) return

  let data: any = {}
  try {
    data = JSON.parse(dataLines.join('\n'))
  } catch {
    data = {}
  }
  handleStreamEvent(event, data)
}

const handleStreamEvent = (event: string, data: any) => {
  switch (event) {
    case 'analysis_summary_delta':
      appendSummary(data.text || '')
      break
    case 'recognized_info':
      recognizedInfo.value = normalizeRecognizedInfo(data)
      visibleInfoCount.value = Math.max(visibleInfoCount.value, 1)
      break
    case 'analysis_done':
      markAnalysisDone()
      break
    case 'task_status':
      handleTaskStatus(data)
      break
    case 'all_done':
      completeAllTasks()
      break
    case 'fallback_notice':
    case 'error':
      fallbackNotice.value = data.message || '网络波动，已切换为本地演示流程'
      break
    default:
      break
  }
}

const appendSummary = (text: string) => {
  const cleanText = sanitizeSummaryText(text)
  if (!cleanText) return
  summaryText.value += cleanText
  const paragraphCount = summaryText.value.split(/\n+/).filter((item) => item.trim()).length
  visibleInfoCount.value = Math.max(visibleInfoCount.value, Math.min(paragraphCount + 1, recognizedItems.value.length))
}

const sanitizeSummaryText = (value: string) => {
  return BLOCKED_SUMMARY_WORDS.reduce((result, word) => result.split(word).join(''), value)
}

const markAnalysisDone = () => {
  if (analysisDone.value) return
  analysisDone.value = true
  phase.value = 'arranging'
  visibleInfoCount.value = recognizedItems.value.length
  schedule(() => {
    showTaskChain.value = true
  }, 300)
}

const handleTaskStatus = (data: any) => {
  const step = clampStep(Number(data.step) || 0)
  if (!step) return
  if (!analysisDone.value) {
    markAnalysisDone()
  }
  showTaskChain.value = true
  phase.value = 'arranging'
  const nextCount = data.status === 'running' ? step - 1 : step
  completedTaskCount.value = Math.max(completedTaskCount.value, clampStep(nextCount))
}

const runLocalFallbackFlow = async () => {
  if (allDone.value) return
  fallbackNotice.value = '网络波动，已切换为本地演示流程'
  const info = currentRecognizedInfo()
  const paragraphs = [
    '已接收到你的出海需求，正在从自然语言中提取商品、目标市场、平台、人群和核心卖点。\n\n',
    `识别到商品为${info.product}，属于${info.category}，${describeTargetMarketForSummary(info.market)}，主要投放平台为${info.platform}。\n\n`,
    `${buildComplianceFocusText(info)}\n\n`,
    '接下来将基于合规边界，生成本地化营销方向、短视频脚本、数字人方案和投放建议。'
  ]

  for (const paragraph of paragraphs) {
    await typeSummary(paragraph)
  }
  recognizedInfo.value = info
  markAnalysisDone()
  await runTaskFallback()
}

const typeSummary = async (text: string) => {
  const chars = Array.from(text)
  for (const ch of chars) {
    appendSummary(ch)
    await wait(ch === '\n' ? 40 : 12)
  }
}

const runTaskFallback = async () => {
  if (!analysisDone.value) {
    markAnalysisDone()
  }
  showTaskChain.value = true
  phase.value = 'arranging'
  for (const task of taskSteps) {
    if (completedTaskCount.value >= task.step) continue
    await wait(500)
    completedTaskCount.value = task.step
  }
  completeAllTasks()
}

const completeAllTasks = () => {
  completedTaskCount.value = taskSteps.length
  showTaskChain.value = true
  phase.value = 'completed'
  visibleInfoCount.value = recognizedItems.value.length
  allDone.value = true
  tryRedirect()
}

const tryRedirect = () => {
  if (!allDone.value || hasRedirected.value) return
  if (!generationReady.value) {
    schedule(() => {
      if (!generationReady.value) {
        const input = storedAgentInput.value || buildFallbackInput()
        saveAgentResult(buildFallbackResult(input, currentRecognizedInfo()))
        generationReady.value = true
      }
      tryRedirect()
    }, 1500)
    return
  }
  hasRedirected.value = true
  redirectTimer = schedule(() => {
    router.push('/agent/result')
  }, 1500)
}

const currentRecognizedInfo = (): RecognizedInfo => {
  return recognizedInfo.value || inferRecognizedInfo(storedAgentInput.value, userInputText.value)
}

const normalizeRecognizedInfo = (value: Partial<RecognizedInfo>): RecognizedInfo => {
  const fallback = currentRecognizedInfo()
  return {
    product: value.product || fallback.product,
    category: value.category || fallback.category,
    market: value.market || fallback.market,
    platform: value.platform || fallback.platform,
    audience: value.audience || fallback.audience,
    sellingPoints: value.sellingPoints || fallback.sellingPoints
  }
}

const isUnknownTargetMarket = (value: string) => {
  const cleanValue = cleanupExtractedValue(value)
  return !cleanValue || /待补充|待识别|未明确/.test(cleanValue)
}

const describeTargetMarketForSummary = (market: string) => {
  return isUnknownTargetMarket(market) ? '目标市场暂未明确' : `目标市场为${cleanupExtractedValue(market)}`
}

const targetMarketForStorage = (market: string) => {
  return isUnknownTargetMarket(market) ? '' : cleanupExtractedValue(market)
}

const inferRecognizedInfo = (input: AgentInput | null, prompt: string): RecognizedInfo => {
  const product = input?.productName || extractPromptProductName(prompt) || '待分析商品'
  const category = input?.category || inferCategoryByProduct(product)
  const market = input?.targetMarket || inferTargetMarket(prompt) || '目标市场待补充'
  const platform = input?.targetPlatform || extractPlatform(prompt) || 'TikTok'
  const audience = input?.targetAudience || extractMatch(prompt, /(?:目标用户|目标人群|受众|面向用户|用户是|人群是)(?:是|为)?([^，,。；;\n]{2,44})/) || '年轻女性 / 学生 / 办公室'
  const sellingPoints = input?.coreSellingPoints?.length
    ? input.coreSellingPoints.join(' / ')
    : extractMatch(prompt, /(?:主打|卖点|核心卖点|突出)(?:是|为)?([^。；;\n]{2,80})/) || inferSellingPointsByProduct(product)

  return {
    product: cleanupExtractedValue(product),
    category: cleanupExtractedValue(category),
    market: cleanupExtractedValue(market),
    platform: platform.trim(),
    audience: cleanupExtractedValue(audience),
    sellingPoints: normalizeSellingPoints(cleanupExtractedValue(sellingPoints))
  }
}

const extractMatch = (value: string, pattern: RegExp) => value.match(pattern)?.[1]?.trim() || ''

const extractPromptProductName = (value: string) => {
  const taggedProduct = extractMatch(value, /(?:商品准入分析|商品|产品|品名|商品名称|产品名称)\s*[:：]\s*([^，,。；;\n]{1,28})/)
  if (taggedProduct) {
    return cleanupProductName(taggedProduct)
  }

  const compactProduct = extractMatch(value, /^(?:帮我|请|麻烦)?(?:分析一下|分析|看看|测一下|识别一下|识别)?\s*([^，,。；;\n]{1,28}?)(?:卖到|卖|出口到|进入|面向|投放到|推广到|上架|发布到|做|去|给|给到|$)/)
  if (compactProduct) {
    return cleanupProductName(compactProduct)
  }

  const naturalProduct = extractMatch(value, /(?:我有|我们有|这是一款|这款|一款|一个|一种|商品是|产品是)([^，,。；;\n]{2,28}?)(?:，|,|。|；|;|想|计划|准备|主打|目标|卖到|出口|做|$)/)
  if (naturalProduct) {
    return cleanupProductName(naturalProduct)
  }

  const compactLine = value
    .split(/\n+/)
    .map((line) => cleanupProductName(line))
    .find((line) => line && line.length <= 18)

  return compactLine || ''
}

const extractPlatform = (value: string) => {
  const platforms = ['TikTok', 'Instagram Reels', 'Instagram', 'YouTube Shorts', 'YouTube', 'Shopee', 'Lazada', 'Amazon', 'Temu', 'eBay', 'Facebook', '小红书', '抖音']
  const lowerValue = value.toLowerCase()
  return platforms.find((platform) => lowerValue.includes(platform.toLowerCase())) || ''
}

const inferTargetMarket = (value: string) => {
  const matched = cleanupExtractedValue(extractMatch(value, /(?:卖到|卖|出口到|进入|面向|投放到|推广到|去|给到)([^，,。；;\n]{1,24}?)(?:市场|用户|消费者|，|,|。|；|;|$)/))
    .replace(/TikTok|Instagram Reels|Instagram|YouTube Shorts|YouTube|Shopee|Lazada|Amazon|Temu|eBay|Facebook|小红书|抖音/gi, '')
    .trim()
  if (matched) return matched

  const markets = ['美国', '英国', '加拿大', '澳大利亚', '德国', '法国', '日本', '韩国', '马来西亚', '新加坡', '泰国', '越南', '印尼', '印度尼西亚', '菲律宾', '印度', '墨西哥', '巴西', '沙特', '阿联酋', '中东', '欧洲', '东南亚', '北美']
  const lowerValue = value.toLowerCase()
  return markets.find((market) => lowerValue.includes(market.toLowerCase())) || ''
}

const cleanupExtractedValue = (value: string) => {
  return value.trim().replace(/^[，,。；;\s]+|[，,。；;\s]+$/g, '')
}

const cleanupProductName = (value: string) => {
  return cleanupExtractedValue(value)
    .split(/(?:生成本地化脚本|数字人口播方案|投放优化建议|商品准入分析)\s*[:：]?/)[0]
    .replace(/^(分析|识别|检测)/, '')
    .trim()
}

const isFoodProduct = (product: string, category = '') => {
  return /食品|饮料|零食|即食|餐|鸡|鸭|肉|鱼|虾|糕|饼|糖|茶|咖啡|炸|烤|卤|吃/.test(`${product} ${category}`)
}

const inferCategoryByProduct = (product: string) => {
  if (product.includes('榨汁杯') || product.includes('杯')) return '小家电 / 食品接触用品'
  if (isFoodProduct(product)) return '食品饮料 / 即食食品'
  return '跨境电商商品'
}

const inferSellingPointsByProduct = (product: string) => {
  if (isFoodProduct(product)) return '口味 / 便捷 / 场景化'
  if (product.includes('榨汁杯') || product.includes('杯')) return '便携 / 健康 / 易清洗'
  return '实用 / 本地化 / 易展示'
}

const buildComplianceFocusText = (info: RecognizedInfo) => {
  if (isFoodProduct(info.product, info.category)) {
    return '该商品涉及食品销售场景，后续合规分析将重点关注配料表、过敏原提示、保质期、产地标识、食品准入与平台广告表达边界。'
  }
  if (info.category.includes('食品接触') || info.product.includes('杯')) {
    return '该商品涉及食品接触场景，后续合规分析将重点关注材质说明、食品接触认证、电池参数和充电安全描述。'
  }
  return '该商品后续合规分析将重点关注目标市场准入要求、必要认证、禁限售规则和广告敏感表达。'
}

const normalizeSellingPoints = (value: string) => {
  const points = value
    .split(/[\/,，;；、和与]/)
    .map((item) => cleanupExtractedValue(item))
    .filter(Boolean)
  return points.length ? points.slice(0, 4).join(' / ') : '便携 / 健康 / 易清洗'
}

const segmentText = (text: string): TextPart[] => {
  const info = currentInfo.value
  const keywords = [
    { text: info.product, tone: 'cyan' as Tone },
    { text: info.category, tone: 'cyan' as Tone },
    { text: info.market, tone: 'blue' as Tone },
    { text: info.platform, tone: 'violet' as Tone },
    { text: 'TikTok 短视频', tone: 'violet' as Tone },
    { text: '年轻女生', tone: 'orange' as Tone },
    { text: '年轻女性', tone: 'orange' as Tone },
    { text: '食品接触场景', tone: 'orange' as Tone },
    { text: '合规边界', tone: 'blue' as Tone }
  ].filter((item) => item.text)

  const parts: TextPart[] = []
  let rest = text
  while (rest) {
    const next = keywords
      .map((keyword) => ({ ...keyword, index: rest.indexOf(keyword.text) }))
      .filter((keyword) => keyword.index >= 0)
      .sort((a, b) => a.index - b.index || b.text.length - a.text.length)[0]

    if (!next) {
      parts.push({ text: rest })
      break
    }
    if (next.index > 0) {
      parts.push({ text: rest.slice(0, next.index) })
    }
    parts.push({ text: next.text, tone: next.tone })
    rest = rest.slice(next.index + next.text.length)
  }
  return parts
}

const buildFallbackInput = (): AgentInput => {
  const info = currentRecognizedInfo()
  return {
    requestId: createRequestId(),
    rawPrompt: userInputText.value,
    productName: info.product,
    category: info.category,
    targetMarket: targetMarketForStorage(info.market),
    targetPlatform: info.platform,
    targetAudience: info.audience,
    coreSellingPoints: info.sellingPoints.split(/[\/,，;；]/).map((item) => item.trim()).filter(Boolean)
  }
}

const buildFallbackResult = (input: AgentInput, info: RecognizedInfo): AgentResult => {
  const sellingPoints = info.sellingPoints.split(/[\/,，;；]/).map((item) => item.trim()).filter(Boolean)
  const isFood = isFoodProduct(info.product, info.category)
  return {
    schemaVersion: 2,
    requestId: input.requestId || createRequestId(),
    recognizedInfo: {
      productName: info.product,
      category: info.category,
      targetMarket: targetMarketForStorage(info.market) || '目标市场待补充',
      targetPlatform: info.platform,
      targetAudience: info.audience,
      coreSellingPoints: sellingPoints.length ? sellingPoints : inferSellingPointsByProduct(info.product).split(' / '),
      imageUnderstanding: '电脑端过渡页已完成基础商品识别。'
    },
    overview: {
      complianceRiskLevel: '中风险',
      marketStrategy: isFood ? '先补齐食品标签与准入信息，再生成本地化素材。' : '先补齐必要认证与商品信息，再生成本地化素材。',
      recommendedVideoStyle: '生活场景化竖屏短视频',
      recommendedDigitalHuman: '亲和型本地化生活方式讲解者'
    },
    compliance: {
      title: '合规分析结果',
      summary: isFood
        ? '当前已完成基础识别，建议补充配料表、过敏原、保质期、产地与食品准入材料，以便进一步判断准入与广告表达边界。'
        : '当前已完成基础识别，建议补充商品材质、规格、认证或检测材料，以便进一步判断准入与广告表达边界。',
      riskTags: isFood ? ['食品标签', '准入材料', '功效表达'] : ['信息缺口', '认证材料', '功效表达'],
      missingInfo: isFood ? ['配料表', '过敏原提示', '保质期', '产地/生产信息'] : ['材质/成分/规格', '认证或检测报告', '目标售价区间'],
      suggestions: ['补充商品图片或详细描述，以便进行合规风险评估。'],
      forbiddenExpressions: ['100%安全', '永久有效', '官方认证', '治疗/治愈', '保证通过'],
      saferExpressions: ['适合日常使用', '建议查看材质说明', '以实际认证材料为准']
    },
    localization: {
      direction: '场景化种草 + 实用价值说明',
      reason: isFood ? '目标市场用户更容易被真实试吃、风味描述、价格场景和清晰食品标签信息吸引。' : '目标市场用户更容易被真实生活场景、明确卖点和清晰产品演示吸引。',
      keywords: isFood ? ['crispy chicken', 'ready to eat', 'taste test', 'snack time'] : ['portable', 'daily use', 'easy to use', 'lifestyle'],
      tone: '自然、可信、少夸张',
      sceneSuggestions: isFood ? ['开箱试吃', '朋友聚餐', '夜宵场景'] : ['开箱展示', '办公室/宿舍', '通勤随身携带']
    },
    script: {
      title: '短视频脚本',
      duration: '20-25s',
      opening: { time: '0-3s', content: '用早餐来不及准备的真实生活小问题切入，展示目标用户的高频痛点。' },
      middle: { time: '3-20s', content: isFood ? '展示包装、色泽、口感反馈和食用场景，强调风味与便利性，避免绝对化或健康功效承诺。' : '展示商品外观、核心卖点和使用步骤，强调日常场景价值，避免绝对化承诺。' },
      ending: { time: '20-25s', content: '用温和行动号召引导查看商品详情与材质说明。' },
      storyboard: []
    },
    digitalHuman: {
      persona: '亲和型本地生活方式讲解者',
      tone: '自然、可靠、少夸张',
      videoRatio: '9:16',
      subtitleAdvice: '目标市场语言字幕 + 关键卖点英文短词',
      visualStyle: '明亮、干净、生活方式感',
      shootingStyle: '手持近景 + 产品特写 + 场景演示'
    },
    promotion: {
      platforms: [info.platform],
      contentTags: isFood ? ['试吃测评', '风味表达', '夜宵场景', '开箱展示'] : ['生活场景', '实用卖点', '开箱演示'],
      focusMetrics: ['完播率', '点击率', '收藏率'],
      optimizationAdvice: '先用场景化短视频测试用户兴趣，再根据评论补充材质、认证和使用限制信息。'
    },
    agentMessage: {
      summary: '已完成电脑端过渡页分析与任务编排，建议继续补充商品材料和认证信息。',
      missingInfoNotice: isFood ? '建议补充配料表、过敏原提示、保质期和目标市场食品准入材料。' : '建议补充材质、规格、认证或检测材料。',
      quickActions: ['补充商品图片', '换成印尼市场', '进入视频剪辑']
    },
    isMock: true,
    model: 'local-fallback'
  }
}

const createRequestId = () => {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

onMounted(() => {
  startRecognizedReveal()
  startResultGeneration()
  startStreamingAnalysis()
})

onBeforeUnmount(() => {
  abortController?.abort()
  timers.forEach((timer) => window.clearTimeout(timer))
  if (redirectTimer) window.clearTimeout(redirectTimer)
  if (resultFallbackTimer) window.clearTimeout(resultFallbackTimer)
})
</script>

<style scoped>
.desktop-agent-transition {
  position: relative;
  min-height: 100vh;
  overflow-x: hidden;
  padding: 40px 30px;
  background:
    linear-gradient(144deg, #f8fafc 0%, #ffffff 50%, rgba(124, 58, 237, 0.05) 100%),
    linear-gradient(120deg, rgba(6, 182, 212, 0.08), rgba(249, 115, 22, 0.04));
  color: #0a2463;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', sans-serif;
}

.desktop-agent-transition__ambient {
  position: fixed;
  z-index: 0;
  pointer-events: none;
  border-radius: 999px;
  filter: blur(64px);
}

.desktop-agent-transition__ambient--cyan {
  left: calc(50% - 526px);
  top: 80px;
  width: 384px;
  height: 384px;
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.15), rgba(124, 58, 237, 0.15));
}

.desktop-agent-transition__ambient--orange {
  left: calc(50% + 10px);
  top: 236px;
  width: 500px;
  height: 500px;
  background: linear-gradient(135deg, rgba(249, 115, 22, 0.1), rgba(6, 182, 212, 0.1));
}

.desktop-agent-shell {
  position: relative;
  z-index: 1;
  width: min(1052px, calc(100vw - 60px));
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.desktop-agent-header {
  min-height: 54px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
}

.desktop-agent-header__left {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.desktop-agent-header__back {
  width: 40px;
  height: 40px;
  padding: 0;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #ffffff;
  color: #45556c;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.1), 0 1px 2px rgba(15, 23, 42, 0.1);
  cursor: pointer;
}

.desktop-agent-header__back svg {
  width: 16px;
  height: 16px;
}

.desktop-agent-header__mark {
  position: relative;
  width: 48px;
  height: 48px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  box-shadow: 0 10px 15px rgba(124, 58, 237, 0.3), 0 4px 6px rgba(124, 58, 237, 0.3);
  color: #ffffff;
}

.desktop-agent-header__mark svg {
  width: 24px;
  height: 24px;
}

.desktop-agent-header__mark i {
  position: absolute;
  right: -4px;
  top: -4px;
  width: 14px;
  height: 14px;
  border: 2px solid #ffffff;
  border-radius: 999px;
  background: #05df72;
  opacity: 0.82;
}

.desktop-agent-header__copy h1 {
  margin: 0;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 24px;
  line-height: 30px;
  font-weight: 700;
  letter-spacing: 0;
}

.desktop-agent-header__copy p {
  margin: 4px 0 0;
  color: #62748e;
  font-size: 14px;
  line-height: 20px;
}

.desktop-agent-status,
.desktop-panel-pill,
.desktop-recognized-panel__badge {
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  white-space: nowrap;
  font-weight: 700;
}

.desktop-agent-status {
  min-height: 30.5px;
  padding: 6px 13px;
  gap: 6px;
  border: 1px solid rgba(186, 230, 253, 0.9);
  background: #ecfeff;
  color: #0891b2;
  font-size: 11px;
  line-height: 16.5px;
}

.desktop-agent-status.is-arranging {
  border-color: rgba(196, 181, 253, 0.78);
  background: #f5f3ff;
  color: #7c3aed;
}

.desktop-agent-status.is-completed {
  border-color: #b9f8cf;
  background: #f0fdf4;
  color: #00a63e;
}

.desktop-agent-status svg,
.desktop-panel-pill svg,
.desktop-recognized-panel__badge svg {
  width: 12px;
  height: 12px;
}

.desktop-agent-spinner {
  width: 11px;
  height: 11px;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 999px;
  display: inline-block;
  animation: desktopSpin 780ms linear infinite;
}

.desktop-user-prompt,
.desktop-summary-card,
.desktop-recognized-panel,
.desktop-task-chain-card,
.desktop-complete-notice {
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.1), 0 1px 2px rgba(15, 23, 42, 0.1);
}

.desktop-user-prompt {
  min-height: 87px;
  padding: 21px;
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.desktop-user-prompt__icon {
  width: 32px;
  height: 32px;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: linear-gradient(135deg, #f1f5f9 0%, #f8fafc 100%);
  color: #7c3aed;
}

.desktop-user-prompt__icon svg {
  width: 16px;
  height: 16px;
}

.desktop-user-prompt__copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.desktop-user-prompt__copy > span {
  color: #90a1b9;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.desktop-user-prompt__copy p {
  margin: 0;
  color: #314158;
  font-size: 16px;
  line-height: 26px;
}

.desktop-user-prompt__copy strong,
.desktop-summary-body strong {
  font-weight: 700;
}

.is-cyan {
  color: #0891b2;
}

.is-blue {
  color: #0a2463;
}

.is-violet {
  color: #7c3aed;
}

.is-orange {
  color: #f97316;
}

.desktop-agent-main-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 337px;
  gap: 20px;
}

.desktop-summary-card {
  position: relative;
  min-height: 423.5px;
  overflow: hidden;
  padding: 24px;
  border-color: #f1f5f9;
  box-shadow: 0 4px 6px -1px rgba(226, 232, 240, 0.6), 0 2px 4px -2px rgba(226, 232, 240, 0.6);
}

.desktop-summary-card__top-line {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(124, 58, 237, 0.4), transparent);
}

.desktop-panel-heading {
  min-height: 40.5px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.desktop-panel-heading__title {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.desktop-panel-heading__icon {
  width: 36px;
  height: 36px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  box-shadow: 0 1px 3px rgba(124, 58, 237, 0.2), 0 1px 2px rgba(124, 58, 237, 0.2);
  color: #ffffff;
}

.desktop-panel-heading__icon svg {
  width: 16px;
  height: 16px;
}

.desktop-panel-heading__title span:last-child {
  display: flex;
  flex-direction: column;
}

.desktop-panel-heading__title strong {
  color: #0a2463;
  font-size: 16px;
  line-height: 24px;
}

.desktop-panel-heading__title small {
  color: #90a1b9;
  font-size: 11px;
  line-height: 16.5px;
}

.desktop-panel-pill {
  min-height: 26.5px;
  padding: 5px 10px;
  gap: 6px;
  border: 1px solid rgba(196, 181, 253, 0.72);
  background: #f5f3ff;
  color: #7c3aed;
  font-size: 11px;
  line-height: 16.5px;
}

.desktop-panel-pill.is-completed {
  border-color: #b9f8cf;
  background: #f0fdf4;
  color: #00a63e;
}

.desktop-fallback-notice {
  margin: 16px 0 0;
  padding: 9px 11px;
  border-radius: 12px;
  background: rgba(249, 115, 22, 0.08);
  color: #c2410c;
  font-size: 12px;
  line-height: 18px;
}

.desktop-summary-body {
  min-height: 198px;
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.desktop-summary-body p {
  margin: 0;
  color: #314158;
  font-size: 16px;
  line-height: 26px;
}

.desktop-summary-body p.is-latest {
  color: #0a2463;
  animation: latestSummaryIn 260ms ease both;
}

.desktop-type-cursor {
  width: 1px;
  height: 17px;
  margin-left: 2px;
  display: inline-block;
  vertical-align: -3px;
  background: #7c3aed;
  animation: cursorBlink 920ms steps(1) infinite;
}

.desktop-recognized-empty {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: #62748e;
  font-size: 13px;
  line-height: 20px;
}

.desktop-summary-placeholder {
  min-height: 20px;
  margin-top: -48px;
  margin-left: 48px;
  display: inline-flex;
  align-items: center;
}

.desktop-analysis-done {
  margin-top: 18px;
  padding: 12px 14px;
  border: 1px solid rgba(6, 182, 212, 0.18);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background: linear-gradient(90deg, rgba(236, 254, 255, 0.72), rgba(245, 243, 255, 0.7));
}

.desktop-analysis-done p {
  margin: 0;
  color: #0891b2;
  font-size: 13px;
  line-height: 20px;
  font-weight: 700;
}

.desktop-analysis-done > span {
  min-height: 24px;
  padding: 4px 9px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(255, 255, 255, 0.72);
  color: #7c3aed;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
}

.desktop-recognized-panel {
  min-height: 423.5px;
  padding: 20px;
  background: linear-gradient(129deg, #ffffff 0%, #ffffff 50%, rgba(124, 58, 237, 0.05) 100%);
}

.desktop-recognized-panel__head {
  min-height: 42.5px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
}

.desktop-recognized-panel__head h2 {
  margin: 0;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.desktop-recognized-panel__head h2 svg {
  width: 16px;
  height: 16px;
  color: #06b6d4;
}

.desktop-recognized-panel__head p {
  margin: 4px 0 0;
  color: #90a1b9;
  font-size: 11px;
  line-height: 16.5px;
}

.desktop-recognized-panel__badge {
  min-height: 21px;
  padding: 3px 8px;
  gap: 4px;
  border: 1px solid #b9f8cf;
  background: #f0fdf4;
  color: #00a63e;
  font-size: 10px;
  line-height: 15px;
}

.desktop-recognized-panel__badge.is-empty {
  border-color: rgba(6, 182, 212, 0.25);
  background: #ecfeff;
  color: #0891b2;
}

.desktop-recognized-list {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.desktop-recognized-item {
  min-height: 48px;
  padding: 9px 12px;
  border: 1px solid rgba(6, 182, 212, 0.28);
  border-radius: 14px;
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) 14px;
  align-items: center;
  gap: 10px;
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.18), rgba(124, 58, 237, 0.14));
  box-shadow: 0 10px 20px -18px rgba(124, 58, 237, 0.45);
}

.desktop-recognized-item.is-violet {
  border-color: rgba(124, 58, 237, 0.3);
  background: linear-gradient(90deg, rgba(124, 58, 237, 0.14), rgba(249, 115, 22, 0.1));
}

.desktop-recognized-item.is-orange {
  border-color: rgba(249, 115, 22, 0.28);
  background: linear-gradient(90deg, rgba(249, 115, 22, 0.12), rgba(124, 58, 237, 0.12));
}

.desktop-recognized-item__icon {
  width: 28px;
  height: 28px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #06b6d4, #7c3aed);
  color: #ffffff;
}

.desktop-recognized-item.is-orange .desktop-recognized-item__icon {
  background: linear-gradient(135deg, #f97316, #7c3aed);
}

.desktop-recognized-item__icon svg {
  width: 14px;
  height: 14px;
}

.desktop-recognized-item__copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.desktop-recognized-item__copy small {
  color: #90a1b9;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.desktop-recognized-item__copy strong {
  overflow: hidden;
  color: #0a2463;
  font-size: 12px;
  line-height: 15px;
  font-weight: 700;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.desktop-recognized-item__check {
  width: 14px;
  height: 14px;
  color: #0891b2;
}

.desktop-recognized-empty {
  margin-top: 28px;
}

.desktop-task-chain-card {
  min-height: 200.5px;
  padding: 25px;
}

.desktop-task-chain-card__header {
  min-height: 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.desktop-task-chain-card__title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.desktop-task-chain-card__title > span {
  width: 32px;
  height: 32px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  color: #ffffff;
}

.desktop-task-chain-card__title svg {
  width: 16px;
  height: 16px;
}

.desktop-task-chain-card__title h2 {
  margin: 0;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 18px;
  line-height: 27px;
  font-weight: 700;
}

.desktop-task-chain-card__header p {
  margin: 0;
  color: #62748e;
  font-size: 12px;
  line-height: 16px;
}

.desktop-task-chain-card__header strong {
  color: #7c3aed;
}

.desktop-task-chain {
  position: relative;
  min-height: 98.5px;
  margin-top: 20px;
  padding: 0 8px;
}

.desktop-task-chain__track,
.desktop-task-chain__progress {
  position: absolute;
  left: 20px;
  top: 20px;
  height: 2px;
  border-radius: 999px;
}

.desktop-task-chain__track {
  right: 20px;
  background: #e2e8f0;
}

.desktop-task-chain__progress {
  max-width: calc(100% - 40px);
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
  transition: width 260ms ease;
}

.desktop-task-chain__list {
  position: relative;
  z-index: 1;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  list-style: none;
}

.desktop-task-chain__item {
  min-height: 98.5px;
  display: flex;
  align-items: center;
  flex-direction: column;
  text-align: center;
}

.desktop-task-chain__node {
  width: 40px;
  height: 40px;
  border: 2px solid #ffffff;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f1f5f9;
  color: #90a1b9;
  font-size: 12px;
  line-height: 1;
  font-weight: 700;
  box-shadow: inset 0 0 0 1px #cbd5e1, 0 1px 3px rgba(15, 23, 42, 0.08);
  transition: background 220ms ease, color 220ms ease, box-shadow 220ms ease, transform 220ms ease;
}

.desktop-task-chain__node svg {
  width: 20px;
  height: 20px;
}

.desktop-task-chain__item.is-completed .desktop-task-chain__node {
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  color: #ffffff;
  box-shadow: 0 4px 6px rgba(124, 58, 237, 0.3), 0 2px 4px rgba(124, 58, 237, 0.3);
}

.desktop-task-chain__item.is-current .desktop-task-chain__node {
  background: #ffffff;
  color: #7c3aed;
  box-shadow: inset 0 0 0 2px #7c3aed, 0 0 0 8px rgba(124, 58, 237, 0.11), 0 4px 12px rgba(124, 58, 237, 0.2);
  transform: scale(1.04);
  animation: desktopPulse 1200ms ease-in-out infinite;
}

.desktop-task-chain__item strong {
  margin-top: 8px;
  color: #0a2463;
  font-size: 12px;
  line-height: 15px;
  font-weight: 700;
}

.desktop-task-chain__item.is-pending strong,
.desktop-task-chain__item.is-pending small {
  color: #90a1b9;
}

.desktop-task-chain__item small {
  width: min(156px, 92%);
  margin-top: 8px;
  color: #62748e;
  font-size: 10px;
  line-height: 13.75px;
}

.desktop-complete-notice {
  min-height: 78px;
  padding: 17px 21px;
  border-color: #b9f8cf;
  display: flex;
  align-items: center;
  gap: 16px;
  background: linear-gradient(90deg, #f0fdf4, rgba(236, 254, 255, 0.6));
}

.desktop-complete-notice__icon {
  width: 40px;
  height: 40px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: linear-gradient(135deg, #05df72 0%, #00bc7d 100%);
  color: #ffffff;
  box-shadow: 0 1px 3px rgba(0, 201, 80, 0.3), 0 1px 2px rgba(0, 201, 80, 0.3);
}

.desktop-complete-notice__icon svg {
  width: 20px;
  height: 20px;
}

.desktop-complete-notice__copy {
  min-width: 0;
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
}

.desktop-complete-notice__copy strong {
  color: #016630;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
}

.desktop-complete-notice__copy small {
  color: rgba(0, 130, 54, 0.8);
  font-size: 14px;
  line-height: 20px;
}

.desktop-complete-notice__status {
  min-height: 30.5px;
  padding: 7px 13px;
  border: 1px solid rgba(185, 248, 207, 0.7);
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.7);
  color: #008236;
  font-size: 11px;
  line-height: 16.5px;
  font-weight: 700;
  white-space: nowrap;
}

.desktop-dot-loader {
  height: 6px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.desktop-dot-loader i {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  display: inline-block;
  background: #7c3aed;
  animation: desktopDot 900ms ease-in-out infinite;
}

.desktop-dot-loader.is-green i {
  background: #00bc7d;
}

.desktop-dot-loader i:nth-child(2) {
  animation-delay: 120ms;
}

.desktop-dot-loader i:nth-child(3) {
  animation-delay: 240ms;
}

.desktop-recognized-list-enter-active,
.desktop-recognized-list-leave-active,
.desktop-chain-in-enter-active,
.desktop-chain-in-leave-active,
.desktop-complete-in-enter-active,
.desktop-complete-in-leave-active {
  transition: opacity 260ms ease, transform 260ms ease;
}

.desktop-recognized-list-enter-from,
.desktop-recognized-list-leave-to,
.desktop-chain-in-enter-from,
.desktop-chain-in-leave-to,
.desktop-complete-in-enter-from,
.desktop-complete-in-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

@keyframes desktopSpin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes desktopDot {
  0%,
  80%,
  100% {
    transform: translateY(0);
    opacity: 0.45;
  }

  40% {
    transform: translateY(-2px);
    opacity: 1;
  }
}

@keyframes cursorBlink {
  0%,
  45% {
    opacity: 1;
  }

  46%,
  100% {
    opacity: 0;
  }
}

@keyframes latestSummaryIn {
  from {
    opacity: 0.65;
  }

  to {
    opacity: 1;
  }
}

@keyframes desktopPulse {
  0%,
  100% {
    transform: scale(1.04);
  }

  50% {
    transform: scale(1.1);
  }
}

@media (max-width: 1024px) {
  .desktop-agent-transition {
    padding: 32px 24px;
  }

  .desktop-agent-shell {
    width: min(940px, calc(100vw - 48px));
  }

  .desktop-agent-main-grid {
    grid-template-columns: minmax(0, 1fr) 320px;
  }
}
</style>
