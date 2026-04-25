<template>
  <div class="product-page">
    <header class="product-header">
      <div class="product-shell product-header__inner">
        <button type="button" class="brand-lockup" @click="goHome" aria-label="返回首页">
          <span class="brand-lockup__mark">
            <img :src="brandLogo" alt="" aria-hidden="true" class="brand-lockup__icon" />
          </span>
          <span class="brand-lockup__copy">
            <strong>数字丝路</strong>
            <small>Digital Silk Road</small>
          </span>
        </button>

        <nav class="product-nav" aria-label="产品页导航">
          <button type="button" class="product-nav__item" @click="goHome">
            首页
          </button>
          <button type="button" class="product-nav__item" :class="{ 'is-active': activeSection === 'product' }" @click="scrollToSection('product')">
            产品
          </button>
          <button type="button" class="product-nav__item" @click="goAboutPage">
            关于
          </button>
        </nav>

        <div class="product-header__actions">
          <button type="button" class="product-button product-button--primary product-button--compact" @click="createProject">
            开始使用
          </button>
        </div>
      </div>
    </header>

    <main class="product-main">
      <section ref="heroSection" data-section="product" class="product-section product-hero">
        <div class="product-shell product-shell--hero">
          <div class="product-hero__badge">
            <img :src="badgeSpark" alt="" aria-hidden="true" />
            <span>完整的跨境营销解决方案</span>
          </div>

          <h1>产品矩阵</h1>

          <p class="product-hero__description">
            <span>从合规判断到内容生成，从数字人营销到数据优化</span>
            <span>五大核心模块，构建完整的跨境电商AI工作流</span>
          </p>
        </div>
      </section>

      <section ref="modulesSection" data-section="product" class="product-section module-section">
        <div class="product-shell">
          <article
            v-for="(module, index) in modules"
            :key="module.title"
            class="module-row"
            :class="{ 'module-row--reverse': index % 2 === 1 }"
          >
            <div class="module-copy">
              <div class="module-copy__badge" :class="module.badgeClass">
                <img :src="module.icon" alt="" aria-hidden="true" />
              </div>
              <h2>{{ module.title }}</h2>
              <p class="module-copy__subtitle">{{ module.subtitle }}</p>
              <p class="module-copy__description">{{ module.description }}</p>
              <ul class="module-copy__list">
                <li v-for="point in module.points" :key="point">
                  <img :src="bulletCheck" alt="" aria-hidden="true" />
                  <span>{{ point }}</span>
                </li>
              </ul>
            </div>

            <div class="module-visual" :class="module.visualClass">
              <img :src="cornerDecor" alt="" aria-hidden="true" class="module-visual__decor" />
              <div class="module-visual__emoji">{{ module.visual }}</div>
            </div>
          </article>
        </div>
      </section>

      <section ref="scenariosSection" data-section="product" class="product-section scenario-section">
        <div class="product-shell">
          <div class="section-heading">
            <h2>适用场景</h2>
            <p>为不同类型的跨境业务提供专业解决方案</p>
          </div>

          <div class="scenario-grid">
            <article v-for="scenario in scenarios" :key="scenario.title" class="scenario-card">
              <div class="scenario-card__badge" :class="scenario.badgeClass">
                <img :src="scenario.icon" alt="" aria-hidden="true" />
              </div>
              <h3>{{ scenario.title }}</h3>
              <p class="scenario-card__description">{{ scenario.description }}</p>
              <ul class="scenario-card__list">
                <li v-for="point in scenario.points" :key="point">{{ point }}</li>
              </ul>
            </article>
          </div>
        </div>
      </section>

      <section class="product-section cta-section">
        <div class="product-shell product-shell--cta">
          <div class="cta-card">
            <img :src="cornerDecor" alt="" aria-hidden="true" class="cta-card__decor" />
            <img :src="ctaSpark" alt="" aria-hidden="true" class="cta-card__icon" />
            <h2>准备好开始了吗？</h2>
            <p>立即体验完整的跨境营销AI工作流</p>
            <button type="button" class="product-button product-button--light" @click="createProject">
              免费开始使用
            </button>
          </div>
        </div>
      </section>
    </main>

    <MarketingFooter />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { MarketingFooter } from '@/components/common'
import badgeSpark from '@/assets/product/badge-spark.svg'
import bulletCheck from '@/assets/product/bullet-check.svg'
import cornerDecor from '@/assets/product/corner-decor.png'
import ctaSpark from '@/assets/product/cta-spark.svg'
import featureAnalytics from '@/assets/product/feature-analytics.svg'
import featureCompliance from '@/assets/product/feature-compliance.svg'
import featureEditing from '@/assets/product/feature-editing.svg'
import featureLocalization from '@/assets/product/feature-localization.svg'
import featureMultimodal from '@/assets/product/feature-multimodal.svg'
import scenarioAgency from '@/assets/product/scenario-agency.svg'
import scenarioFactory from '@/assets/product/scenario-factory.svg'
import scenarioSeller from '@/assets/product/scenario-seller.svg'

type SectionKey = 'product'

const router = useRouter()
const brandLogo = '/logo_circle.png'
const activeSection = ref<SectionKey>('product')

const heroSection = ref<HTMLElement | null>(null)
const modulesSection = ref<HTMLElement | null>(null)
const scenariosSection = ref<HTMLElement | null>(null)

const modules = [
  {
    title: '智能合规检测系统',
    subtitle: 'Compliance Intelligence',
    description: '基于全球法规数据库的AI实时分析引擎',
    icon: featureCompliance,
    badgeClass: 'module-copy__badge--blue',
    visualClass: 'module-visual--blue',
    visual: '🛡️',
    points: ['覆盖200+国家地区准入法规', '实时更新政策变化', '风险等级智能评估', '合规建议自动生成', '多品类专业知识库']
  },
  {
    title: '多语种内容生成引擎',
    subtitle: 'Localization AI Engine',
    description: '深度理解文化差异的本地化内容创作',
    icon: featureLocalization,
    badgeClass: 'module-copy__badge--violet',
    visualClass: 'module-visual--violet',
    visual: '🌐',
    points: ['支持50+语言翻译', '文化适配智能优化', '营销话术本地化', 'SEO关键词自动优化', '品牌语调一致性保持']
  },
  {
    title: '多模态内容创作系统',
    subtitle: 'Multi-Modal Content Creation',
    description: 'AI生成商品图片、短视频、数字人口播，多种形式自由组合',
    icon: featureMultimodal,
    badgeClass: 'module-copy__badge--orange',
    visualClass: 'module-visual--orange',
    visual: '🎨',
    points: ['AI商品图片生成（主图/场景图/广告图）', '短视频片段智能生成', '300+特色语音包', '镜头拼接与视觉包装', '多种内容形式自由组合输出']
  },
  {
    title: '智能剪辑与成片系统',
    subtitle: 'Video Editing & Rendering',
    description: '智能拼接多模态内容，AI优化剪辑，一键输出成片',
    icon: featureEditing,
    badgeClass: 'module-copy__badge--green',
    visualClass: 'module-visual--green',
    visual: '🎬',
    points: ['图片/视频/数字人智能拼接', 'AI镜头转场与视觉包装', '智能配乐与音效', '多平台尺寸自动适配', '品牌元素一键植入']
  },
  {
    title: '数据智能分析中心',
    subtitle: 'Analytics & Insights',
    description: '全链路数据追踪与AI优化建议',
    icon: featureAnalytics,
    badgeClass: 'module-copy__badge--indigo',
    visualClass: 'module-visual--indigo',
    visual: '📊',
    points: ['实时数据监控看板', '转化漏斗分析', 'A/B测试智能建议', '市场趋势预测', 'ROI优化方案']
  }
]

const scenarios = [
  {
    title: '中小卖家',
    description: '快速完成合规检测与内容生成，降低出海门槛',
    icon: scenarioSeller,
    badgeClass: 'scenario-card__badge--blue',
    points: ['节省80%合规调研时间', '降低90%内容制作成本', '提升3倍市场拓展速度']
  },
  {
    title: '外贸工厂',
    description: '建立品牌形象，打造专业营销物料',
    icon: scenarioFactory,
    badgeClass: 'scenario-card__badge--violet',
    points: ['品牌数字化升级', '多语种营销物料', '全球市场覆盖']
  },
  {
    title: '代运营服务商',
    description: '提升服务效率与专业度，服务更多客户',
    icon: scenarioAgency,
    badgeClass: 'scenario-card__badge--indigo',
    points: ['批量处理客户需求', '标准化服务流程', '提升客户满意度']
  }
]

const createProject = () => {
  router.push('/dramas/create')
}

const goHome = () => {
  router.push('/')
}

const goAboutPage = () => {
  router.push('/about')
}

const getSectionElement = (key: SectionKey) => {
  switch (key) {
    case 'product':
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
      threshold: [0.22, 0.45, 0.68],
      rootMargin: '-18% 0px -50% 0px'
    }
  )

  ;[
    heroSection.value,
    modulesSection.value,
    scenariosSection.value
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
.product-page {
  min-height: var(--app-vh, 100vh);
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
  color: #0a2463;
}

.product-shell {
  width: min(1075px, calc(100% - 32px));
  margin: 0 auto;
  padding-inline: 32px;
}

.product-shell--hero {
  width: min(768px, calc(100% - 32px));
}

.product-shell--cta {
  width: min(896px, calc(100% - 32px));
}

.product-header {
  position: sticky;
  top: 0;
  z-index: 40;
  backdrop-filter: blur(18px);
  background: rgba(255, 255, 255, 0.8);
  border-bottom: 1px solid rgba(226, 232, 240, 0.6);
}

.product-header__inner {
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

.product-nav {
  display: flex;
  align-items: center;
  gap: 32px;
}

.product-nav__item,
.product-footer__link {
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

.product-nav__item {
  position: relative;
  font-weight: 500;
}

.product-nav__item::after {
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

.product-nav__item:hover,
.product-footer__link:hover {
  color: #0a2463;
}

.product-nav__item.is-active::after,
.product-nav__item:hover::after {
  transform: scaleX(1);
}

.product-button {
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
  transition: transform 180ms ease, box-shadow 180ms ease;
}

.product-button:hover {
  transform: translateY(-1px);
}

.product-button--primary {
  color: #ffffff;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
  box-shadow: 0 16px 32px -22px rgba(249, 115, 22, 0.72);
}

.product-button--light {
  background: #ffffff;
  color: #0a2463;
  box-shadow: 0 18px 36px -24px rgba(15, 23, 42, 0.32);
}

.product-button--compact {
  min-height: 44px;
  padding: 10px 20px;
  border-radius: 12px;
}

.product-section {
  scroll-margin-top: 104px;
}

.product-hero {
  padding: 96px 0 80px;
}

.product-hero .product-shell--hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.product-hero__badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 8px 16px;
  border-radius: 999px;
  border: 1px solid rgba(6, 182, 212, 0.2);
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.1) 0%, rgba(124, 58, 237, 0.1) 100%);
  color: #0a2463;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.product-hero__badge img {
  width: 16px;
  height: 16px;
}

.product-hero h1 {
  margin-top: 24px;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 48px;
  line-height: 48px;
  font-weight: 700;
  color: #0a2463;
}

.product-hero__description {
  margin-top: 24px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 20px;
  line-height: 32px;
  color: #45556c;
}

.module-section {
  padding-bottom: 80px;
}

.module-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(420px, 481.5px);
  align-items: center;
  gap: 48px;
}

.module-row + .module-row {
  margin-top: 64px;
}

.module-row--reverse .module-copy {
  order: 2;
}

.module-row--reverse .module-visual {
  order: 1;
}

.module-copy__badge,
.scenario-card__badge {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.module-copy__badge img,
.scenario-card__badge img {
  width: 32px;
  height: 32px;
}

.module-copy__badge--blue,
.scenario-card__badge--blue {
  background: linear-gradient(135deg, #2b7fff 0%, #00b8db 100%);
}

.module-copy__badge--violet,
.scenario-card__badge--violet {
  background: linear-gradient(135deg, #ad46ff 0%, #f6339a 100%);
}

.module-copy__badge--orange {
  background: linear-gradient(135deg, #ff6900 0%, #fb2c36 100%);
}

.module-copy__badge--green {
  background: linear-gradient(135deg, #00c950 0%, #00bc7d 100%);
}

.module-copy__badge--indigo,
.scenario-card__badge--indigo {
  background: linear-gradient(135deg, #4f7bff 0%, #6366f1 100%);
}

.module-copy h2 {
  margin-top: 24px;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 30px;
  line-height: 36px;
  font-weight: 700;
  color: #0a2463;
}

.module-copy__subtitle {
  margin-top: 8px;
  font-family: 'IBM Plex Sans', sans-serif;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  color: #06b6d4;
}

.module-copy__description {
  margin-top: 16px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 18px;
  line-height: 28px;
  color: #45556c;
}

.module-copy__list {
  list-style: none;
  margin-top: 24px;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.module-copy__list li {
  display: flex;
  align-items: center;
  gap: 12px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  color: #314158;
}

.module-copy__list li img {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.module-visual {
  position: relative;
  min-height: 481.5px;
  border-radius: 24px;
  overflow: hidden;
  box-shadow: 0 25px 50px -12px rgba(15, 23, 42, 0.25);
  display: flex;
  align-items: center;
  justify-content: center;
}

.module-visual__decor {
  position: absolute;
  inset: 0 auto auto 0;
  width: 60px;
  opacity: 0.2;
}

.module-visual__emoji {
  position: relative;
  font-size: 128px;
  line-height: 1;
}

.module-visual--blue {
  background: linear-gradient(135deg, #2b7fff 0%, #00b8db 100%);
}

.module-visual--violet {
  background: linear-gradient(135deg, #ad46ff 0%, #f6339a 100%);
}

.module-visual--orange {
  background: linear-gradient(135deg, #ff6900 0%, #fb2c36 100%);
}

.module-visual--green {
  background: linear-gradient(135deg, #00c950 0%, #00bc7d 100%);
}

.module-visual--indigo {
  background: linear-gradient(135deg, #4f7bff 0%, #6366f1 100%);
}

.scenario-section {
  padding: 80px 0;
  background: #ffffff;
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

.scenario-grid {
  margin-top: 64px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 32px;
}

.scenario-card {
  min-height: 338px;
  padding: 32px;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  background: linear-gradient(156deg, #ffffff 0%, #f8fafc 100%);
  box-shadow: 0 10px 20px -24px rgba(15, 23, 42, 0.24);
}

.scenario-card h3 {
  margin-top: 24px;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 24px;
  line-height: 32px;
  font-weight: 700;
  color: #0a2463;
}

.scenario-card__description {
  margin-top: 12px;
  min-height: 48px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  color: #45556c;
}

.scenario-card__list {
  list-style: none;
  margin-top: 24px;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.scenario-card__list li {
  position: relative;
  padding-left: 14px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  color: #314158;
}

.scenario-card__list li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 7px;
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
}

.cta-section {
  padding: 80px 0;
}

.cta-card {
  position: relative;
  overflow: hidden;
  min-height: 336px;
  border-radius: 24px;
  padding: 48px;
  background: linear-gradient(90deg, #0a2463 0%, #06b6d4 50%, #7c3aed 100%);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.cta-card__decor {
  position: absolute;
  inset: 0 auto auto 0;
  width: 60px;
  opacity: 0.2;
}

.cta-card__icon {
  width: 48px;
  height: 48px;
}

.cta-card h2 {
  margin-top: 24px;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 30px;
  line-height: 36px;
  font-weight: 700;
  color: #ffffff;
}

.cta-card p {
  margin-top: 16px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 18px;
  line-height: 28px;
  color: #cefafe;
}

.cta-card .product-button {
  margin-top: 32px;
}

.product-footer {
  background: linear-gradient(159deg, #0f172b 0%, #0a2463 50%, #0f172b 100%);
  color: #ffffff;
}

.product-footer__inner {
  padding-top: 64px;
  padding-bottom: 28px;
  --text-secondary: #90a1b9;
  --accent: #53eafd;
}

.product-footer__top {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 0.6fr) minmax(0, 0.6fr);
  gap: 48px;
}

.brand-lockup--footer .brand-lockup__mark {
  width: 52px;
  height: 52px;
  padding: 4px;
  border-radius: 16px;
}

.brand-lockup--footer .brand-lockup__copy strong,
.brand-lockup--footer .brand-lockup__copy small {
  color: #ffffff;
}

.brand-lockup--footer .brand-lockup__copy small {
  color: #53eafd;
}

.product-footer__summary {
  margin-top: 16px;
  max-width: 448px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  color: #cad5e2;
}

.product-footer__meta {
  margin-top: 24px;
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
}

.product-footer__meta-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  color: #90a1b9;
  text-decoration: none;
}

.product-footer__meta-link img {
  width: 16px;
  height: 16px;
}

.product-footer__column {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.product-footer__column h3 {
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 18px;
  line-height: 27px;
  font-weight: 600;
  color: #ffffff;
}

.product-footer__column .product-footer__link {
  text-align: left;
  color: #cad5e2;
}

.product-footer__bottom {
  margin-top: 48px;
  padding-top: 33px;
  border-top: 1px solid rgba(49, 65, 88, 0.5);
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 56px;
  flex-wrap: nowrap;
  overflow-x: auto;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  color: #90a1b9;
  white-space: nowrap;
}

.product-footer__meta-link--bottom,
.product-footer__legal,
.product-footer__beian {
  flex: 0 0 auto;
}

.product-footer__beian {
  justify-content: flex-start;
}

@media (max-width: 1040px) {
  .module-row,
  .module-row--reverse {
    grid-template-columns: 1fr;
    gap: 24px;
  }

  .module-row--reverse .module-copy,
  .module-row--reverse .module-visual {
    order: initial;
  }

  .scenario-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 920px) {
  .product-shell {
    width: min(100%, calc(100% - 24px));
    padding-inline: 20px;
  }

  .product-nav {
    display: none;
  }

  .product-footer__top {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .brand-lockup__copy small {
    display: none;
  }

  .product-hero h1 {
    font-size: 40px;
    line-height: 44px;
  }

  .product-hero__description,
  .section-heading p {
    font-size: 18px;
    line-height: 28px;
  }

  .module-visual {
    min-height: 280px;
  }

  .module-visual__emoji {
    font-size: 96px;
  }

  .scenario-grid,
  .product-footer__top {
    grid-template-columns: 1fr;
  }

  .scenario-card,
  .cta-card {
    padding: 24px;
  }

  .product-footer__meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .product-footer__bottom {
    gap: 36px;
  }
}
</style>
