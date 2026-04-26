<template>
  <div class="pricing-page">
    <header class="pricing-header">
      <div class="pricing-shell pricing-header__inner">
        <button type="button" class="brand-lockup" @click="goHome" aria-label="返回首页">
          <span class="brand-lockup__mark">
            <img :src="brandLogo" alt="" aria-hidden="true" class="brand-lockup__icon" />
          </span>
          <span class="brand-lockup__copy">
            <strong>数字丝路</strong>
            <small>Digital Silk Road</small>
          </span>
        </button>

        <nav class="pricing-nav" aria-label="定价页导航">
          <button type="button" class="pricing-nav__item" @click="goHome">
            首页
          </button>
          <button type="button" class="pricing-nav__item" @click="goProductPage">
            产品
          </button>
          <button
            type="button"
            class="pricing-nav__item"
            :class="{ 'is-active': activeSection === 'pricing' }"
            @click="scrollToSection('pricing')"
          >
            定价
          </button>
          <button
            type="button"
            class="pricing-nav__item"
            @click="goAboutPage"
          >
            关于
          </button>
        </nav>

        <div class="pricing-header__actions">
          <button type="button" class="pricing-button pricing-button--primary pricing-button--compact" @click="createProject">
            开始使用
          </button>
        </div>
      </div>
    </header>

    <main class="pricing-main">
      <section ref="pricingSection" data-section="pricing" class="pricing-section pricing-plans">
        <div class="pricing-shell">
          <div class="pricing-heading">
            <h1>选择适合您的方案</h1>
            <p>灵活的定价方案，满足从初创卖家到大型企业的不同需求</p>
            <p>14天免费试用，随时可取消</p>
          </div>

          <div class="pricing-grid" aria-label="数字丝路套餐方案">
            <article
              v-for="plan in plans"
              :key="plan.name"
              class="pricing-card"
              :class="{
                'pricing-card--featured': plan.featured,
                'pricing-card--custom': plan.customPrice,
                'pricing-card--starter': plan.variant === 'starter',
                'pricing-card--enterprise': plan.variant === 'enterprise'
              }"
            >
              <div v-if="plan.featured" class="pricing-card__popular">
                <img :src="popularBadgeStar" alt="" aria-hidden="true" />
                <span>最受欢迎</span>
              </div>

              <div class="pricing-card__icon-wrap" :class="`pricing-card__icon-wrap--${plan.variant}`">
                <img :src="plan.icon" alt="" aria-hidden="true" class="pricing-card__icon" />
              </div>

              <h2>{{ plan.name }}</h2>
              <p class="pricing-card__english">{{ plan.englishName }}</p>
              <p class="pricing-card__description">{{ plan.description }}</p>

              <div class="pricing-card__price">
                <template v-if="plan.customPrice">
                  <span class="pricing-card__price-custom">{{ plan.price }}</span>
                </template>
                <template v-else>
                  <span class="pricing-card__price-value">{{ plan.price }}</span>
                  <span class="pricing-card__price-suffix">{{ plan.priceSuffix }}</span>
                </template>
              </div>

              <button
                type="button"
                class="pricing-card__action"
                :class="{ 'pricing-card__action--featured': plan.featured }"
                @click="handlePlanAction(plan.action)"
              >
                {{ plan.buttonLabel }}
              </button>

              <ul class="pricing-card__features">
                <li v-for="feature in plan.features" :key="feature.label" class="pricing-card__feature">
                  <img :src="feature.icon" alt="" aria-hidden="true" />
                  <span>{{ feature.label }}</span>
                </li>
              </ul>
            </article>
          </div>
        </div>
      </section>
    </main>

    <footer ref="aboutSection" data-section="about" class="pricing-footer">
      <div class="pricing-shell pricing-footer__inner">
        <div class="pricing-footer__top">
          <section class="pricing-footer__brand">
            <button type="button" class="brand-lockup brand-lockup--footer" @click="goHome">
              <span class="brand-lockup__mark">
                <img :src="brandLogo" alt="" aria-hidden="true" class="brand-lockup__icon" />
              </span>
              <span class="brand-lockup__copy">
                <strong>数字丝路</strong>
                <small>Digital Silk Road</small>
              </span>
            </button>

            <p class="pricing-footer__summary">
              跨境电商 AI 一体化平台，提供市场准入合规判断、多语种本地化内容生成、数字人营销表达、智能视频制作的全链路解决方案。
            </p>

          </section>

          <section class="pricing-footer__column">
            <h3>产品</h3>
            <button type="button" class="pricing-footer__link" @click="goProductPage">功能概览</button>
            <button type="button" class="pricing-footer__link" @click="goToWorkbench">工作台</button>
          </section>

          <section class="pricing-footer__column">
            <h3>关于</h3>
            <button type="button" class="pricing-footer__link" @click="goAboutPage">团队介绍</button>
          </section>
        </div>

        <div class="pricing-footer__bottom">
          <span class="pricing-footer__meta-link pricing-footer__meta-link--bottom">
            <img :src="contactEmail" alt="" aria-hidden="true" />
            <span>1549380456@qq.com</span>
          </span>
          <span class="pricing-footer__meta-link pricing-footer__meta-link--bottom">
            <img :src="contactLocation" alt="" aria-hidden="true" />
            <span>泉州 · 中国</span>
          </span>
          <span class="pricing-footer__legal">© 2026 数字丝路 Digital Silk Road. All rights reserved.</span>
          <BeianFooter class="pricing-footer__beian" />
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { BeianFooter } from '@/components/common'
import contactEmail from '@/assets/landing/contact-email.svg'
import contactLocation from '@/assets/landing/contact-location.svg'
import checkIcon from '@/assets/pricing/check.svg'
import enterpriseBadge from '@/assets/pricing/enterprise-badge.svg'
import popularBadgeStar from '@/assets/pricing/popular-badge-star.svg'
import professionalBadge from '@/assets/pricing/professional-badge.svg'
import professionalFeatureIcon from '@/assets/pricing/professional-feature-1.svg'
import starterBadge from '@/assets/pricing/starter-badge.svg'

type SectionKey = 'pricing' | 'about'
type PlanActionKey = 'start' | 'sales'
type PlanVariant = 'starter' | 'professional' | 'enterprise'

interface PricingPlan {
  name: string
  englishName: string
  description: string
  price: string
  priceSuffix?: string
  buttonLabel: string
  variant: PlanVariant
  featured?: boolean
  customPrice?: boolean
  icon: string
  action: PlanActionKey
  features: Array<{
    label: string
    icon: string
  }>
}

const router = useRouter()
const brandLogo = '/logo_circle.png'
const activeSection = ref<SectionKey>('pricing')

const pricingSection = ref<HTMLElement | null>(null)
const aboutSection = ref<HTMLElement | null>(null)

const plans: PricingPlan[] = [
  {
    name: '入门版',
    englishName: 'Starter',
    description: '适合初次出海的中小卖家',
    price: '¥999',
    priceSuffix: '/月',
    buttonLabel: '开始使用',
    variant: 'starter',
    icon: starterBadge,
    action: 'start',
    features: [
      { label: '每月100次合规检测', icon: checkIcon },
      { label: '10种语言内容生成', icon: checkIcon },
      { label: '20个数字人视频', icon: checkIcon },
      { label: '基础数据分析', icon: checkIcon },
      { label: '5GB云存储', icon: checkIcon },
      { label: '邮件支持', icon: checkIcon }
    ]
  },
  {
    name: '专业版',
    englishName: 'Professional',
    description: '适合成长期的跨境电商',
    price: '¥2,999',
    priceSuffix: '/月',
    buttonLabel: '立即升级',
    variant: 'professional',
    featured: true,
    icon: professionalBadge,
    action: 'start',
    features: [
      { label: '每月500次合规检测', icon: professionalFeatureIcon },
      { label: '30种语言内容生成', icon: professionalFeatureIcon },
      { label: '100个数字人视频', icon: professionalFeatureIcon },
      { label: '高级数据分析', icon: professionalFeatureIcon },
      { label: '50GB云存储', icon: professionalFeatureIcon },
      { label: 'A/B测试功能', icon: professionalFeatureIcon },
      { label: '优先技术支持', icon: professionalFeatureIcon },
      { label: '自定义数字人形象', icon: professionalFeatureIcon }
    ]
  },
  {
    name: '企业版',
    englishName: 'Enterprise',
    description: '适合代运营与大型团队',
    price: '定制',
    buttonLabel: '联系销售',
    variant: 'enterprise',
    customPrice: true,
    icon: enterpriseBadge,
    action: 'sales',
    features: [
      { label: '无限次合规检测', icon: checkIcon },
      { label: '全部50+语言', icon: checkIcon },
      { label: '无限数字人视频', icon: checkIcon },
      { label: 'AI优化建议', icon: checkIcon },
      { label: '500GB云存储', icon: checkIcon },
      { label: '团队协作管理', icon: checkIcon },
      { label: '专属客户成功经理', icon: checkIcon },
      { label: 'API接口调用', icon: checkIcon },
      { label: '私有化部署选项', icon: checkIcon }
    ]
  }
] as const

const createProject = () => {
  router.push('/dramas/create')
}

const goHome = () => {
  router.push('/')
}

const goProductPage = () => {
  router.push('/products')
}

const goAboutPage = () => {
  router.push('/about')
}

const goToWorkbench = () => {
  router.push('/dramas')
}

const contactSales = () => {
  if (typeof window !== 'undefined') {
    window.location.href = 'mailto:1549380456@qq.com'
  }
}

const handlePlanAction = (action: PlanActionKey) => {
  if (action === 'sales') {
    contactSales()
    return
  }

  createProject()
}

const getSectionElement = (key: SectionKey) => {
  switch (key) {
    case 'pricing':
      return pricingSection.value
    case 'about':
      return aboutSection.value
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
      rootMargin: '-16% 0px -52% 0px'
    }
  )

  ;[pricingSection.value, aboutSection.value]
    .filter((element): element is HTMLElement => Boolean(element))
    .forEach((element) => observer?.observe(element))
})

onBeforeUnmount(() => {
  observer?.disconnect()
  observer = null
})
</script>

<style scoped>
.pricing-page {
  min-height: var(--app-vh, 100vh);
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
  color: #0a2463;
}

.pricing-shell {
  width: min(1075px, calc(100% - 32px));
  margin: 0 auto;
  padding-inline: 32px;
}

.pricing-header {
  position: sticky;
  top: 0;
  z-index: 40;
  backdrop-filter: blur(18px);
  background: rgba(255, 255, 255, 0.8);
  border-bottom: 1px solid rgba(226, 232, 240, 0.6);
}

.pricing-header__inner {
  min-height: 64px;
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
  width: 40px;
  height: 40px;
  padding: 4px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(226, 232, 240, 0.96);
  box-shadow: 0 10px 15px -12px rgba(15, 23, 42, 0.24);
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

.pricing-nav {
  display: flex;
  align-items: center;
  gap: 32px;
}

.pricing-nav__item,
.pricing-footer__link {
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

.pricing-nav__item {
  position: relative;
  font-weight: 500;
}

.pricing-nav__item::after {
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

.pricing-nav__item:hover,
.pricing-footer__link:hover {
  color: #0a2463;
}

.pricing-nav__item.is-active::after,
.pricing-nav__item:hover::after {
  transform: scaleX(1);
}

.pricing-button {
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

.pricing-button:hover,
.pricing-card__action:hover {
  transform: translateY(-1px);
}

.pricing-button--primary {
  color: #ffffff;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
  box-shadow: 0 16px 32px -22px rgba(249, 115, 22, 0.72);
}

.pricing-button--compact {
  min-height: 44px;
  padding: 10px 20px;
  border-radius: 12px;
}

.pricing-main {
  flex: 1 0 auto;
}

.pricing-section {
  scroll-margin-top: 104px;
}

.pricing-plans {
  padding: 96px 0 clamp(320px, 64vh, 760px);
}

.pricing-heading {
  max-width: 768px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.pricing-heading h1 {
  margin: 0;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 48px;
  line-height: 48px;
  font-weight: 700;
  color: #0a2463;
}

.pricing-heading p {
  margin: 0;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 20px;
  line-height: 32.5px;
  color: #45556c;
}

.pricing-heading p:first-of-type {
  margin-top: 24px;
}

.pricing-grid {
  margin-top: 80px;
  display: grid;
  grid-template-columns: 315.664px 347.6px 315.664px;
  justify-content: center;
  align-items: start;
  gap: 16.036px;
}

.pricing-card {
  position: relative;
  min-height: 800px;
  padding: 32px;
  border: 2px solid #e2e8f0;
  border-radius: 24px;
  background: #ffffff;
  box-shadow: 0 10px 24px -30px rgba(15, 23, 42, 0.24);
}

.pricing-card--featured {
  min-height: 880px;
  padding: 35.2px;
  margin-top: -40px;
  border: 2.2px solid #f97316;
  border-radius: 26.4px;
  box-shadow: 0 27.5px 55px rgba(0, 0, 0, 0.25);
}

.pricing-card__popular {
  position: absolute;
  left: 50%;
  top: -15.4px;
  transform: translateX(-50%);
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-height: 35.2px;
  padding: 6.6px 17.6px;
  border-radius: 999px;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
  box-shadow: 0 12px 24px -18px rgba(249, 115, 22, 0.72);
}

.pricing-card__popular img {
  width: 17.6px;
  height: 17.6px;
}

.pricing-card__popular span {
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 12px;
  line-height: 22px;
  font-weight: 700;
  color: #ffffff;
  white-space: nowrap;
}

.pricing-card__icon-wrap {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 16px;
}

.pricing-card__icon-wrap--starter {
  background: linear-gradient(135deg, #2b7fff 0%, #00b8db 100%);
}

.pricing-card__icon-wrap--professional {
  width: 61.6px;
  height: 61.6px;
  border-radius: 17.6px;
  background: linear-gradient(135deg, #ad46ff 0%, #f6339a 100%);
}

.pricing-card__icon-wrap--enterprise {
  background: linear-gradient(135deg, #ff6900 0%, #fb2c36 100%);
}

.pricing-card__icon {
  width: 28px;
  height: 28px;
}

.pricing-card--featured .pricing-card__icon {
  width: 30.8px;
  height: 30.8px;
}

.pricing-card h2 {
  margin: 22px 0 0;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 24px;
  line-height: 32px;
  font-weight: 700;
  color: #0a2463;
}

.pricing-card--featured h2 {
  margin-top: 24.2px;
  font-size: 26.4px;
  line-height: 35.2px;
}

.pricing-card__english {
  margin: 4px 0 0;
  font-family: 'IBM Plex Sans', sans-serif;
  font-size: 16px;
  line-height: 24px;
  font-weight: 600;
  color: #06b6d4;
}

.pricing-card--featured .pricing-card__english {
  font-size: 17.6px;
  line-height: 26.4px;
}

.pricing-card__description {
  margin: 16px 0 0;
  min-height: 48px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  color: #45556c;
}

.pricing-card--featured .pricing-card__description {
  min-height: 52.8px;
  font-size: 17.6px;
  line-height: 26.4px;
}

.pricing-card__price {
  margin-top: 24px;
  min-height: 48px;
  display: flex;
  align-items: flex-end;
  gap: 4px;
}

.pricing-card--featured .pricing-card__price {
  margin-top: 25.2px;
  min-height: 52.8px;
}

.pricing-card__price-value,
.pricing-card__price-custom {
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 48px;
  line-height: 48px;
  font-weight: 700;
  color: #0a2463;
}

.pricing-card--featured .pricing-card__price-value {
  font-size: 52.8px;
  line-height: 52.8px;
}

.pricing-card__price-suffix {
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  color: #62748e;
}

.pricing-card--featured .pricing-card__price-suffix {
  font-size: 17.6px;
  line-height: 26.4px;
}

.pricing-card__price-custom {
  width: 1em;
  line-height: 0.88;
  white-space: normal;
  word-break: break-all;
}

.pricing-card__action {
  margin-top: 32px;
  width: 100%;
  min-height: 48px;
  border: 0;
  border-radius: 16px;
  background: #f1f5f9;
  color: #0a2463;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  cursor: pointer;
  transition: transform 180ms ease, box-shadow 180ms ease, background-color 180ms ease;
}

.pricing-card__action--featured {
  min-height: 52.8px;
  border-radius: 17.6px;
  color: #ffffff;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
  box-shadow: 0 18px 32px -22px rgba(249, 115, 22, 0.72);
}

.pricing-card__features {
  list-style: none;
  margin-top: 32px;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pricing-card--featured .pricing-card__features {
  margin-top: 35.2px;
  gap: 17.6px;
}

.pricing-card__feature {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  color: #314158;
}

.pricing-card__feature img {
  width: 20px;
  height: 20px;
  margin-top: 2px;
  flex-shrink: 0;
}

.pricing-card--featured .pricing-card__feature {
  gap: 13.2px;
  font-size: 17.6px;
  line-height: 26.4px;
}

.pricing-card--featured .pricing-card__feature img {
  width: 22px;
  height: 22px;
  margin-top: 2.2px;
}

.pricing-footer {
  background: linear-gradient(159deg, #0f172b 0%, #0a2463 50%, #0f172b 100%);
  color: #ffffff;
}

.pricing-footer__inner {
  padding-top: 64px;
  padding-bottom: 28px;
  --text-secondary: #90a1b9;
  --accent: #53eafd;
}

.pricing-footer__top {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 0.6fr) minmax(0, 0.6fr);
  gap: 48px;
}

.brand-lockup--footer .brand-lockup__copy strong,
.brand-lockup--footer .brand-lockup__copy small {
  color: #ffffff;
}

.brand-lockup--footer .brand-lockup__copy small {
  color: #53eafd;
}

.pricing-footer__summary {
  margin: 16px 0 0;
  max-width: 448px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  color: #cad5e2;
}

.pricing-footer__meta {
  margin-top: 24px;
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
}

.pricing-footer__meta-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  color: #90a1b9;
  text-decoration: none;
}

.pricing-footer__meta-link img {
  width: 16px;
  height: 16px;
}

.pricing-footer__column {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pricing-footer__column h3 {
  margin: 0;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 18px;
  line-height: 27px;
  font-weight: 600;
  color: #ffffff;
}

.pricing-footer__column .pricing-footer__link {
  text-align: left;
  color: #cad5e2;
}

.pricing-footer__column .pricing-footer__link:hover {
  color: #ffffff;
}

.pricing-footer__bottom {
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

.pricing-footer__meta-link--bottom,
.pricing-footer__legal,
.pricing-footer__beian {
  flex: 0 0 auto;
}

.pricing-footer__beian {
  justify-content: flex-start;
}

@media (max-width: 1080px) {
  .pricing-grid {
    grid-template-columns: minmax(0, 420px);
    gap: 24px;
  }

  .pricing-card,
  .pricing-card--featured {
    min-height: auto;
    margin-top: 0;
  }
}

@media (max-width: 920px) {
  .pricing-shell {
    width: min(100%, calc(100% - 24px));
    padding-inline: 20px;
  }

  .pricing-nav {
    display: none;
  }

  .pricing-footer__top {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .brand-lockup__copy small {
    display: none;
  }

  .pricing-plans {
    padding: 72px 0 220px;
  }

  .pricing-heading h1 {
    font-size: 40px;
    line-height: 44px;
  }

  .pricing-heading p {
    font-size: 18px;
    line-height: 28px;
  }

  .pricing-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .pricing-card,
  .pricing-card--featured {
    padding: 24px;
    border-radius: 24px;
  }

  .pricing-card--featured {
    border-width: 2px;
  }

  .pricing-card__price-value,
  .pricing-card__price-custom,
  .pricing-card--featured .pricing-card__price-value {
    font-size: 42px;
    line-height: 42px;
  }

  .pricing-footer__top {
    grid-template-columns: 1fr;
  }

  .pricing-footer__meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .pricing-footer__bottom {
    gap: 36px;
  }
}
</style>
