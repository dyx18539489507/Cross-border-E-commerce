<template>
  <footer class="marketing-footer">
    <div class="marketing-footer__inner">
      <div class="marketing-footer__top">
        <section class="marketing-footer__brand">
          <button type="button" class="brand-lockup brand-lockup--footer" @click="goHome">
            <span class="brand-lockup__mark">
              <img :src="brandLogo" alt="" aria-hidden="true" class="brand-lockup__icon" />
            </span>
            <span class="brand-lockup__copy">
              <strong>数字丝路</strong>
              <small>Digital Silk Road</small>
            </span>
          </button>

          <p class="marketing-footer__summary">
            跨境电商 AI 一体化平台，提供市场准入合规判断、多语种本地化内容生成、数字人营销表达、智能视频制作的全链路解决方案。
          </p>
        </section>

        <section class="marketing-footer__column">
          <h3>产品</h3>
          <button type="button" class="marketing-footer__link" @click="goProductPage">功能概览</button>
          <button type="button" class="marketing-footer__link" @click="goToWorkbench">工作台</button>
        </section>

        <section class="marketing-footer__column">
          <h3>关于</h3>
          <button type="button" class="marketing-footer__link" @click="goAboutPage">团队介绍</button>
        </section>
      </div>

      <div class="marketing-footer__bottom">
        <span class="marketing-footer__meta-link marketing-footer__meta-link--bottom">
          <img :src="contactEmail" alt="" aria-hidden="true" />
          <span>1549380456@qq.com</span>
        </span>
        <span class="marketing-footer__meta-link marketing-footer__meta-link--bottom">
          <img :src="contactLocation" alt="" aria-hidden="true" />
          <span>泉州 · 中国</span>
        </span>
        <span class="marketing-footer__legal">© 2026 数字丝路 Digital Silk Road. All rights reserved.</span>
        <BeianFooter class="marketing-footer__beian" />
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BeianFooter from './BeianFooter.vue'
import contactEmail from '@/assets/landing/contact-email.svg'
import contactLocation from '@/assets/landing/contact-location.svg'

const router = useRouter()
const route = useRoute()
const brandLogo = '/logo_circle.png'

const scrollToPageTop = (behavior: ScrollBehavior = 'smooth') => {
  if (typeof window === 'undefined') return
  window.scrollTo({ top: 0, left: 0, behavior })
}

const navigateToTop = async (path: string) => {
  if (route.path === path) {
    scrollToPageTop()
    return
  }

  await router.push(path)
  await nextTick()
  scrollToPageTop('auto')
}

const goHome = () => {
  navigateToTop('/')
}

const goProductPage = () => {
  navigateToTop('/products')
}

const goToWorkbench = () => {
  navigateToTop('/dramas')
}

const goAboutPage = () => {
  navigateToTop('/about')
}
</script>

<style scoped>
.marketing-footer {
  background: linear-gradient(159deg, #0f172b 0%, #0a2463 50%, #0f172b 100%);
  color: #ffffff;
}

.marketing-footer__inner {
  width: min(1075px, calc(100% - 32px));
  margin: 0 auto;
  padding: 64px 32px 28px;
  --text-secondary: #90a1b9;
  --accent: #53eafd;
}

.marketing-footer__top {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 0.6fr) minmax(0, 0.6fr);
  gap: 48px;
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
  width: 52px;
  height: 52px;
  padding: 4px;
  border-radius: 16px;
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
  color: #ffffff;
}

.brand-lockup__copy small {
  font-family: 'IBM Plex Sans', 'Segoe UI', sans-serif;
  font-size: 12px;
  line-height: 16px;
  color: #53eafd;
}

.marketing-footer__summary {
  margin-top: 16px;
  max-width: 448px;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  color: #cad5e2;
}

.marketing-footer__column {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.marketing-footer__column h3 {
  margin: 0;
  font-family: 'Urbanist', 'Noto Sans SC', sans-serif;
  font-size: 18px;
  line-height: 27px;
  font-weight: 600;
  color: #ffffff;
}

.marketing-footer__link {
  width: fit-content;
  padding: 0;
  border: 0;
  background: transparent;
  text-align: left;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 16px;
  line-height: 24px;
  color: #cad5e2;
  cursor: pointer;
  transition: color 180ms ease;
}

.marketing-footer__link:hover {
  color: #ffffff;
}

.marketing-footer__bottom {
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

.marketing-footer__meta-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #90a1b9;
  text-decoration: none;
}

.marketing-footer__meta-link img {
  width: 16px;
  height: 16px;
}

.marketing-footer__meta-link--bottom,
.marketing-footer__legal,
.marketing-footer__beian {
  min-width: 0;
  flex: 0 0 auto;
  white-space: nowrap;
}

.marketing-footer__beian {
  justify-content: flex-start;
  column-gap: 12px;
}

.marketing-footer__beian.beian-records {
  min-width: 0;
  column-gap: 12px;
}

.marketing-footer__beian :deep(.beian-record-link) {
  font-size: 11px;
  line-height: 18px;
}

.marketing-footer__beian :deep(.beian-record-icon) {
  width: 16px;
  height: 16px;
}

@media (max-width: 920px) {
  .marketing-footer__inner {
    width: min(100%, calc(100% - 24px));
    padding-inline: 20px;
  }

  .marketing-footer__top {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .brand-lockup__copy small {
    display: none;
  }

  .marketing-footer__top {
    grid-template-columns: 1fr;
  }

  .marketing-footer__bottom {
    gap: 12px;
    overflow-x: auto;
    font-size: 11px;
  }

  .marketing-footer__beian {
    column-gap: 12px;
  }

  .marketing-footer__beian :deep(.beian-record-link) {
    font-size: 10px;
  }
}
</style>
