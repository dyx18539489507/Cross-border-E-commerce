<!--
/**
 * 模块说明：数字丝路官网首页与丝路 Agent 启动入口。
 * 业务场景：用户在首页用自然语言描述商品、目标市场、平台、人群和卖点，并可附加商品图片。
 * 核心职责：先在前端做轻量字段抽取并缓存输入，再跳转到 Agent 过渡页发起后端流式分析。
 */
-->
<template>
  <div class="landing-page">
    <header class="landing-header">
      <div class="landing-shell landing-header__inner">
        <button type="button" class="brand-lockup" @click="scrollToSection('home')" aria-label="回到首页顶部">
          <span class="brand-lockup__mark">
            <img :src="brandLogo" alt="" aria-hidden="true" class="brand-lockup__icon" />
          </span>
          <span class="brand-lockup__copy">
            <strong>数字丝路</strong>
            <small>Digital Silk Road</small>
          </span>
        </button>

        <nav class="landing-nav" aria-label="首页导航">
          <button type="button" class="landing-nav__item" :class="{ 'is-active': activeSection === 'home' }" @click="scrollToSection('home')">
            首页
          </button>
          <button type="button" class="landing-nav__item" @click="openProductPage">
            产品
          </button>
          <button type="button" class="landing-nav__item" @click="openAboutPage">
            关于
          </button>
        </nav>

        <div class="landing-header__actions">
          <button type="button" class="landing-button landing-button--primary landing-button--compact" @click="createProject">
            开始使用
          </button>
        </div>
      </div>
    </header>

    <main class="landing-main">
      <section ref="heroSection" data-section="home" class="landing-section hero-section">
        <div class="landing-shell landing-shell--narrow hero-section__inner">
          <div class="hero-section__badge">
            <img :src="heroBadge" alt="" aria-hidden="true" class="hero-section__badge-icon" />
            <span>AI 驱动的跨境电商一体化平台</span>
          </div>

          <h1 class="hero-section__title">
            <span class="hero-section__title-top">数字丝路</span>
            <span class="hero-section__title-bottom">智启全球市场</span>
          </h1>

          <p class="hero-section__description">
            <span>从合规判断到内容生成，从数字人视频到数据优化</span>
            <span>为跨境卖家提供全链路 AI 解决方案</span>
          </p>

          <div class="hero-section__actions">
            <button type="button" class="landing-button landing-button--primary" @click="createProject">
              <span>立即开始</span>
              <img :src="heroArrow" alt="" aria-hidden="true" class="landing-button__icon" />
            </button>
            <button type="button" class="landing-button landing-button--agent" @click="scrollToAgentSection">
              <Cpu class="landing-button__agent-icon" aria-hidden="true" />
              <span class="landing-button__agent-text">体验丝路 Agent</span>
            </button>
            <button type="button" class="landing-button landing-button--secondary" @click="openProductPage">
              了解更多
            </button>
          </div>

          <div class="hero-stats" aria-label="平台核心数据">
            <article v-for="stat in stats" :key="stat.label" class="hero-stat">
              <strong class="hero-stat__value">{{ stat.value }}</strong>
              <span class="hero-stat__label">{{ stat.label }}</span>
            </article>
          </div>
        </div>
      </section>

      <section ref="productSection" data-section="home" class="landing-section feature-section">
        <div class="landing-shell">
          <div class="section-heading">
            <h2>核心能力</h2>
            <p>完整的跨境营销工作流，一站式解决方案</p>
          </div>

          <div class="feature-grid">
            <article v-for="feature in features" :key="feature.title" class="feature-card">
              <div class="feature-card__icon" :class="feature.tone">
                <img :src="feature.icon" alt="" aria-hidden="true" />
              </div>
              <h3>{{ feature.title }}</h3>
              <p>{{ feature.description }}</p>
            </article>
          </div>
        </div>
      </section>

      <section ref="agentSection" id="agent" class="landing-section agent-section" aria-labelledby="agent-title">
        <div class="landing-shell">
          <div class="agent-section__intro">
            <div class="agent-section__eyebrow">
              <Suitcase class="agent-section__eyebrow-icon" aria-hidden="true" />
              <span>SILK ROAD AGENT · 跨境电商智能体</span>
            </div>

            <h2 id="agent-title">
              <span>丝路 Agent：</span>
              <strong>一句话生成完整出海营销路径</strong>
            </h2>

            <p>
              输入商品信息,Agent自动理解需求、识别风险、生成本地化内容,并规划数字人营销与投放方案。
            </p>
          </div>

          <div class="agent-section__grid">
            <article class="agent-input-card" aria-label="丝路 Agent 输入演示">
              <div class="agent-input-card__header">
                <span class="agent-input-card__icon" aria-hidden="true">
                  <SuitcaseLine />
                </span>
                <div>
                  <h3>告诉丝路 Agent</h3>
                </div>
              </div>

              <div class="agent-input-card__field" :class="{ 'is-invalid': hasAgentPromptInvalidState }">
                <textarea
                  ref="agentPromptTextarea"
                  v-model="agentPrompt"
                  :class="{ 'is-invalid': hasAgentPromptInvalidState }"
                  :placeholder="agentPromptPlaceholder"
                  aria-label="输入商品与目标市场"
                  :aria-invalid="hasAgentPromptInvalidState"
                  required
                  rows="4"
                  @input="clearAgentPromptError"
                ></textarea>
              </div>
              <p v-if="agentPromptError" class="agent-input-card__error">{{ agentPromptError }}</p>

              <label class="agent-image-upload-box">
                <input type="file" accept="image/*" @change="handleAgentImageChange" />
                <span class="agent-image-upload-box__icon" aria-hidden="true">
                  <UploadFilled />
                </span>
                <span class="agent-image-upload-box__content">
                  <strong>{{ agentImageName || '上传商品图片（可选）' }}</strong>
                </span>
              </label>

              <div class="agent-input-card__tags" aria-label="快捷指令标签">
                <button v-for="tag in agentPromptTags" :key="tag" type="button" @click="appendAgentTag(tag)">
                  {{ tag }}
                </button>
              </div>

              <button
                type="button"
                class="agent-input-card__action"
                :disabled="agentStarting"
                @pointerenter="preloadRouteByPath('/agent/transition')"
                @focus="preloadRouteByPath('/agent/transition')"
                @click="openAgentPage"
              >
                <Lightning class="agent-input-card__action-icon" aria-hidden="true" />
                <span>{{ agentStarting ? '分析中' : '启动丝路 Agent' }}</span>
                <ArrowRight class="agent-input-card__action-icon" aria-hidden="true" />
              </button>
            </article>

            <article class="agent-chain-card" aria-label="丝路 Agent 能力链路">
              <div class="agent-chain-card__header">
                <span class="agent-chain-card__title">
                  <MagicStick aria-hidden="true" />
                  <span>智能任务链</span>
                </span>
              </div>

              <div class="agent-chain">
                <article
                  v-for="step in agentSteps"
                  :key="step.title"
                  class="agent-chain__item"
                >
                  <div class="agent-chain__icon-wrap" :class="`agent-chain__icon-wrap--${step.tone}`">
                    <component :is="step.icon" class="agent-chain__icon" aria-hidden="true" />
                    <span class="agent-chain__index">{{ step.index }}</span>
                  </div>
                  <div class="agent-chain__content">
                    <h3>{{ step.title }}</h3>
                    <p>{{ step.description }}</p>
                  </div>
                  <CircleCheck class="agent-chain__check" aria-hidden="true" />
                </article>
              </div>
            </article>
          </div>

          <p class="agent-section__note">
            不是简单填表生成,而是由
            <span class="agent-text-cyan">Agent 理解模糊需求</span>、<span class="agent-text-violet">主动补全信息</span>、<span class="agent-text-orange">自动规划任务</span>,并持续优化跨境营销方案。
          </p>
        </div>
      </section>

      <section ref="workflowSection" data-section="home" class="landing-section workflow-section">
        <div class="landing-shell">
          <div class="section-heading">
            <h2>工作流程</h2>
            <p>5步完成从商品到营销的全链路</p>
          </div>

          <div class="workflow-grid">
            <article v-for="step in workflowSteps" :key="step.index" class="workflow-step">
              <strong class="workflow-step__index">{{ step.index }}</strong>
              <h3>{{ step.title }}</h3>
              <p>{{ step.description }}</p>
            </article>
          </div>
        </div>
      </section>

      <section ref="testimonialsSection" data-section="home" class="landing-section testimonial-section">
        <div class="landing-shell">
          <div class="section-heading">
            <h2>客户见证</h2>
            <p>来自客户的真实反馈</p>
          </div>

          <div class="testimonial-grid">
            <article v-for="testimonial in testimonials" :key="testimonial.name" class="testimonial-card">
              <div class="testimonial-card__stars" aria-hidden="true">
                <img v-for="index in 5" :key="index" :src="ratingStar" alt="" class="testimonial-card__star" />
              </div>
              <blockquote>{{ testimonial.quote }}</blockquote>
              <div class="testimonial-card__author">
                <span class="testimonial-card__avatar">
                  <img :src="testimonial.avatar" alt="" aria-hidden="true" />
                </span>
                <div>
                  <strong>{{ testimonial.name }}</strong>
                  <span>{{ testimonial.role }}</span>
                </div>
              </div>
            </article>
          </div>
        </div>
      </section>

      <section class="landing-section cta-section">
        <img :src="ctaDecor" alt="" aria-hidden="true" class="cta-section__decor" />
        <div class="landing-shell landing-shell--narrow cta-section__inner">
          <span class="cta-section__badge">
            <img :src="ctaGlobe" alt="" aria-hidden="true" />
          </span>
          <h2>开启全球市场新篇章</h2>
          <p>从合规准入到智能营销，让AI助力商品更快走向全球</p>
          <button type="button" class="landing-button landing-button--light" @click="createProject">
            <span>免费开始使用</span>
            <img :src="ctaArrow" alt="" aria-hidden="true" class="landing-button__icon" />
          </button>
        </div>
      </section>
    </main>

    <MarketingFooter />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { agentAPI } from '@/api/agent'
import { preloadRouteByPath } from '@/router'
import {
  ArrowRight,
  CircleCheck,
  Cpu,
  Goods,
  Guide,
  Lightning,
  MagicStick,
  Suitcase,
  SuitcaseLine,
  TrendCharts,
  UploadFilled,
  User
} from '@element-plus/icons-vue'
import { MarketingFooter } from '@/components/common'
import type { AgentInput } from '@/types/agent'
import { clearAgentResult, saveAgentInput, saveAgentUserInput } from '@/utils/agentStorage'
import avatarChenYue from '@/assets/landing/avatar-chen-yue.webp'
import avatarLiMing from '@/assets/landing/avatar-li-ming.webp'
import avatarWangXuan from '@/assets/landing/avatar-wang-xuan.webp'
import ctaArrow from '@/assets/landing/cta-arrow.svg'
import ctaDecor from '@/assets/landing/cta-decor.png'
import ctaGlobe from '@/assets/landing/cta-globe.svg'
import featureAnalytics from '@/assets/landing/feature-analytics.svg'
import featureCompliance from '@/assets/landing/feature-compliance.svg'
import featureLocalization from '@/assets/landing/feature-localization.svg'
import featureMultimodal from '@/assets/landing/feature-multimodal.svg'
import heroArrow from '@/assets/landing/hero-arrow.svg'
import heroBadge from '@/assets/landing/hero-badge.svg'
import ratingStar from '@/assets/landing/rating-star.svg'

type SectionKey = 'home' | 'about'

const router = useRouter()
const activeSection = ref<SectionKey>('home')
const brandLogo = '/logo_circle.png'

const heroSection = ref<HTMLElement | null>(null)
const productSection = ref<HTMLElement | null>(null)
const agentSection = ref<HTMLElement | null>(null)
const workflowSection = ref<HTMLElement | null>(null)
const testimonialsSection = ref<HTMLElement | null>(null)
const agentPromptTextarea = ref<HTMLTextAreaElement | null>(null)
const agentPrompt = ref('')
const agentPromptError = ref('')
const agentPromptInvalid = ref(false)
const agentStarting = ref(false)
const agentImageDataUrl = ref('')
const agentImageName = ref('')
const agentPromptPlaceholder = '请输入商品信息。例如：我有一款便携榨汁杯，想卖到马来西亚，主要做 TikTok 短视频，目标用户是年轻女生，主打便携和健康。'
const defaultAgentUserInput = '请补充商品、目标市场、内容平台、目标用户和核心卖点。'
const hasAgentPromptInvalidState = computed(() => agentPromptInvalid.value || !!agentPromptError.value)

const stats = [
  { value: '1000+', label: '商品类目' },
  { value: '200+', label: '覆盖国家' },
  { value: '50+', label: '支持语言' },
  { value: '95%', label: '合规准确率' }
]

const features = [
  {
    title: '智能合规判断',
    description: '基于全球200+国家法规数据库，AI实时分析商品准入风险',
    icon: featureCompliance,
    tone: 'feature-card__icon--blue'
  },
  {
    title: '多语种本地化',
    description: '支持50+语言的AI内容生成，深度理解文化差异与营销习惯',
    icon: featureLocalization,
    tone: 'feature-card__icon--violet'
  },
  {
    title: '多模态内容生成',
    description: 'AI生成商品图片、短视频、数字人口播，多种形式自由组合',
    icon: featureMultimodal,
    tone: 'feature-card__icon--orange'
  },
  {
    title: '数据智能分析',
    description: '全链路数据追踪，AI优化建议，持续提升转化率',
    icon: featureAnalytics,
    tone: 'feature-card__icon--green'
  }
]

const workflowSteps = [
  { index: '01', title: '商品录入', description: '快速导入商品信息' },
  { index: '02', title: '合规检测', description: 'AI分析准入风险' },
  { index: '03', title: '脚本/分镜', description: '智能生成营销脚本' },
  { index: '04', title: '内容创作', description: '图片/视频/数字人' },
  { index: '05', title: '成片输出', description: '剪辑优化与发布' }
]

const agentPromptTags = [
  '商品准入分析',
  '生成本地化脚本',
  '数字人口播方案',
  '投放优化建议'
]

const agentSteps = [
  {
    index: '1',
    icon: Goods,
    tone: 'cyan',
    title: '商品理解',
    description: '识别商品类目、卖点、目标市场和用户人群'
  },
  {
    index: '2',
    icon: CircleCheck,
    tone: 'blue',
    title: '合规分析',
    description: '判断禁限售风险、认证要求和广告敏感表达'
  },
  {
    index: '3',
    icon: Guide,
    tone: 'violet',
    title: '本地化内容',
    description: '生成目标市场语言下的标题、卖点和短视频脚本'
  },
  {
    index: '4',
    icon: User,
    tone: 'pink',
    title: '数字人方案',
    description: '推荐数字人形象、口播语气、字幕语言和视频风格'
  },
  {
    index: '5',
    icon: TrendCharts,
    tone: 'orange',
    title: '投放建议',
    description: '给出平台选择、内容方向和优化指标建议'
  }
]

const testimonials = [
  {
    quote: '"使用数字丝路后，合规问题减少90%，视频制作效率提升10倍，ROI提升3倍。"',
    name: '李明',
    role: '跨境电商卖家 · 深圳某3C品牌',
    avatar: avatarLiMing
  },
  {
    quote: '"本地化内容质量非常出色，我们在新市场的转化率提升了240%。"',
    name: '陈悦',
    role: '品牌市场负责人 · 杭州某服饰品牌',
    avatar: avatarChenYue
  },
  {
    quote: '"为客户提供服务时，数字丝路让我们的效率和专业度都大幅提升。"',
    name: '王轩',
    role: '代运营服务商 · 某跨境代运营公司',
    avatar: avatarWangXuan
  }
]

const createProject = () => {
  router.push({ path: '/projects/create', query: { source: 'manual' } })
}

const openProductPage = () => {
  router.push('/products')
}

const openAboutPage = () => {
  router.push('/about')
}

const openAgentPage = async () => {
  if (!agentPrompt.value.trim()) {
    agentPromptInvalid.value = true
    agentPromptError.value = ''
    agentPromptTextarea.value?.focus()
    return
  }

  if (agentStarting.value) {
    return
  }

  agentStarting.value = true
  // 首页先做一次轻量抽取，是为了让过渡页立刻有商品/市场等可展示字段，后端模型稍后再给出更完整判断。
  const localExtracted = extractAgentInputFromPrompt(agentPrompt.value)
  const baseInput: AgentInput = {
    requestId: createAgentRequestId(),
    ...localExtracted,
    rawPrompt: agentPrompt.value.trim(),
    imageDataUrl: isUsableAgentImage(agentImageDataUrl.value) ? agentImageDataUrl.value : ''
  }

  try {
    const remoteExtracted = await extractAgentInputWithBackend(baseInput)
    const input: AgentInput = {
      ...baseInput,
      ...remoteExtracted,
      requestId: baseInput.requestId,
      rawPrompt: baseInput.rawPrompt,
      imageDataUrl: baseInput.imageDataUrl
    }

    preloadRouteByPath('/agent/transition')
    // Agent 输入和原话分开存储：结构化字段服务分析，原话服务用户可见摘要和信息缺口判断。
    saveAgentInput(input)
    saveAgentUserInput(input.rawPrompt || defaultAgentUserInput)
    clearAgentResult()
    await router.push('/agent')
  } finally {
    agentStarting.value = false
  }
}

const extractAgentInputWithBackend = async (input: AgentInput) => {
  try {
    // 后端抽取和正式 Agent 使用同一套规则，首页提前调用可以避免前后端识别口径不一致。
    return await agentAPI.extract(input)
  } catch (error) {
    console.warn('Silkroad Agent extract fallback to local parser.', error)
    return input
  }
}

const clearAgentPromptError = () => {
  if (agentPrompt.value.trim()) {
    agentPromptInvalid.value = false
  }
  if (agentPromptError.value && agentPrompt.value.trim()) {
    agentPromptError.value = ''
  }
}

const parseAgentSellingPoints = (value: string) => {
  return value
    .split(/[,，;；\n]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

/**
 * 功能：从首页自然语言里提取 Agent 初始上下文。
 * 参数：value 为用户输入的商品描述，可能包含商品、国家、平台、人群、卖点和材质信息。
 * 返回：AgentInput 的部分字段；缺失字段保留为空，交由后端 Agent 在信息缺口中提示。
 */
const extractAgentInputFromPrompt = (value: string): AgentInput => {
  const prompt = value.trim()
  if (!prompt) return {}

  const productName = extractPromptProductName(prompt)

  const targetMarket = cleanupExtractedText(extractFirstMatch(prompt, [
    /(?:卖到|卖|出口到|进入|面向|投放到|推广到|去|给到)([^，,。；;\n]{1,24}?)(?:市场|用户|消费者|，|,|。|；|;|$)/,
    /(?:目标市场|目标国家|国家|市场)(?:是|为|:|：)?([^，,。；;\n]{2,24})/
  ])) || inferKnownMarket(prompt)

  const targetAudience = cleanupExtractedText(extractFirstMatch(prompt, [
    /(?:目标用户|目标人群|受众|面向用户)(?:是|为|:|：)?([^，,。；;\n]{2,44})/,
    /(?:用户是|人群是)([^，,。；;\n]{2,44})/
  ]))

  const category = cleanupExtractedText(extractFirstMatch(prompt, [
    /(?:商品类目|产品类目|类目|品类|属于)(?:是|为|:|：)?([^，,。；;\n]{2,28})/
  ]))

  const materialSpec = cleanupExtractedText(extractFirstMatch(prompt, [
    /(?:材质|成分|容量|规格|尺寸|型号)(?:是|为|:|：)?([^。；;\n]{2,52})/
  ]))

  const usageScenario = cleanupExtractedText(extractFirstMatch(prompt, [
    /(?:使用场景|应用场景|场景)(?:是|为|:|：)?([^。；;\n]{2,64})/
  ]))

  const pointsText = cleanupExtractedText(extractFirstMatch(prompt, [
    /(?:核心卖点|卖点|主打|突出)(?:是|为|:|：)?([^。；;\n]{2,80})/
  ]))

  return {
    productName: cleanupExtractedText(productName),
    category: category || inferCategoryByProduct(productName),
    targetMarket,
    targetPlatform: extractTargetPlatform(prompt),
    targetAudience,
    coreSellingPoints: parseAgentSellingPoints(pointsText).length
      ? parseAgentSellingPoints(pointsText)
      : inferSellingPointsByProduct(productName, category),
    materialSpec,
    usageScenario
  }
}

const extractFirstMatch = (value: string, patterns: RegExp[]) => {
  for (const pattern of patterns) {
    const match = value.match(pattern)
    const captured = match?.[1]?.trim()
    if (captured) return captured
  }
  return ''
}

const extractPromptProductName = (value: string) => {
  const taggedProduct = extractFirstMatch(value, [
    /(?:商品准入分析|商品|产品|品名|商品名称|产品名称)\s*[:：]\s*([^，,。；;\n]{1,28})/,
    /^(?:商品准入分析|商品|产品|品名|商品名称|产品名称)\s*[:：]\s*([^\n]{1,28})/
  ])
  if (taggedProduct) {
    return cleanupExtractedProductName(taggedProduct)
  }

  const naturalProduct = extractFirstMatch(value, [
    /^(?:帮我|请|麻烦)?(?:分析一下|分析|看看|测一下|识别一下|识别)?\s*([^，,。；;\n]{1,28}?)(?:卖到|卖|出口到|进入|面向|投放到|推广到|上架|发布到|做|去|给|给到|$)/,
    /(?:我有|我们有|这是一款|这款|一款|一个|一种|商品是|产品是)([^，,。；;\n]{2,28}?)(?:，|,|。|；|;|想|计划|准备|主打|目标|卖到|出口|做|$)/,
    /(?:销售|卖)([^，,。；;\n]{2,28}?)(?:，|,|。|；|;|到|去|$)/
  ])
  if (naturalProduct) {
    return cleanupExtractedProductName(naturalProduct)
  }

  const compactPrompt = value
    .split(/\n+/)
    .map((line) => cleanupExtractedProductName(line))
    .find((line) => line && line.length <= 18 && !agentPromptTags.some((tag) => line.includes(tag)))

  return compactPrompt || ''
}

const cleanupExtractedProductName = (value: string) => {
  return cleanupExtractedText(value)
    .split(/(?:生成本地化脚本|数字人口播方案|投放优化建议|商品准入分析)\s*[:：]?/)[0]
    .replace(/^(分析|识别|检测)/, '')
    .trim()
}

const cleanupExtractedText = (value?: string) => {
  return (value || '')
    .replace(/TikTok|Instagram Reels|Instagram|YouTube Shorts|YouTube|Shopee|Lazada|Amazon|Temu|eBay|Facebook|小红书|抖音/gi, '')
    .replace(/^(的|到|给|为|是|做|在)/, '')
    .replace(/(市场|平台|用户|人群)$/g, '')
    .trim()
}

const inferKnownMarket = (value: string) => {
  const markets = [
    '美国',
    '英国',
    '加拿大',
    '澳大利亚',
    '德国',
    '法国',
    '日本',
    '韩国',
    '马来西亚',
    '新加坡',
    '泰国',
    '越南',
    '印尼',
    '印度尼西亚',
    '菲律宾',
    '印度',
    '墨西哥',
    '巴西',
    '沙特',
    '阿联酋',
    '中东',
    '欧洲',
    '东南亚',
    '北美'
  ]
  const lowerValue = value.toLowerCase()
  return markets.find((market) => lowerValue.includes(market.toLowerCase())) || ''
}

const isFoodProduct = (product = '', category = '') => {
  return /食品|饮料|零食|即食|餐|鸡|鸭|肉|鱼|虾|糕|饼|糖|茶|咖啡|炸|烤|卤|吃/.test(`${product} ${category}`)
}

const inferCategoryByProduct = (product = '') => {
  if (!product) return ''
  if (product.includes('榨汁杯') || product.includes('杯')) return '小家电 / 食品接触用品'
  if (isFoodProduct(product)) return '食品饮料 / 即食食品'
  return ''
}

const inferSellingPointsByProduct = (product = '', category = '') => {
  if (!product) return []
  if (isFoodProduct(product, category)) return ['口味', '便捷', '场景化']
  if (product.includes('榨汁杯') || product.includes('杯')) return ['便携', '健康', '易清洗']
  return []
}

const extractTargetPlatform = (value: string) => {
  const platforms = [
    'TikTok',
    'Instagram Reels',
    'Instagram',
    'YouTube Shorts',
    'YouTube',
    'Shopee',
    'Lazada',
    'Amazon',
    'Temu',
    'eBay',
    'Facebook',
    '小红书',
    '抖音'
  ]
  const lowerValue = value.toLowerCase()
  return platforms.find((platform) => lowerValue.includes(platform.toLowerCase())) || ''
}

const createAgentRequestId = () => {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

const isUsableAgentImage = (value: string) => {
  return value.trim().toLowerCase().startsWith('data:image/')
}

const appendAgentTag = (tag: string) => {
  const prefix = agentPrompt.value.trim()
  agentPrompt.value = prefix ? `${prefix}\n${tag}: ` : `${tag}: `
}

const handleAgentImageChange = (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    agentImageDataUrl.value = ''
    agentImageName.value = ''
    return
  }
  if (!file.type.startsWith('image/')) {
    input.value = ''
    agentImageDataUrl.value = ''
    agentImageName.value = ''
    return
  }

  // 图片以 Data URL 暂存到会话中，过渡页和后端视觉模型可以在不依赖正式上传接口的情况下直接读取。
  const reader = new FileReader()
  reader.onload = () => {
    agentImageDataUrl.value = typeof reader.result === 'string' ? reader.result : ''
    agentImageName.value = file.name
  }
  reader.readAsDataURL(file)
}

const scrollToAgentSection = () => {
  if (!agentSection.value) return

  const top = agentSection.value.getBoundingClientRect().top + window.scrollY
  window.scrollTo({
    top,
    behavior: 'smooth',
  })
}

const getSectionElement = (key: SectionKey) => {
  switch (key) {
    case 'home':
      return heroSection.value
    default:
      return null
  }
}

const scrollToSection = (key: SectionKey) => {
  activeSection.value = key
  getSectionElement(key)?.scrollIntoView({
    behavior: 'smooth',
    block: 'start'
  })
}

let observer: IntersectionObserver | null = null

onMounted(() => {
  if (typeof IntersectionObserver === 'undefined') return

  observer = new IntersectionObserver(
    (entries) => {
      const visibleEntry = entries
        .filter((entry) => entry.isIntersecting)
        .sort((left, right) => right.intersectionRatio - left.intersectionRatio)[0]

      if (!visibleEntry) return

      const section = visibleEntry.target.getAttribute('data-section') as SectionKey | null
      if (section) {
        activeSection.value = section
      }
    },
    {
      threshold: [0.2, 0.45, 0.7],
      rootMargin: '-18% 0px -52% 0px'
    }
  )

  ;[
    heroSection.value,
    productSection.value,
    workflowSection.value,
    testimonialsSection.value
  ]
    .filter((element): element is HTMLElement => Boolean(element))
    .forEach((element) => observer?.observe(element))
})

onBeforeUnmount(() => {
  observer?.disconnect()
  observer = null
})
</script>

<style scoped>
.landing-page {
  min-height: var(--app-vh, 100vh);
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
  color: #0a2463;
}

.landing-shell {
  width: min(1075px, calc(100% - 32px));
  margin: 0 auto;
  padding-inline: 32px;
}

.landing-shell--narrow {
  width: min(896px, calc(100% - 32px));
}

.landing-header {
  position: sticky;
  top: 0;
  z-index: 40;
  backdrop-filter: blur(18px);
  background: rgba(255, 255, 255, 0.8);
  border-bottom: 1px solid rgba(226, 232, 240, 0.6);
}

.landing-header__inner {
  min-height: 76px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.brand-lockup {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
  color: inherit;
  text-align: left;
}

.brand-lockup__mark {
  width: 56px;
  height: 56px;
  padding: 4px;
  border-radius: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(226, 232, 240, 0.96);
  box-shadow: 0 16px 28px -24px rgba(15, 23, 42, 0.28);
  flex-shrink: 0;
}

.brand-lockup__icon {
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: 999px;
}

.brand-lockup__copy {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.brand-lockup__copy strong {
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 18px;
  line-height: 28px;
  font-weight: 700;
  color: #0a2463;
}

.brand-lockup__copy small {
  font-family: 'IBM Plex Sans', 'Segoe UI', sans-serif;
  font-size: 12px;
  line-height: 16px;
  color: #62748e;
}

.landing-nav {
  display: flex;
  align-items: center;
  gap: 32px;
}

.landing-nav__item,
.landing-footer__link {
  padding: 0;
  border: 0;
  background: transparent;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  color: #314158;
  cursor: pointer;
  transition: color 180ms ease, opacity 180ms ease;
}

.landing-nav__item {
  position: relative;
  font-weight: 500;
}

.landing-nav__item::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: -6px;
  height: 2px;
  border-radius: 999px;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
  transform: scaleX(0);
  transform-origin: left;
  transition: transform 180ms ease;
}

.landing-nav__item:hover,
.landing-footer__link:hover {
  color: #0a2463;
}

.landing-nav__item.is-active::after,
.landing-nav__item:hover::after {
  transform: scaleX(1);
}

.landing-header__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.landing-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 56px;
  padding: 16px 24px;
  border: 0;
  border-radius: 16px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  cursor: pointer;
  transition: transform 180ms ease, box-shadow 180ms ease, opacity 180ms ease;
}

.landing-button:hover {
  transform: translateY(-1px);
}

.landing-button--primary {
  color: #ffffff;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
  box-shadow: 0 16px 32px -22px rgba(249, 115, 22, 0.72);
}

.landing-button--secondary {
  min-height: 60px;
  background: #ffffff;
  color: #0a2463;
  border: 2px solid #e2e8f0;
  box-shadow: 0 10px 24px -24px rgba(15, 23, 42, 0.28);
}

.landing-button--agent {
  position: relative;
  min-height: 60px;
  width: 207px;
  padding: 16px 32px;
  color: #06b6d4;
  border: 2px solid transparent;
  border-radius: 16px;
  background:
    linear-gradient(#ffffff, #ffffff) padding-box,
    linear-gradient(164deg, #06b6d4 0%, #7c3aed 100%) border-box;
  box-shadow: 0 10px 24px -22px rgba(15, 23, 42, 0.35);
}

.landing-button--agent:hover {
  box-shadow: 0 18px 34px -24px rgba(124, 58, 237, 0.5);
}

.landing-button__agent-icon {
  width: 20px;
  height: 20px;
  flex: 0 0 auto;
}

.landing-button__agent-text {
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  white-space: nowrap;
}

.landing-button--light {
  background: #ffffff;
  color: #0a2463;
  box-shadow: 0 18px 36px -24px rgba(15, 23, 42, 0.32);
}

.landing-button--compact {
  min-height: 44px;
  padding: 10px 20px;
  border-radius: 12px;
}

.landing-button__icon {
  width: 20px;
  height: 20px;
}

.landing-main {
  position: relative;
}

.landing-section {
  scroll-margin-top: 104px;
}

.hero-section {
  padding: 128px 0 80px;
}

.hero-section__inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.hero-section__badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 8px 16px;
  border-radius: 999px;
  border: 1px solid rgba(6, 182, 212, 0.2);
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.1) 0%, rgba(124, 58, 237, 0.1) 100%);
  color: #06b6d4;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.hero-section__badge-icon {
  width: 16px;
  height: 16px;
}

.hero-section__title {
  margin-top: 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 72px;
  line-height: 90px;
  font-weight: 700;
  letter-spacing: -0.04em;
}

.hero-section__title-top {
  color: #0a2463;
}

.hero-section__title-bottom,
.hero-stat__value {
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 52%, #f97316 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.hero-section__description {
  margin-top: 24px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  color: #45556c;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 20px;
  line-height: 32px;
}

.hero-section__actions {
  margin-top: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 28px;
  flex-wrap: wrap;
}

.hero-stats {
  width: 100%;
  margin-top: 80px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 32px;
}

.hero-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  text-align: center;
}

.hero-stat__value {
  font-family: 'IBM Plex Sans', sans-serif;
  font-size: 36px;
  line-height: 40px;
  font-weight: 700;
}

.hero-stat__label {
  color: #45556c;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
}

.agent-section {
  position: relative;
  box-sizing: border-box;
  min-height: 100vh;
  min-height: 100svh;
  padding: 80px 0;
  scroll-margin-top: 0;
  overflow: hidden;
  background: linear-gradient(138deg, #0a2463 0%, #0f2f7a 50%, #1e3a8a 100%);
}

.agent-section::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(circle at 42% 16%, rgba(6, 182, 212, 0.22) 0, rgba(6, 182, 212, 0) 240px),
    radial-gradient(circle at 58% 74%, rgba(124, 58, 237, 0.22) 0, rgba(124, 58, 237, 0) 260px),
    linear-gradient(135deg, rgba(255, 255, 255, 0.04) 0 1px, transparent 1px 32px);
  opacity: 1;
}

.agent-section .landing-shell {
  position: relative;
  width: min(1103px, 100%);
  padding-inline: 32px;
}

.agent-section__intro {
  min-height: 178px;
  text-align: center;
}

.agent-section__eyebrow {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 38px;
  padding: 8px 16px;
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.1);
  color: #cefafe;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.agent-section__eyebrow-icon {
  width: 16px;
  height: 16px;
  color: #06b6d4;
}

.agent-section__intro h2 {
  margin-top: 20px;
  display: flex;
  justify-content: center;
  gap: 12px;
  flex-wrap: wrap;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 48px;
  line-height: 48px;
  font-weight: 700;
  color: #ffffff;
}

.agent-section__intro h2 strong {
  background: linear-gradient(90deg, #06b6d4 0%, #a78bfa 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.agent-section__intro p {
  width: min(768px, 100%);
  margin: 16px auto 0;
  color: rgba(206, 250, 254, 0.8);
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 18px;
  line-height: 28px;
}

.agent-section__grid {
  margin-top: 48px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: stretch;
  gap: 24px;
}

.agent-input-card,
.agent-chain-card {
  min-height: 530px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.1);
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.25);
}

.agent-input-card {
  padding: 28px;
}

.agent-input-card__header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.agent-input-card__icon {
  position: relative;
  width: 44px;
  height: 44px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  box-shadow: 0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.1);
  color: #ffffff;
}

.agent-input-card__icon svg {
  width: 24px;
  height: 24px;
}

.agent-input-card__header div {
  min-width: 0;
}

.agent-input-card h3 {
  margin: 0;
  color: #ffffff;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
}

.agent-input-card__header p {
  margin-top: 0;
  color: rgba(162, 244, 253, 0.7);
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 12px;
  line-height: 16px;
}

.agent-input-card__field {
  position: relative;
  margin-top: 20px;
  min-height: 80px;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  background: rgba(15, 23, 43, 0.4);
}

.agent-input-card__field textarea {
  width: 100%;
  height: 92px;
  padding: 0;
  border: 0;
  outline: 0;
  resize: none;
  overflow-y: auto;
  scrollbar-color: rgba(206, 250, 254, 0.28) transparent;
  scrollbar-width: thin;
  scrollbar-gutter: stable;
  background: transparent;
  color: rgba(236, 254, 255, 0.9);
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 23px;
}

.agent-input-card__field.is-invalid {
  border-color: rgba(248, 113, 113, 0.78);
  background: rgba(127, 29, 29, 0.16);
  box-shadow: 0 0 0 3px rgba(248, 113, 113, 0.1);
}

.agent-input-card__field textarea.is-invalid::placeholder {
  color: rgba(254, 202, 202, 0.62);
}

.agent-input-card__field textarea::placeholder {
  color: rgba(206, 250, 254, 0.5);
}

.agent-input-card__field textarea::-webkit-scrollbar {
  width: 5px;
}

.agent-input-card__field textarea::-webkit-scrollbar-button {
  display: none;
  width: 0;
  height: 0;
}

.agent-input-card__field textarea::-webkit-scrollbar-track {
  border-radius: 999px;
  background: transparent;
}

.agent-input-card__field textarea::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgba(206, 250, 254, 0.24);
  box-shadow: inset 0 0 0 1px rgba(6, 182, 212, 0.08);
}

.agent-input-card__field textarea::-webkit-scrollbar-thumb:hover {
  background: rgba(206, 250, 254, 0.38);
}

.agent-input-card__field textarea::-webkit-scrollbar-corner {
  background: transparent;
}

.agent-input-card__error {
  margin-top: 8px;
  color: #fecaca;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 13px;
  line-height: 20px;
}

.agent-image-upload-box {
  margin-top: 12px;
  min-height: 48px;
  padding: 10px 14px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(15, 23, 43, 0.28);
  color: rgba(206, 250, 254, 0.86);
  cursor: pointer;
  transition: border-color 180ms ease, background 180ms ease, transform 180ms ease;
}

.agent-image-upload-box:hover {
  border-color: rgba(6, 182, 212, 0.45);
  background: rgba(6, 182, 212, 0.08);
  transform: translateY(-1px);
}

.agent-image-upload-box input {
  display: none;
}

.agent-image-upload-box__icon {
  width: 28px;
  height: 28px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.22), rgba(124, 58, 237, 0.2));
  color: #cefafe;
}

.agent-image-upload-box__icon svg {
  width: 17px;
  height: 17px;
}

.agent-image-upload-box__content {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.agent-image-upload-box__content strong {
  overflow: hidden;
  color: #cefafe;
  font-size: 13px;
  line-height: 20px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.agent-text-cyan {
  color: #06b6d4;
  font-weight: 700;
}

.agent-text-violet {
  color: #a78bfa;
  font-weight: 700;
}

.agent-text-orange {
  color: #f97316;
  font-weight: 700;
}

.agent-input-card__tags {
  margin-top: 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.agent-input-card__tags button {
  min-height: 30px;
  padding: 7px 13px;
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.05);
  color: #cefafe;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 12px;
  line-height: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: transform 180ms ease, border-color 180ms ease, background-color 180ms ease;
}

.agent-input-card__tags button::before {
  content: '# ';
}

.agent-input-card__tags button:hover {
  transform: translateY(-1px);
  border-color: rgba(167, 139, 250, 0.48);
  background: rgba(255, 255, 255, 0.09);
}

.agent-input-card__action {
  margin-top: 20px;
  width: 100%;
  min-height: 52px;
  border: 0;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #ffffff;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 50%, #f97316 100%);
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  cursor: pointer;
  transition: transform 180ms ease, box-shadow 180ms ease;
}

.agent-input-card__action:hover {
  transform: translateY(-1px);
  box-shadow: 0 20px 34px -22px rgba(6, 182, 212, 0.7);
}

.agent-input-card__action:disabled {
  cursor: wait;
  opacity: 0.72;
  transform: none;
  box-shadow: none;
}

.agent-input-card__action-icon {
  width: 20px;
  height: 20px;
  flex: 0 0 auto;
}

.agent-chain-card {
  padding: 28px 28px 1px;
}

.agent-chain-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.agent-chain-card__title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #ffffff;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
}

.agent-chain-card__title svg {
  width: 20px;
  height: 20px;
  color: #06b6d4;
}

.agent-chain {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.agent-chain__item {
  position: relative;
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr) 16px;
  align-items: center;
  gap: 12px;
  min-height: 70px;
  padding: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.05);
}

.agent-chain__item:not(:last-child)::after {
  content: '';
  position: absolute;
  left: 26px;
  top: 70px;
  width: 1px;
  height: 10px;
  background: linear-gradient(180deg, rgba(6, 182, 212, 0.6) 0%, rgba(6, 182, 212, 0) 100%);
}

.agent-chain__icon-wrap {
  position: relative;
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  box-shadow: 0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.1);
}

.agent-chain__icon-wrap--cyan {
  background: linear-gradient(135deg, #00d3f3 0%, #2b7fff 100%);
}

.agent-chain__icon-wrap--blue {
  background: linear-gradient(135deg, #2b7fff 0%, #615fff 100%);
}

.agent-chain__icon-wrap--violet {
  background: linear-gradient(135deg, #615fff 0%, #ad46ff 100%);
}

.agent-chain__icon-wrap--pink {
  background: linear-gradient(135deg, #ad46ff 0%, #f6339a 100%);
}

.agent-chain__icon-wrap--orange {
  background: linear-gradient(135deg, #f6339a 0%, #ff6900 100%);
}

.agent-chain__icon {
  width: 20px;
  height: 20px;
}

.agent-chain__index {
  position: absolute;
  right: -4px;
  top: -4px;
  width: 16px;
  height: 16px;
  border: 1px solid #06b6d4;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #0a2463;
  color: #06b6d4;
  font-family: 'IBM Plex Sans', sans-serif;
  font-size: 9px;
  line-height: 14px;
  font-weight: 700;
}

.agent-chain__content {
  min-width: 0;
}

.agent-chain__content h3 {
  margin: 0;
  color: #ffffff;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.agent-chain__content p {
  margin-top: 2px;
  color: rgba(206, 250, 254, 0.6);
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 12px;
  line-height: 20px;
}

.agent-chain__check {
  width: 16px;
  height: 16px;
  color: #06b6d4;
  opacity: 0.76;
}

.agent-section__note {
  width: min(768px, 100%);
  margin: 40px auto 0;
  color: rgba(206, 250, 254, 0.7);
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 26px;
  font-weight: 400;
  text-align: center;
}

.feature-section,
.testimonial-section {
  padding: 80px 0;
  background: #ffffff;
}

.workflow-section {
  padding: 80px 0;
  background: linear-gradient(156deg, #f8fafc 0%, #f1f5f9 100%);
}

.section-heading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  text-align: center;
}

.section-heading h2 {
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 36px;
  line-height: 40px;
  font-weight: 700;
  color: #0a2463;
}

.section-heading p {
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 20px;
  line-height: 28px;
  color: #45556c;
}

.feature-grid {
  margin-top: 64px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 32px;
}

.feature-card {
  min-height: 216px;
  padding: 32px;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  background: linear-gradient(156deg, #ffffff 0%, #f8fafc 100%);
}

.feature-card__icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.feature-card__icon img {
  width: 28px;
  height: 28px;
}

.feature-card__icon--blue {
  background: linear-gradient(135deg, #2b7fff 0%, #00b8db 100%);
}

.feature-card__icon--violet {
  background: linear-gradient(135deg, #ad46ff 0%, #f6339a 100%);
}

.feature-card__icon--orange {
  background: linear-gradient(135deg, #ff6900 0%, #fb2c36 100%);
}

.feature-card__icon--green {
  background: linear-gradient(135deg, #00c950 0%, #00bc7d 100%);
}

.feature-card h3 {
  margin-top: 24px;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 24px;
  line-height: 32px;
  font-weight: 700;
  color: #0a2463;
}

.feature-card p {
  margin-top: 12px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 26px;
  color: #45556c;
}

.workflow-grid {
  margin-top: 64px;
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 24px;
}

.workflow-step {
  position: relative;
  min-height: 168px;
  padding: 24px;
  border-radius: 16px;
  background: #ffffff;
  box-shadow: 0 10px 15px rgba(15, 23, 42, 0.1), 0 4px 6px rgba(15, 23, 42, 0.08);
}

.workflow-step:not(:last-child)::after {
  content: '';
  position: absolute;
  right: -24px;
  top: 83px;
  width: 24px;
  height: 2px;
  background: linear-gradient(90deg, #06b6d4 0%, rgba(6, 182, 212, 0) 100%);
}

.workflow-step__index {
  display: block;
  font-family: 'IBM Plex Sans', sans-serif;
  font-size: 48px;
  line-height: 48px;
  font-weight: 700;
  background: linear-gradient(140deg, #06b6d4 0%, #7c3aed 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.workflow-step h3 {
  margin-top: 16px;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 18px;
  line-height: 28px;
  font-weight: 700;
  color: #0a2463;
}

.workflow-step p {
  margin-top: 8px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  color: #45556c;
}

.testimonial-grid {
  margin-top: 64px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 32px;
}

.testimonial-card {
  min-height: 268px;
  padding: 32px;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  background: linear-gradient(140deg, #ffffff 0%, #f8fafc 100%);
}

.testimonial-card__stars {
  display: flex;
  gap: 4px;
}

.testimonial-card__star {
  width: 20px;
  height: 20px;
}

.testimonial-card blockquote {
  margin-top: 16px;
  min-height: 78px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 26px;
  color: #314158;
}

.testimonial-card__author {
  margin-top: 24px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.testimonial-card__avatar {
  width: 48px;
  height: 48px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: #e2e8f0;
  border: 1px solid #dbe4ee;
  flex-shrink: 0;
}

.testimonial-card__avatar img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
}

.testimonial-card__author strong {
  display: block;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  color: #0a2463;
}

.testimonial-card__author span {
  display: block;
  margin-top: 2px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  color: #45556c;
}

.cta-section {
  position: relative;
  overflow: hidden;
  padding: 80px 0;
  background: linear-gradient(158deg, #0a2463 0%, #06b6d4 50%, #7c3aed 100%);
}

.cta-section__decor {
  position: absolute;
  left: 0;
  top: 0;
  width: 60px;
  opacity: 0.3;
  pointer-events: none;
}

.cta-section__inner {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.cta-section__badge {
  width: 64px;
  height: 64px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.cta-section__badge img {
  width: 64px;
  height: 64px;
}

.cta-section h2 {
  margin-top: 24px;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 48px;
  line-height: 48px;
  font-weight: 700;
  color: #ffffff;
}

.cta-section p {
  margin-top: 24px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 20px;
  line-height: 28px;
  color: #cefafe;
}

.cta-section .landing-button {
  margin-top: 40px;
}

.landing-footer {
  background: linear-gradient(159deg, #0f172b 0%, #0a2463 50%, #0f172b 100%);
  color: #ffffff;
}

.landing-footer__inner {
  padding-top: 64px;
  padding-bottom: 28px;
  --text-secondary: #90a1b9;
  --accent: #53eafd;
}

.landing-footer__top {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 0.6fr) minmax(0, 0.6fr);
  gap: 48px;
}

.brand-lockup--footer .brand-lockup__copy strong,
.brand-lockup--footer .brand-lockup__copy small {
  color: #ffffff;
}

.brand-lockup--footer .brand-lockup__mark {
  width: 52px;
  height: 52px;
  padding: 4px;
  border-radius: 16px;
}

.brand-lockup--footer .brand-lockup__copy small {
  color: #53eafd;
}

.landing-footer__summary {
  margin-top: 16px;
  max-width: 448px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  color: #cad5e2;
}

.landing-footer__meta {
  margin-top: 24px;
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
}

.landing-footer__meta-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  color: #90a1b9;
  text-decoration: none;
}

.landing-footer__meta-link img {
  width: 16px;
  height: 16px;
}

.landing-footer__column {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.landing-footer__column h3 {
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 18px;
  line-height: 27px;
  font-weight: 600;
  color: #ffffff;
}

.landing-footer__column .landing-footer__link {
  text-align: left;
  color: #cad5e2;
}

.landing-footer__bottom {
  margin-top: 20px;
  padding-top: 24px;
  border-top: 1px solid rgba(49, 65, 88, 0.5);
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 16px;
  flex-wrap: nowrap;
  overflow-x: visible;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 12px;
  line-height: 18px;
  color: #90a1b9;
  white-space: nowrap;
}

.landing-footer__meta-link--bottom,
.landing-footer__legal,
.landing-footer__beian {
  min-width: 0;
  flex: 0 0 auto;
  white-space: nowrap;
}

.landing-footer__beian {
  justify-content: flex-start;
  column-gap: 12px;
}

.landing-footer__beian.beian-records {
  min-width: 0;
  column-gap: 12px;
}

.landing-footer__beian :deep(.beian-record-link) {
  font-size: 11px;
  line-height: 18px;
}

.landing-footer__beian :deep(.beian-record-icon) {
  width: 16px;
  height: 16px;
}

@media (max-width: 1100px) {
  .workflow-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .workflow-step::after {
    display: none;
  }
}

@media (max-width: 920px) {
  .landing-shell {
    width: min(100%, calc(100% - 24px));
    padding-inline: 20px;
  }

  .landing-nav {
    display: none;
  }

  .hero-section__title {
    font-size: 56px;
    line-height: 68px;
  }

  .hero-stats,
  .feature-grid,
  .testimonial-grid,
  .landing-footer__top {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .workflow-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .workflow-step::after {
    display: none;
  }

  .cta-section h2 {
    font-size: 40px;
    line-height: 44px;
  }

  .agent-section__intro h2 {
    font-size: 42px;
    line-height: 50px;
  }

  .agent-section__grid {
    grid-template-columns: 1fr;
  }

  .agent-input-card,
  .agent-chain-card {
    min-height: auto;
  }
}

@media (max-width: 640px) {
  .landing-header__inner {
    gap: 12px;
  }

  .brand-lockup__copy small,
  .landing-header__actions .landing-link {
    display: none;
  }

  .landing-button--compact {
    padding-inline: 16px;
  }

  .hero-section {
    padding-top: 88px;
  }

  .hero-section__badge {
    font-size: 12px;
    line-height: 16px;
  }

  .hero-section__title {
    font-size: 40px;
    line-height: 50px;
  }

  .hero-section__description,
  .section-heading p,
  .cta-section p {
    font-size: 18px;
    line-height: 28px;
  }

  .hero-section__actions {
    width: 100%;
    gap: 12px;
  }

  .hero-section__actions .landing-button {
    width: min(100%, 280px);
  }

  .hero-stats,
  .feature-grid,
  .workflow-grid,
  .testimonial-grid,
  .landing-footer__top {
    grid-template-columns: 1fr;
  }

  .agent-section {
    min-height: auto;
    padding: 64px 0;
  }

  .agent-section .landing-shell {
    padding-inline: 20px;
  }

  .agent-section__intro {
    min-height: auto;
  }

  .agent-section__intro h2 {
    font-size: 36px;
    line-height: 42px;
  }

  .agent-section__intro p,
  .agent-section__note {
    font-size: 16px;
    line-height: 26px;
  }

  .agent-section__grid {
    margin-top: 32px;
    gap: 20px;
  }

  .agent-input-card,
  .agent-chain-card {
    border-radius: 16px;
  }

  .agent-input-card {
    padding: 22px;
  }

  .agent-input-card h3 {
    font-size: 16px;
    line-height: 24px;
  }

  .agent-input-card__field {
    min-height: 112px;
    padding: 16px;
    border-radius: 16px;
  }

  .agent-input-card__tags {
    gap: 10px;
  }

  .agent-input-card__tags button {
    flex: 1 1 calc(50% - 10px);
    padding-inline: 10px;
  }

  .agent-chain-card {
    padding: 22px;
  }

  .agent-chain-card__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .agent-chain__item {
    grid-template-columns: 40px minmax(0, 1fr) 16px;
    gap: 12px;
  }

  .agent-chain__item:not(:last-child)::after {
    left: 26px;
  }

  .agent-chain__content {
    padding: 0;
  }

  .feature-card,
  .workflow-step,
  .testimonial-card {
    padding: 24px;
  }

  .cta-section {
    padding: 64px 0;
  }

  .cta-section h2 {
    font-size: 32px;
    line-height: 38px;
  }

  .landing-footer__meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .landing-footer__bottom {
    gap: 12px;
    font-size: 11px;
  }

  .landing-footer__beian {
    column-gap: 12px;
  }

  .landing-footer__beian :deep(.beian-record-link) {
    font-size: 10px;
  }
}
</style>
