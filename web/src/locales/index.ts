import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'
import enUS from './en-US'

const FRONTEND_LANGUAGE_MAP: Record<string, 'zh-CN' | 'en-US'> = {
  zh: 'zh-CN',
  'zh-cn': 'zh-CN',
  en: 'en-US',
  'en-us': 'en-US'
}

export const toFrontendLanguage = (lang: string): 'zh-CN' | 'en-US' => {
  const normalized = String(lang || '').trim().toLowerCase()
  return FRONTEND_LANGUAGE_MAP[normalized] || (normalized.startsWith('zh') ? 'zh-CN' : 'en-US')
}

export const toBackendLanguage = (lang: string): 'zh' | 'en' => {
  return toFrontendLanguage(lang) === 'zh-CN' ? 'zh' : 'en'
}

const syncDocumentLanguage = (lang: string) => {
  if (typeof document === 'undefined') return
  const normalized = toFrontendLanguage(lang)
  document.documentElement.lang = normalized
  document.documentElement.dataset.language = normalized
}

// 从 localStorage 获取保存的语言，默认为中文
const getStoredLanguage = (): 'zh-CN' | 'en-US' => {
  const stored = localStorage.getItem('language')
  if (stored) return toFrontendLanguage(stored)
  
  // 自动检测浏览器语言
  const browserLang = navigator.language.toLowerCase()
  if (browserLang.startsWith('zh')) return 'zh-CN'
  return 'en-US'
}

const initialLanguage = getStoredLanguage()

const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: initialLanguage,
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS
  }
})

syncDocumentLanguage(initialLanguage)

export default i18n

// 导出语言切换函数
export const setLanguage = (lang: string) => {
  const normalized = toFrontendLanguage(lang)
  i18n.global.locale.value = normalized as any
  localStorage.setItem('language', normalized)
  syncDocumentLanguage(normalized)
}

export const getCurrentLanguage = () => {
  return toFrontendLanguage(String(i18n.global.locale.value || initialLanguage))
}
