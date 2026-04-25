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
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { MarketingFooter } from '@/components/common'
import avatarChenYue from '@/assets/landing/avatar-chen-yue.webp'
import avatarLiMing from '@/assets/landing/avatar-li-ming.webp'
import avatarWangFang from '@/assets/landing/avatar-wang-fang.webp'
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
const workflowSection = ref<HTMLElement | null>(null)
const testimonialsSection = ref<HTMLElement | null>(null)

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
    name: '王芳',
    role: '代运营服务商 · 某跨境代运营公司',
    avatar: avatarWangFang
  }
]

const createProject = () => {
  router.push('/dramas/create')
}

const openProductPage = () => {
  router.push('/products')
}

const openAboutPage = () => {
  router.push('/about')
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

  .hero-stats,
  .feature-grid,
  .workflow-grid,
  .testimonial-grid,
  .landing-footer__top {
    grid-template-columns: 1fr;
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
