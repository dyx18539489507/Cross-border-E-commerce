<template>
  <main class="mobile-agent-transition" aria-labelledby="mobile-agent-title">
    <div class="mobile-agent-transition__ambient mobile-agent-transition__ambient--cyan" aria-hidden="true"></div>
    <div class="mobile-agent-transition__ambient mobile-agent-transition__ambient--orange" aria-hidden="true"></div>

    <header class="mobile-agent-header">
      <button type="button" class="mobile-agent-header__back" aria-label="返回首页" @click="goHome">
        <ArrowLeft aria-hidden="true" />
      </button>

      <div class="mobile-agent-header__copy">
        <h1 id="mobile-agent-title">丝路 Agent</h1>
        <p>正在规划出海营销方案</p>
      </div>

      <span class="mobile-agent-header__status" :class="{ 'is-completed': isHeaderCompleted }">
        <CircleCheck v-if="isHeaderCompleted" aria-hidden="true" />
        <span v-else class="mobile-tiny-loader" aria-hidden="true"></span>
        <span>{{ statusLabel }}</span>
      </span>
    </header>

    <div class="mobile-agent-transition__scroll">
      <section class="mobile-user-bubble" aria-label="用户输入需求">
        <p>{{ userInputText }}</p>
      </section>

      <section v-if="uploadedImagePreview" class="mobile-user-image-card" aria-label="已上传商品图片">
        <div class="mobile-user-image-card__frame">
          <img :src="uploadedImagePreview" alt="已上传商品图片" />
        </div>
      </section>

      <section class="mobile-analysis-card" aria-label="Agent 分析内容">
        <div class="mobile-analysis-card__header">
          <div class="mobile-section-title">
            <span class="mobile-section-title__icon" aria-hidden="true">
              <MagicStick />
            </span>
            <span v-if="!analysisDone" class="mobile-analysis-title-dots" aria-hidden="true">
              <i></i><i></i><i></i>
            </span>
          </div>

        </div>

        <p v-if="fallbackNotice" class="mobile-analysis-card__notice">{{ fallbackNotice }}</p>

        <div class="mobile-analysis-card__body" aria-live="polite">
          <template v-if="summaryParagraphs.length">
            <p
              v-for="(paragraph, index) in summaryParagraphs"
              :key="`${paragraph}-${index}`"
              :class="{ 'is-latest': index === summaryParagraphs.length - 1 && !analysisDone }"
            >
              {{ paragraph }}
            </p>
          </template>
          <div v-if="recognizedInfo" class="mobile-recognized-tags" aria-label="已识别信息">
            <span v-for="tag in recognizedTags" :key="tag.label" class="mobile-recognized-tag" :class="`is-${tag.tone}`">
              <small>{{ tag.label }}</small>
              <strong>{{ tag.value }}</strong>
            </span>
          </div>
        </div>

        <div v-if="analysisDone" class="mobile-analysis-card__done">
          <p>内容分析已完成，正在编排 Agent 任务链……</p>
          <span>
            <span class="mobile-tiny-loader" aria-hidden="true"></span>
            任务链生成中
          </span>
        </div>
      </section>

      <section v-if="showTaskChain" class="mobile-task-card" aria-labelledby="mobile-task-title">
        <div class="mobile-task-card__header">
          <div class="mobile-section-title">
            <span class="mobile-section-title__plain-icon" aria-hidden="true">
              <Lightning />
            </span>
            <h2 id="mobile-task-title">Agent 自动任务链</h2>
          </div>
          <span class="mobile-task-card__progress">
            <strong>{{ completedTaskCount }}</strong><span>/{{ taskSteps.length }}</span>
          </span>
        </div>

        <div class="mobile-task-chain">
          <span class="mobile-task-chain__line" aria-hidden="true"></span>
          <span class="mobile-task-chain__progress-line" :style="{ height: taskProgressHeight }" aria-hidden="true"></span>

          <article
            v-for="task in taskSteps"
            :key="task.step"
            class="mobile-task-chain__item"
            :class="taskClass(task.step)"
          >
            <span class="mobile-task-chain__node" aria-hidden="true">
              <CircleCheck v-if="task.step <= completedTaskCount" />
              <span v-else>{{ task.step }}</span>
            </span>
            <span class="mobile-task-chain__copy">
              <strong>{{ task.name }}</strong>
              <small>{{ task.description }}</small>
            </span>
          </article>
        </div>
      </section>

      <section v-if="phase === 'completed'" class="mobile-complete-notice" aria-live="polite">
        <span class="mobile-complete-notice__icon" aria-hidden="true">
          <CircleCheck />
        </span>
        <span class="mobile-complete-notice__copy">
          <strong>所有任务节点已完成</strong>
          <small>正在进入生成结果页……</small>
        </span>
        <span class="mobile-complete-notice__jumping">
          <span class="mobile-dot-loader" aria-hidden="true"><i></i><i></i><i></i></span>
          自动跳转中
        </span>
      </section>

      <div ref="autoScrollAnchor" class="mobile-auto-scroll-anchor" aria-hidden="true"></div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { AgentInput, AgentResult } from '@/types/agent'
import { agentAPI } from '@/api/agent'
import { readAgentInput, readAgentUserInput, saveAgentResult } from '@/utils/agentStorage'
import {
  ArrowLeft,
  CircleCheck,
  Lightning,
  MagicStick
} from '@element-plus/icons-vue'

type Phase = 'analyzing' | 'arranging' | 'completed'

interface MobileRecognizedInfo {
  product: string
  category: string
  market: string
  platform: string
  audience: string
  sellingPoints: string
}

interface TaskStep {
  step: number
  name: string
  description: string
}

const DEFAULT_USER_INPUT = '我有一款便携榨汁杯，想卖到马来西亚，主要做 TikTok 短视频，目标用户是年轻女生，主打便携和健康。'

const router = useRouter()
const phase = ref<Phase>('analyzing')
const summaryText = ref('')
const recognizedInfo = ref<MobileRecognizedInfo | null>(null)
const analysisDone = ref(false)
const showTaskChain = ref(false)
const completedTaskCount = ref(0)
const fallbackNotice = ref('')
const generationReady = ref(false)
const allDone = ref(false)
const hasRedirected = ref(false)
const autoScrollAnchor = ref<HTMLElement | null>(null)

let abortController: AbortController | null = null
let redirectTimer: number | undefined
let resultFallbackTimer: number | undefined
let autoScrollFrame: number | undefined
const timers: number[] = []

function isUsableAgentImage(value: string) {
  return value.trim().toLowerCase().startsWith('data:image/')
}

const storedAgentInput = ref<AgentInput | null>(readAgentInput())
const userInputText = computed(() => {
  return storedAgentInput.value?.rawPrompt?.trim() || readAgentUserInput().trim() || DEFAULT_USER_INPUT
})
const uploadedImagePreview = computed(() => {
  const imageDataUrl = storedAgentInput.value?.imageDataUrl || ''
  return isUsableAgentImage(imageDataUrl) ? imageDataUrl : ''
})

const isHeaderCompleted = computed(() => {
  return phase.value === 'completed' && allDone.value && generationReady.value
})

const statusLabel = computed(() => {
  return isHeaderCompleted.value ? '已完成' : '分析中'
})

const summaryParagraphs = computed(() => {
  return summaryText.value
    .split(/\n+/)
    .map((item) => item.trim())
    .filter(Boolean)
})

const recognizedTags = computed(() => {
  const info = recognizedInfo.value
  if (!info) return []
  return [
    { label: '商品', value: info.product, tone: 'cyan' },
    { label: '商品类目', value: info.category, tone: 'cyan' },
    { label: '目标市场', value: info.market, tone: 'cyan' },
    { label: '目标平台', value: info.platform, tone: 'cyan' },
    { label: '目标人群', value: info.audience, tone: 'violet' },
    { label: '核心卖点', value: info.sellingPoints, tone: 'violet' }
  ].filter((tag) => tag.value)
})

const taskSteps: TaskStep[] = [
  { step: 1, name: '商品理解', description: '确认商品类目、卖点与使用场景' },
  { step: 2, name: '合规风险识别', description: '匹配目标市场规则与广告敏感表达' },
  { step: 3, name: '本地化方向', description: '生成符合马来西亚用户习惯的内容方向' },
  { step: 4, name: '短视频脚本', description: '生成开头、中段、结尾三段式脚本' },
  { step: 5, name: '数字人方案', description: '推荐数字人形象、口播语气与字幕语言' },
  { step: 6, name: '投放优化', description: '规划平台、内容方向与关键指标' }
]

const taskProgressHeight = computed(() => {
  if (completedTaskCount.value <= 0) return '0%'
  return `${Math.min(100, ((completedTaskCount.value - 1) / (taskSteps.length - 1)) * 100)}%`
})

const taskClass = (step: number) => {
  return {
    'is-completed': step <= completedTaskCount.value,
    'is-current': step === completedTaskCount.value + 1 && phase.value === 'arranging',
    'is-pending': step > completedTaskCount.value + 1
  }
}

const schedule = (callback: () => void, delay: number) => {
  const timer = window.setTimeout(callback, delay)
  timers.push(timer)
  return timer
}

const wait = (delay: number) => new Promise<void>((resolve) => {
  schedule(resolve, delay)
})

const scrollToLatest = (behavior: ScrollBehavior = 'smooth') => {
  if (autoScrollFrame) {
    window.cancelAnimationFrame(autoScrollFrame)
  }

  autoScrollFrame = window.requestAnimationFrame(async () => {
    await nextTick()
    autoScrollAnchor.value?.scrollIntoView({
      behavior,
      block: 'end'
    })
  })
}

const goHome = () => {
  router.push('/')
}

const buildAnalyzePayload = () => {
  const input = storedAgentInput.value
  return {
    scene: 'mobile_transition',
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
      scrollToLatest()
      break
    case 'analysis_done':
      markAnalysisDone()
      break
    case 'task_status':
      showTaskChain.value = true
      phase.value = 'arranging'
      completedTaskCount.value = Math.max(completedTaskCount.value, Number(data.step) || 0)
      scrollToLatest()
      break
    case 'all_done':
      completeAllTasks()
      break
    case 'fallback_notice':
    case 'error':
      fallbackNotice.value = data.message || '网络波动，已切换为本地演示流程'
      scrollToLatest()
      break
    default:
      break
  }
}

const appendSummary = (text: string) => {
  if (!text) return
  summaryText.value += text
  scrollToLatest()
}

const markAnalysisDone = () => {
  analysisDone.value = true
  phase.value = 'arranging'
  scrollToLatest()
  schedule(() => {
    showTaskChain.value = true
    scrollToLatest()
  }, 300)
}

const runLocalFallbackFlow = async () => {
  if (allDone.value) return
  fallbackNotice.value = '网络波动，已切换为本地演示流程'
  const info = currentRecognizedInfo()
  const paragraphs = [
    '已接收到你的出海需求，正在从自然语言中提取商品、目标市场、平台、人群和核心卖点。\n\n',
    `识别到商品为${info.product}，属于${info.category}，目标市场为${info.market}，主要投放平台为${info.platform}。\n\n`,
    '该商品涉及食品接触场景，后续合规分析将重点关注杯体材质、食品接触认证、电池容量和充电方式。\n\n',
    '接下来将基于合规边界，生成本地化营销方向、短视频脚本、数字人方案和投放建议。'
  ]

  for (const paragraph of paragraphs) {
    appendSummary(paragraph)
    await wait(360)
  }
  recognizedInfo.value = info
  scrollToLatest()
  markAnalysisDone()
  await runTaskFallback()
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
    scrollToLatest()
  }
  completeAllTasks()
}

const completeAllTasks = () => {
  completedTaskCount.value = taskSteps.length
  showTaskChain.value = true
  phase.value = 'completed'
  allDone.value = true
  scrollToLatest()
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
  }, 1200)
}

const currentRecognizedInfo = (): MobileRecognizedInfo => {
  return recognizedInfo.value || inferRecognizedInfo(storedAgentInput.value, userInputText.value)
}

const normalizeRecognizedInfo = (value: Partial<MobileRecognizedInfo>): MobileRecognizedInfo => {
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

const inferRecognizedInfo = (input: AgentInput | null, prompt: string): MobileRecognizedInfo => {
  const product = input?.productName || extractMatch(prompt, /(?:我有|这是一款|一款)([^，,。；;\n]{2,28}?)(?:，|,|。|想|主打|卖到|做|$)/) || '便携榨汁杯'
  const category = input?.category || (product.includes('榨汁杯') ? '小家电 / 食品接触用品' : '跨境电商商品')
  const market = input?.targetMarket || extractMatch(prompt, /(?:卖到|目标市场|进入|推广到)([^，,。；;\n]{2,24}?)(?:市场|，|,|。|$)/) || '马来西亚'
  const platform = input?.targetPlatform || extractPlatform(prompt) || 'TikTok'
  const audience = input?.targetAudience || extractMatch(prompt, /(?:目标用户|目标人群|用户是|人群是)(?:是|为)?([^，,。；;\n]{2,44})/) || '年轻女性 / 学生 / 办公室'
  const sellingPoints = input?.coreSellingPoints?.length
    ? input.coreSellingPoints.join(' / ')
    : extractMatch(prompt, /(?:主打|卖点|核心卖点)(?:是|为)?([^。；;\n]{2,80})/) || '便携 / 健康 / 易清洗'

  return { product, category, market, platform, audience, sellingPoints }
}

const extractMatch = (value: string, pattern: RegExp) => value.match(pattern)?.[1]?.trim() || ''

const extractPlatform = (value: string) => {
  const platforms = ['TikTok', 'Instagram', 'YouTube', 'Shopee', 'Lazada', 'Amazon', 'Temu', 'eBay', 'Facebook', '小红书', '抖音']
  const lowerValue = value.toLowerCase()
  return platforms.find((platform) => lowerValue.includes(platform.toLowerCase())) || ''
}

const buildFallbackInput = (): AgentInput => {
  return {
    requestId: createRequestId(),
    rawPrompt: userInputText.value,
    productName: currentRecognizedInfo().product,
    category: currentRecognizedInfo().category,
    targetMarket: currentRecognizedInfo().market,
    targetPlatform: currentRecognizedInfo().platform,
    targetAudience: currentRecognizedInfo().audience,
    coreSellingPoints: currentRecognizedInfo().sellingPoints.split(/[\/,，;；]/).map((item) => item.trim()).filter(Boolean)
  }
}

const buildFallbackResult = (input: AgentInput, info: MobileRecognizedInfo): AgentResult => {
  const sellingPoints = info.sellingPoints.split(/[\/,，;；]/).map((item) => item.trim()).filter(Boolean)
  return {
    schemaVersion: 2,
    requestId: input.requestId || createRequestId(),
    recognizedInfo: {
      productName: info.product,
      category: info.category,
      targetMarket: info.market,
      targetPlatform: info.platform,
      targetAudience: info.audience,
      coreSellingPoints: sellingPoints.length ? sellingPoints : ['便携', '健康', '易清洗'],
      imageUnderstanding: '移动端过渡页已完成基础商品识别。'
    },
    overview: {
      complianceRiskLevel: '中风险',
      marketStrategy: '先补齐食品接触材料与认证信息，再生成本地化素材。',
      recommendedVideoStyle: '生活场景化竖屏短视频',
      recommendedDigitalHuman: '亲和型本地化生活方式讲解者'
    },
    compliance: {
      title: '合规分析结果',
      summary: '当前已完成基础识别，建议补充商品材质、食品接触认证、电池容量与充电方式，以便进一步判断准入与广告表达边界。',
      riskTags: ['食品接触', '认证材料', '功效表达'],
      missingInfo: ['杯体材质', '食品接触认证', '电池容量', '充电方式'],
      suggestions: ['补充商品图片或详细描述，以便进行合规风险评估。'],
      forbiddenExpressions: ['100%安全', '永久有效', '官方认证', '治疗/治愈', '保证通过'],
      saferExpressions: ['适合日常使用', '建议查看材质说明', '以实际认证材料为准']
    },
    localization: {
      direction: '场景化种草 + 实用价值说明',
      reason: '在目标市场信息不足时，先聚焦高频生活场景和明确卖点，可降低夸大宣传风险。',
      keywords: ['portable', 'daily use', 'easy to use', 'lifestyle'],
      tone: '自然、可信、少夸张',
      sceneSuggestions: ['宿舍早餐', '办公室轻食', '通勤随身携带']
    },
    script: {
      title: '短视频脚本',
      duration: '20-25s',
      opening: { time: '0-3s', content: '用一个真实生活小问题切入，展示目标用户在日常场景中的痛点。' },
      middle: { time: '3-20s', content: '展示商品外观、核心卖点和使用步骤，强调便携、材质说明和适用场景，避免绝对化承诺。' },
      ending: { time: '20-25s', content: '用温和行动号召引导查看商品详情与认证信息。' },
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
      contentTags: ['生活场景', '便携', '健康饮品', '开箱演示'],
      focusMetrics: ['完播率', '点击率', '收藏率'],
      optimizationAdvice: '先用场景化短视频测试用户兴趣，再根据评论补充材质、认证和使用限制信息。'
    },
    agentMessage: {
      summary: '已完成移动端过渡页分析与任务编排，建议继续补充商品材料和认证信息。',
      missingInfoNotice: '建议补充杯体材质、食品接触认证、电池容量与充电方式。',
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
  startResultGeneration()
  startStreamingAnalysis()
})

onBeforeUnmount(() => {
  abortController?.abort()
  timers.forEach((timer) => window.clearTimeout(timer))
  if (redirectTimer) window.clearTimeout(redirectTimer)
  if (resultFallbackTimer) window.clearTimeout(resultFallbackTimer)
  if (autoScrollFrame) window.cancelAnimationFrame(autoScrollFrame)
})
</script>

<style scoped>
.mobile-agent-transition {
  position: relative;
  min-height: 100dvh;
  overflow-x: hidden;
  background:
    linear-gradient(144deg, #f8fafc 0%, #ffffff 50%, rgba(124, 58, 237, 0.05) 100%),
    linear-gradient(120deg, rgba(6, 182, 212, 0.08), rgba(249, 115, 22, 0.04));
  color: #0a2463;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', sans-serif;
}

.mobile-agent-transition__ambient {
  position: fixed;
  z-index: 0;
  pointer-events: none;
  border-radius: 999px;
  filter: blur(64px);
}

.mobile-agent-transition__ambient--cyan {
  left: 40px;
  top: -248px;
  width: 384px;
  height: 384px;
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.15), rgba(124, 58, 237, 0.15));
}

.mobile-agent-transition__ambient--orange {
  left: -167px;
  top: -56px;
  width: 500px;
  height: 500px;
  background: linear-gradient(135deg, rgba(249, 115, 22, 0.1), rgba(6, 182, 212, 0.1));
}

.mobile-agent-header {
  position: sticky;
  top: 0;
  z-index: 4;
  width: min(393px, 100%);
  min-height: calc(61px + env(safe-area-inset-top));
  margin: 0 auto;
  padding: calc(12px + env(safe-area-inset-top)) 20px 12px;
  border-bottom: 1px solid rgba(226, 232, 240, 0.6);
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(18px);
}

.mobile-agent-header__back {
  width: 36px;
  height: 36px;
  padding: 0;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: #ffffff;
  color: #45556c;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.1), 0 1px 2px rgba(15, 23, 42, 0.1);
}

.mobile-agent-header__back svg {
  width: 14px;
  height: 14px;
}

.mobile-agent-header__copy {
  min-width: 0;
  flex: 1 1 auto;
}

.mobile-agent-header__copy h1 {
  margin: 0;
  color: #0a2463;
  font-size: 16px;
  line-height: 20px;
  font-weight: 700;
}

.mobile-agent-header__copy p {
  margin: 0;
  overflow: hidden;
  color: #62748e;
  font-size: 11px;
  line-height: 16.5px;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.mobile-agent-header__status {
  min-height: 24px;
  padding: 4px 8px;
  border: 1px solid rgba(186, 230, 253, 0.9);
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  flex: 0 0 auto;
  background: #ecfeff;
  color: #0891b2;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
}

.mobile-agent-header__status svg {
  width: 10px;
  height: 10px;
}

.mobile-agent-header__status.is-completed {
  border-color: #b9f8cf;
  background: #f0fdf4;
  color: #00a63e;
}

.mobile-agent-transition__scroll {
  position: relative;
  z-index: 1;
  width: min(393px, 100%);
  margin: 0 auto;
  padding: 20px 20px calc(28px + env(safe-area-inset-bottom));
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.mobile-auto-scroll-anchor {
  width: 100%;
  height: 1px;
}

.mobile-user-bubble {
  width: min(293px, 88%);
  min-height: 92px;
  margin-left: auto;
  padding: 12px 16px;
  border-radius: 24px 24px 10px 24px;
  background: linear-gradient(163deg, #06b6d4 0%, #7c3aed 55%, #7c3aed 100%);
  box-shadow: 0 4px 6px rgba(124, 58, 237, 0.2), 0 2px 4px rgba(124, 58, 237, 0.2);
}

.mobile-user-bubble p {
  margin: 0;
  color: #ffffff;
  font-size: 14px;
  line-height: 22.75px;
}

.mobile-user-image-card {
  width: min(220px, 72%);
  margin-left: auto;
  padding: 6px;
  border: 1px solid rgba(124, 58, 237, 0.18);
  border-radius: 20px 20px 8px 20px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.94), rgba(246, 243, 255, 0.9));
  box-shadow: 0 12px 24px -18px rgba(15, 23, 42, 0.42), 0 10px 20px -20px rgba(124, 58, 237, 0.55);
}

.mobile-user-image-card__frame {
  overflow: hidden;
  aspect-ratio: 4 / 3;
  border-radius: 16px 16px 6px 16px;
  background: #f8fafc;
}

.mobile-user-image-card__frame img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.mobile-analysis-card,
.mobile-task-card {
  border: 1px solid #f1f5f9;
  border-radius: 24px;
  background: #ffffff;
  box-shadow: 0 10px 15px -3px rgba(226, 232, 240, 0.6), 0 4px 6px -4px rgba(226, 232, 240, 0.6);
}

.mobile-analysis-card {
  padding: 20px;
}

.mobile-analysis-card__header,
.mobile-task-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.mobile-section-title {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.mobile-section-title h2 {
  margin: 0;
  color: #0a2463;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.mobile-analysis-title-dots {
  height: 12px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.mobile-section-title__icon {
  width: 28px;
  height: 28px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  color: #ffffff;
}

.mobile-section-title__plain-icon {
  color: #7c3aed;
  display: inline-flex;
}

.mobile-section-title__icon svg,
.mobile-section-title__plain-icon svg {
  width: 14px;
  height: 14px;
}

.mobile-analysis-card__notice {
  margin: 12px 0 0;
  padding: 8px 10px;
  border-radius: 12px;
  background: rgba(249, 115, 22, 0.08);
  color: #c2410c;
  font-size: 11px;
  line-height: 17px;
}

.mobile-analysis-card__body {
  margin-top: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.mobile-analysis-card__body p {
  margin: 0;
  color: #314158;
  font-size: 14px;
  line-height: 22.75px;
}

.mobile-analysis-card__body p.is-latest {
  color: #0a2463;
}

.mobile-recognized-tags {
  margin: 2px 0 6px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px 6px;
}

.mobile-recognized-tag {
  max-width: 100%;
  min-height: 26px;
  padding: 5px 9px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 10px;
  line-height: 15px;
}

.mobile-recognized-tag.is-cyan {
  border: 1px solid rgba(6, 182, 212, 0.25);
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.08), rgba(124, 58, 237, 0.08));
}

.mobile-recognized-tag.is-violet {
  border: 1px solid rgba(124, 58, 237, 0.25);
  background: linear-gradient(90deg, rgba(124, 58, 237, 0.08), rgba(249, 115, 22, 0.08));
}

.mobile-recognized-tag small {
  color: #90a1b9;
  font-size: inherit;
  line-height: inherit;
  font-weight: 700;
}

.mobile-recognized-tag strong {
  min-width: 0;
  overflow: hidden;
  color: #0a2463;
  font-size: 11px;
  line-height: 16.5px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mobile-analysis-title-dots i,
.mobile-dot-loader i {
  width: 4px;
  height: 4px;
  border-radius: 999px;
  display: inline-block;
  background: #7c3aed;
  animation: mobileDot 900ms ease-in-out infinite;
}

.mobile-analysis-title-dots i:nth-child(2),
.mobile-dot-loader i:nth-child(2) {
  animation-delay: 120ms;
}

.mobile-analysis-title-dots i:nth-child(3),
.mobile-dot-loader i:nth-child(3) {
  animation-delay: 240ms;
}

.mobile-analysis-card__done {
  margin-top: 14px;
  padding: 12px;
  border: 1px solid rgba(6, 182, 212, 0.18);
  border-radius: 16px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  background: linear-gradient(90deg, rgba(236, 254, 255, 0.72), rgba(245, 243, 255, 0.7));
}

.mobile-analysis-card__done p {
  margin: 0;
  color: #0891b2;
  font-size: 12px;
  line-height: 18px;
  font-weight: 700;
}

.mobile-analysis-card__done > span {
  min-height: 22px;
  padding: 4px 8px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  flex: 0 0 auto;
  background: rgba(255, 255, 255, 0.72);
  color: #7c3aed;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
}

.mobile-tiny-loader {
  width: 10px;
  height: 10px;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 999px;
  display: inline-block;
  animation: mobileSpin 780ms linear infinite;
}

.mobile-task-card {
  padding: 20px;
}

.mobile-task-card__progress {
  color: #90a1b9;
  font-size: 11px;
  line-height: 16.5px;
}

.mobile-task-card__progress strong {
  color: #7c3aed;
  font-weight: 700;
}

.mobile-task-chain {
  position: relative;
  margin-top: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.mobile-task-chain__line,
.mobile-task-chain__progress-line {
  position: absolute;
  left: 15px;
  top: 8px;
  width: 2px;
  border-radius: 999px;
}

.mobile-task-chain__line {
  bottom: 8px;
  background: #f1f5f9;
}

.mobile-task-chain__progress-line {
  background: linear-gradient(180deg, #06b6d4 0%, #7c3aed 100%);
  transition: height 260ms ease;
}

.mobile-task-chain__item {
  position: relative;
  z-index: 1;
  min-height: 43px;
  display: grid;
  grid-template-columns: 32px minmax(0, 1fr);
  gap: 12px;
}

.mobile-task-chain__node {
  width: 32px;
  height: 32px;
  border: 2px solid #ffffff;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f1f5f9;
  color: #90a1b9;
  font-size: 11px;
  font-weight: 700;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.08);
  transition: background 220ms ease, color 220ms ease, box-shadow 220ms ease;
}

.mobile-task-chain__node svg {
  width: 16px;
  height: 16px;
}

.mobile-task-chain__item.is-completed .mobile-task-chain__node {
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  color: #ffffff;
  box-shadow: 0 1px 3px rgba(124, 58, 237, 0.3), 0 1px 2px rgba(124, 58, 237, 0.3);
}

.mobile-task-chain__item.is-current .mobile-task-chain__node {
  border-color: rgba(124, 58, 237, 0.26);
  background: #ffffff;
  color: #7c3aed;
  box-shadow: 0 0 0 6px rgba(124, 58, 237, 0.12), 0 1px 3px rgba(124, 58, 237, 0.22);
}

.mobile-task-chain__copy {
  min-width: 0;
  padding-top: 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.mobile-task-chain__copy strong {
  color: #0a2463;
  font-size: 14px;
  line-height: 17.5px;
  font-weight: 700;
}

.mobile-task-chain__copy small {
  overflow: hidden;
  color: #62748e;
  font-size: 12px;
  line-height: 19.5px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mobile-complete-notice {
  min-height: 67px;
  padding: 15px;
  border: 1px solid #b9f8cf;
  border-radius: 16px;
  display: grid;
  grid-template-columns: 36px minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  background: linear-gradient(90deg, #f0fdf4 0%, rgba(236, 254, 255, 0.6) 100%);
}

.mobile-complete-notice__icon {
  width: 36px;
  height: 36px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #05df72 0%, #00bc7d 100%);
  color: #ffffff;
  box-shadow: 0 1px 3px rgba(0, 201, 80, 0.3), 0 1px 2px rgba(0, 201, 80, 0.3);
}

.mobile-complete-notice__icon svg {
  width: 18px;
  height: 18px;
}

.mobile-complete-notice__copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.mobile-complete-notice__copy strong {
  color: #016630;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.mobile-complete-notice__copy small {
  color: rgba(0, 130, 54, 0.8);
  font-size: 12px;
  line-height: 16px;
}

.mobile-complete-notice__jumping {
  min-height: 24px;
  padding: 4px 9px;
  border: 1px solid rgba(185, 248, 207, 0.7);
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(255, 255, 255, 0.7);
  color: #008236;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
  white-space: nowrap;
}

.mobile-dot-loader {
  height: 4px;
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.mobile-dot-loader i {
  background: #00bc7d;
}

@keyframes mobileSpin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes mobileDot {
  0%,
  80%,
  100% {
    transform: translateY(0);
    opacity: 0.45;
  }

  40% {
    transform: translateY(-3px);
    opacity: 1;
  }
}

@media (max-width: 360px) {
  .mobile-agent-transition__scroll,
  .mobile-agent-header {
    padding-left: 16px;
    padding-right: 16px;
  }

  .mobile-complete-notice {
    grid-template-columns: 36px minmax(0, 1fr);
  }

  .mobile-complete-notice__jumping {
    grid-column: 2;
    justify-self: start;
  }
}
</style>
