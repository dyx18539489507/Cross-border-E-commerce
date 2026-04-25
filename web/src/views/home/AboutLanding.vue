<template>
  <div class="about-page">
    <header class="about-header">
      <div class="about-shell about-header__inner">
        <button type="button" class="brand-lockup" @click="goHome" aria-label="返回首页">
          <span class="brand-lockup__mark">
            <img :src="brandLogo" alt="" aria-hidden="true" class="brand-lockup__icon" />
          </span>
          <span class="brand-lockup__copy">
            <strong>数字丝路</strong>
            <small>Digital Silk Road</small>
          </span>
        </button>

        <nav class="about-nav" aria-label="关于页导航">
          <button type="button" class="about-nav__item" @click="goHome">首页</button>
          <button type="button" class="about-nav__item" @click="goProductPage">产品</button>
          <button type="button" class="about-nav__item is-active" @click="scrollToTop">关于</button>
        </nav>

        <div class="about-header__actions">
          <button type="button" class="about-button about-button--primary" @click="createProject">
            开始使用
          </button>
        </div>
      </div>
    </header>

    <main class="about-main">
      <section class="about-values">
        <div class="about-shell">
          <div class="section-heading">
            <h2>我们坚持用清晰的方法做长期的事</h2>
            <p>使命、愿景与价值观，是数字丝路定义产品与服务方式的基础。</p>
          </div>

          <div class="about-values__grid">
            <article v-for="card in valueCards" :key="card.title" class="about-value-card">
              <div class="about-value-card__icon-wrap" :class="card.iconClass">
                <img :src="card.icon" alt="" aria-hidden="true" class="about-value-card__icon" />
              </div>
              <h3>{{ card.title }}</h3>
              <p>{{ card.description }}</p>
            </article>
          </div>
        </div>
      </section>

      <section class="about-story">
        <div class="about-shell about-story__inner">
          <div class="about-story__content">
            <div class="section-heading section-heading--left">
              <h2>我们的故事</h2>
            </div>

            <div class="about-story__paragraphs">
              <p>数字丝路项目诞生于 2024 年，源于团队对跨境电商真实痛点的持续观察与长期服务经验。</p>
              <p>
                <strong>合规问题</strong>
                是中小卖家出海的第一道门槛。不同国家的平台规则、认证标准、风险表达方式差异巨大，人工检索成本高、反馈慢、判断结果不稳定。
              </p>
              <p>
                <strong>本地化内容制作</strong>
                是第二个关键瓶颈。单纯翻译无法建立文化语境，传统内容团队成本高且产能有限，中小团队很难持续规模化产出。
              </p>
              <p>
                <strong>营销视频生产</strong>
                依然存在明显门槛。角色、分镜、图片、口播、剪辑往往割裂在不同软件和团队之间，导致交付周期长、试错成本高。
              </p>
              <p>
                我们相信，AI 不应该只是替代局部动作，而应该真正重构整条内容链路。
                数字丝路因此被设计成一个完整的工作流系统，让卖家能够更快地进入目标市场、理解目标用户，并稳定输出具有转化能力的内容。
              </p>
            </div>
          </div>

          <aside class="about-story__sidebar">
            <article v-for="item in principleCards" :key="item.title" class="principle-card">
              <span class="principle-card__index">{{ item.index }}</span>
              <div class="principle-card__copy">
                <strong>{{ item.title }}</strong>
                <p>{{ item.description }}</p>
              </div>
            </article>
          </aside>
        </div>
      </section>

      <section class="about-capabilities">
        <div class="about-shell">
          <div class="section-heading">
            <h2>平台的核心能力</h2>
            <p>我们把复杂的出海内容生产拆解为清晰、连续、可执行的四个关键模块。</p>
          </div>

          <div class="about-capabilities__grid">
            <article
              v-for="card in capabilityCards"
              :key="card.title"
              class="capability-card"
              :class="card.className"
            >
              <div class="capability-card__top">
                <span class="capability-card__tag">{{ card.tag }}</span>
                <strong>{{ card.metric }}</strong>
              </div>
              <h3>{{ card.title }}</h3>
              <p>{{ card.description }}</p>
            </article>
          </div>
        </div>
      </section>

    </main>

    <MarketingFooter />
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { MarketingFooter } from '@/components/common'
import missionIcon from '@/assets/about/mission-icon.svg'
import valuesIcon from '@/assets/about/values-icon.svg'
import visionIcon from '@/assets/about/vision-icon.svg'

const router = useRouter()
const brandLogo = '/logo_circle.png'

const valueCards = [
  {
    title: '使命',
    description: '用 AI 技术降低跨境电商门槛，让每一个卖家都能更轻松地进入全球市场。',
    icon: missionIcon,
    iconClass: 'about-value-card__icon-wrap--blue'
  },
  {
    title: '愿景',
    description: '成为跨境电商内容生产领域中，最值得长期信赖的 AI 工作流平台。',
    icon: visionIcon,
    iconClass: 'about-value-card__icon-wrap--violet'
  },
  {
    title: '价值观',
    description: '坚持创新、专业、可靠与长期主义，始终把客户成功放在产品演进的中心。',
    icon: valuesIcon,
    iconClass: 'about-value-card__icon-wrap--orange'
  }
]

const principleCards = [
  {
    index: '01',
    title: '从合规开始',
    description: '让项目在内容生产前就具备明确的市场进入判断。'
  },
  {
    index: '02',
    title: '让表达更本地化',
    description: '让脚本、语言和视觉内容更接近目标市场的用户语境。'
  },
  {
    index: '03',
    title: '让交付更连续',
    description: '把图片、数字人、视频与剪辑纳入同一套执行流程。'
  }
]

const capabilityCards = [
  {
    tag: 'Compliance',
    metric: '多市场准入判断',
    title: '市场准入与合规分析',
    description: '针对目标市场输出更清晰的风险等级、调整建议与后续动作，为卖家提供更稳妥的出海起点。',
    className: 'capability-card--blue'
  },
  {
    tag: 'Localization',
    metric: '多语种表达生成',
    title: '本地化脚本与分镜',
    description: '将商品信息快速转译为更符合目标用户语境的文案、镜头结构与营销表达。',
    className: 'capability-card--violet'
  },
  {
    tag: 'Avatar',
    metric: '数字人品牌呈现',
    title: '数字人口播与品牌表达',
    description: '帮助卖家以更低门槛构建稳定的出镜内容，让产品展示更具可信度与可传播性。',
    className: 'capability-card--orange'
  },
  {
    tag: 'Editing',
    metric: '统一时间线输出',
    title: '视频剪辑与成片交付',
    description: '让多模态素材在统一时间线中完成编排、过渡、导出，提升交付速度与内容一致性。',
    className: 'capability-card--cyan'
  }
]

const createProject = () => {
  router.push('/dramas/create')
}

const goHome = () => {
  router.push('/')
}

const goProductPage = () => {
  router.push('/products')
}

const scrollToTop = () => {
  if (typeof window === 'undefined') return
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

</script>

<style scoped>
.about-page {
  min-height: var(--app-vh, 100vh);
  background:
    radial-gradient(circle at top left, rgba(6, 182, 212, 0.09), transparent 28%),
    linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
  color: #0a2463;
}

.about-shell {
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
}

.about-header {
  position: sticky;
  top: 0;
  z-index: 50;
  backdrop-filter: blur(18px);
  background: rgba(255, 255, 255, 0.78);
  border-bottom: 1px solid rgba(226, 232, 240, 0.84);
}

.about-header__inner {
  min-height: 72px;
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
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(226, 232, 240, 0.92);
  box-shadow: 0 12px 28px -16px rgba(15, 23, 42, 0.34);
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
}

.brand-lockup__copy strong {
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 18px;
  line-height: 24px;
  font-weight: 700;
  color: #0a2463;
}

.brand-lockup__copy small {
  font-family: 'IBM Plex Sans', 'Segoe UI', sans-serif;
  font-size: 12px;
  line-height: 16px;
  color: #62748e;
}

.about-nav {
  display: flex;
  align-items: center;
  gap: 32px;
}

.about-nav__item,
.about-footer__link {
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

.about-nav__item {
  position: relative;
  font-weight: 500;
}

.about-nav__item::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: -8px;
  height: 2px;
  border-radius: 999px;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
  transform: scaleX(0);
  transform-origin: left;
  transition: transform 180ms ease;
}

.about-nav__item:hover,
.about-footer__link:hover {
  color: #0a2463;
}

.about-nav__item.is-active::after,
.about-nav__item:hover::after {
  transform: scaleX(1);
}

.about-header__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.about-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 46px;
  padding: 0 22px;
  border-radius: 14px;
  border: 0;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 15px;
  line-height: 22px;
  font-weight: 700;
  cursor: pointer;
  transition: transform 180ms ease, box-shadow 180ms ease, background-color 180ms ease, color 180ms ease;
}

.about-button:hover {
  transform: translateY(-1px);
}

.about-button--primary {
  color: #ffffff;
  background: linear-gradient(135deg, #f97316 0%, #fb923c 100%);
  box-shadow: 0 20px 36px -24px rgba(249, 115, 22, 0.82);
}

.about-button--ghost,
.about-button--secondary {
  color: #0a2463;
  background: rgba(255, 255, 255, 0.86);
  border: 1px solid rgba(203, 213, 225, 0.86);
  box-shadow: 0 16px 32px -26px rgba(15, 23, 42, 0.26);
}

.about-button--large {
  min-height: 56px;
  padding: 0 28px;
  border-radius: 18px;
  font-size: 16px;
}

.about-main {
  padding-bottom: 80px;
}

.section-eyebrow {
  display: inline-flex;
  align-items: center;
  min-height: 32px;
  padding: 0 14px;
  border-radius: 999px;
  border: 1px solid rgba(186, 230, 253, 0.9);
  background: rgba(240, 249, 255, 0.86);
  font-family: 'IBM Plex Sans', 'Segoe UI', sans-serif;
  font-size: 12px;
  line-height: 18px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  font-weight: 700;
  color: #0284c7;
}

.section-eyebrow--hero,
.section-eyebrow--light {
  border-color: rgba(255, 255, 255, 0.22);
  background: rgba(255, 255, 255, 0.12);
  color: #e0f2fe;
}

.section-heading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  text-align: center;
}

.section-heading--left {
  align-items: flex-start;
  text-align: left;
}

.section-heading h2 {
  margin: 0;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 40px;
  line-height: 46px;
  font-weight: 700;
  color: #0a2463;
}

.section-heading p {
  max-width: 720px;
  margin: 0;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 18px;
  line-height: 30px;
  color: #45556c;
}

.section-heading--light h2,
.section-heading--light p {
  color: #ffffff;
}

.about-values,
.about-story,
.about-capabilities {
  padding-top: 96px;
}

.about-values__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 24px;
  margin-top: 36px;
}

.about-value-card {
  min-height: 266px;
  padding: 32px;
  border-radius: 28px;
  border: 1px solid rgba(226, 232, 240, 0.92);
  background: #ffffff;
  box-shadow: 0 24px 48px -36px rgba(15, 23, 42, 0.2);
  transition: transform 180ms ease, box-shadow 180ms ease;
}

.about-value-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 28px 54px -34px rgba(15, 23, 42, 0.28);
}

.about-value-card__icon-wrap {
  width: 64px;
  height: 64px;
  border-radius: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.about-value-card__icon-wrap--blue {
  background: linear-gradient(135deg, #2b7fff 0%, #00b8db 100%);
}

.about-value-card__icon-wrap--violet {
  background: linear-gradient(135deg, #ad46ff 0%, #f6339a 100%);
}

.about-value-card__icon-wrap--orange {
  background: linear-gradient(135deg, #ff6900 0%, #fb2c36 100%);
}

.about-value-card__icon {
  width: 32px;
  height: 32px;
}

.about-value-card h3 {
  margin: 24px 0 0;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 26px;
  line-height: 32px;
  font-weight: 700;
  color: #0a2463;
}

.about-value-card p {
  margin: 16px 0 0;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 28px;
  color: #45556c;
}

.about-story__inner {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 28px;
}

.about-story__content,
.about-story__sidebar {
  border-radius: 32px;
  border: 1px solid rgba(226, 232, 240, 0.92);
  box-shadow: 0 28px 54px -38px rgba(15, 23, 42, 0.18);
}

.about-story__content {
  padding: 42px 44px;
  background:
    radial-gradient(circle at top right, rgba(6, 182, 212, 0.1), transparent 26%),
    linear-gradient(154deg, #f1f5f9 0%, #ffffff 100%);
}

.about-story__paragraphs {
  margin-top: 28px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.about-story__paragraphs p {
  margin: 0;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 28px;
  color: #45556c;
}

.about-story__paragraphs strong {
  color: #0a2463;
  font-weight: 700;
}

.about-story__sidebar {
  background: #ffffff;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.principle-card {
  padding: 18px 18px 20px;
  border-radius: 24px;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
  border: 1px solid rgba(226, 232, 240, 0.92);
}

.principle-card__index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.14), rgba(124, 58, 237, 0.16));
  font-family: 'Urbanist', sans-serif;
  font-size: 18px;
  line-height: 18px;
  font-weight: 700;
  color: #0a2463;
}

.principle-card__copy {
  margin-top: 14px;
}

.principle-card__copy strong {
  display: block;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 20px;
  line-height: 24px;
  font-weight: 700;
  color: #0a2463;
}

.principle-card__copy p {
  margin: 10px 0 0;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 24px;
  color: #45556c;
}

.about-capabilities__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 24px;
  margin-top: 36px;
}

.capability-card {
  min-height: 238px;
  padding: 30px 30px 28px;
  border-radius: 30px;
  color: #ffffff;
  box-shadow: 0 28px 56px -36px rgba(15, 23, 42, 0.34);
  transition: transform 180ms ease, box-shadow 180ms ease;
}

.capability-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 34px 62px -34px rgba(15, 23, 42, 0.38);
}

.capability-card--blue {
  background: linear-gradient(145deg, #0a2463 0%, #175cd3 100%);
}

.capability-card--violet {
  background: linear-gradient(145deg, #4c1d95 0%, #7c3aed 100%);
}

.capability-card--orange {
  background: linear-gradient(145deg, #9a3412 0%, #f97316 100%);
}

.capability-card--cyan {
  background: linear-gradient(145deg, #155e75 0%, #06b6d4 100%);
}

.capability-card__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.capability-card__tag {
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.14);
  border: 1px solid rgba(255, 255, 255, 0.2);
  font-family: 'IBM Plex Sans', sans-serif;
  font-size: 12px;
  line-height: 30px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.capability-card__top strong {
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 22px;
  color: rgba(255, 255, 255, 0.82);
}

.capability-card h3 {
  margin: 42px 0 0;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 30px;
  line-height: 36px;
  font-weight: 700;
}

.capability-card p {
  margin: 14px 0 0;
  max-width: 420px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 28px;
  color: rgba(255, 255, 255, 0.82);
}

.about-footer {
  padding: 88px 0 40px;
  background: #0a1022;
  color: #dbe7f3;
}

.about-footer__inner {
  display: flex;
  flex-direction: column;
  gap: 28px;
  --text-secondary: rgba(219, 231, 243, 0.72);
  --accent: #53eafd;
}

.about-footer__top {
  display: grid;
  grid-template-columns: minmax(0, 1.3fr) 180px 180px;
  gap: 56px;
}

.brand-lockup--footer .brand-lockup__mark {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.12);
  box-shadow: none;
}

.brand-lockup--footer .brand-lockup__copy strong {
  color: #ffffff;
}

.brand-lockup--footer .brand-lockup__copy small {
  color: rgba(255, 255, 255, 0.62);
}

.about-footer__summary {
  max-width: 520px;
  margin: 24px 0 0;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 15px;
  line-height: 26px;
  color: rgba(219, 231, 243, 0.76);
}

.about-footer__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 14px 20px;
  margin-top: 22px;
}

.about-footer__meta-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 22px;
  color: rgba(219, 231, 243, 0.72);
  text-decoration: none;
}

.about-footer__meta-link img {
  width: 16px;
  height: 16px;
}

.about-footer__column {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.about-footer__column h3 {
  margin: 0 0 6px;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 18px;
  line-height: 24px;
  font-weight: 700;
  color: #ffffff;
}

.about-footer__column .about-footer__link {
  width: fit-content;
  color: rgba(219, 231, 243, 0.72);
}

.about-footer__bottom {
  padding-top: 24px;
  border-top: 1px solid rgba(148, 163, 184, 0.18);
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 40px;
  flex-wrap: nowrap;
  overflow-x: auto;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 13px;
  line-height: 20px;
  color: rgba(219, 231, 243, 0.56);
  white-space: nowrap;
}

.about-footer__meta-link--bottom,
.about-footer__legal,
.about-footer__beian {
  flex: 0 0 auto;
}

@media (max-width: 1160px) {
  .about-story__inner {
    grid-template-columns: 1fr;
  }

  .about-values__grid,
  .about-capabilities__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .about-footer__top {
    grid-template-columns: 1fr 1fr;
  }

  .about-footer__brand {
    grid-column: 1 / -1;
  }
}

@media (max-width: 860px) {
  .about-shell {
    width: min(100%, calc(100% - 32px));
  }

  .about-header__inner {
    min-height: 68px;
    flex-wrap: wrap;
    justify-content: center;
    padding: 14px 0;
  }

  .about-nav {
    order: 3;
    width: 100%;
    justify-content: center;
    gap: 22px;
  }

  .about-header__actions {
    display: none;
  }

  .section-heading p,
  .about-story__paragraphs p {
    font-size: 16px;
    line-height: 28px;
  }

  .about-values__grid,
  .about-capabilities__grid {
    grid-template-columns: 1fr;
  }

  .section-heading h2,
  .about-story__content h2 {
    font-size: 32px;
    line-height: 38px;
  }

  .about-story__content,
  .about-story__sidebar {
    padding: 32px 24px;
  }

  .about-footer__top {
    grid-template-columns: 1fr;
  }

  .about-footer__bottom {
    gap: 36px;
  }
}

@media (max-width: 640px) {
  .about-main {
    padding-bottom: 64px;
  }

  .capability-card h3 {
    font-size: 26px;
    line-height: 32px;
  }
}
</style>
