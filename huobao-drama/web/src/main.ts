import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import './assets/styles/element/index.scss'

import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router'
import i18n from './locales'
import './assets/styles/main.css'

// Apply theme before app mounts to prevent flash
// 在应用挂载前应用主题，防止闪烁
const savedTheme = localStorage.getItem('theme')
if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
  document.documentElement.classList.add('dark')
}

// Keep viewport height and virtual keyboard state in CSS variables/classes.
// This improves iOS/Android keyboard overlap handling for fixed/sticky bars.
const setupViewportMetrics = () => {
  const root = document.documentElement

  const updateViewport = () => {
    const viewportHeight = window.visualViewport?.height || window.innerHeight
    root.style.setProperty('--app-vh', `${viewportHeight}px`)

    const keyboardDelta = window.innerHeight - viewportHeight
    root.classList.toggle('keyboard-open', keyboardDelta > 140)
  }

  updateViewport()
  window.addEventListener('resize', updateViewport, { passive: true })
  window.visualViewport?.addEventListener('resize', updateViewport, { passive: true })
  window.visualViewport?.addEventListener('scroll', updateViewport, { passive: true })
}

setupViewportMetrics()

const setupInputFocusAssist = () => {
  const isEditable = (target: EventTarget | null): target is HTMLElement => {
    if (!(target instanceof HTMLElement)) return false
    if (target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement || target instanceof HTMLSelectElement) {
      return true
    }
    return target.isContentEditable
  }

  const bringIntoView = (target: HTMLElement) => {
    if (!document.documentElement.classList.contains('keyboard-open')) return
    const viewportHeight = window.visualViewport?.height || window.innerHeight
    const anchor = (target.closest('.el-form-item') as HTMLElement | null) || target
    const rect = anchor.getBoundingClientRect()
    const visibleBottom = viewportHeight - 88
    const visibleTop = 56

    if (rect.bottom > visibleBottom || rect.top < visibleTop) {
      anchor.scrollIntoView({
        block: 'center',
        inline: 'nearest',
        behavior: 'smooth'
      })
    }
  }

  document.addEventListener(
    'focusin',
    (event) => {
      if (!isEditable(event.target)) return
      window.setTimeout(() => bringIntoView(event.target), 180)
      window.setTimeout(() => bringIntoView(event.target), 420)
    },
    true
  )
}

setupInputFocusAssist()

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(ElementPlus)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.mount('#app')
