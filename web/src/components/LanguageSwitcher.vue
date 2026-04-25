<template>
  <el-dropdown @command="handleCommand">
    <span class="language-switcher">
      <el-icon><Switch /></el-icon>
      <span class="lang-text">{{ currentLangText }}</span>
    </span>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item command="zh-CN" :disabled="currentLang === 'zh-CN'">
          🇨🇳 {{ t('languageSwitcher.options.zh') }}
        </el-dropdown-item>
        <el-dropdown-item command="en-US" :disabled="currentLang === 'en-US'">
          🇺🇸 {{ t('languageSwitcher.options.en') }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLanguage, toBackendLanguage, toFrontendLanguage } from '@/locales'
import { ElMessage, ElMessageBox } from 'element-plus'
import { settingsAPI } from '@/api/settings'

const { locale, t } = useI18n()

const loading = ref(false)
const currentLang = computed(() => toFrontendLanguage(String(locale.value)))

const currentLangText = computed(() => {
  return currentLang.value === 'zh-CN'
    ? t('languageSwitcher.current.zh')
    : t('languageSwitcher.current.en')
})

const handleCommand = async (lang: string) => {
  if (loading.value) return
  const normalizedLang = toFrontendLanguage(lang)
  if (normalizedLang === currentLang.value) return
  
  // 将 zh-CN/en-US 转换为 zh/en (后端格式)
  const backendLang = toBackendLanguage(normalizedLang)
  const currentBackendLang = toBackendLanguage(currentLang.value)
  
  const confirmMessage = backendLang === 'zh'
    ? t('languageSwitcher.confirm.zh')
    : t('languageSwitcher.confirm.en')
  
  try {
    await ElMessageBox.confirm(
      confirmMessage,
      t('languageSwitcher.confirm.title'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
        dangerouslyUseHTMLString: false
      }
    )

    loading.value = true
    
    // 调用后端API更新语言设置
    const res = await settingsAPI.updateLanguage(backendLang)
    console.log('Backend language updated:', res)
    
    // 更新前端语言
    setLanguage(normalizedLang)
    
    ElMessage.success({
      message: res?.message || (
        backendLang === 'zh'
          ? t('languageSwitcher.messages.switchedToZh')
          : t('languageSwitcher.messages.switchedToEn')
      ),
      duration: 3000
    })
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to switch language:', error)
      
      // 安全获取错误消息
      let errorMessage = t('languageSwitcher.messages.unknownError')
      if (error?.message) {
        errorMessage = error.message
      } else if (error?.response?.data?.error?.message) {
        errorMessage = error.response.data.error.message
      } else if (typeof error === 'string') {
        errorMessage = error
      }
      
      // 双语错误提示
      const errorMsg = currentBackendLang === 'zh'
        ? t('languageSwitcher.messages.switchFailedZh', { error: errorMessage })
        : t('languageSwitcher.messages.switchFailedEn', { error: errorMessage })
      
      ElMessage.error({
        message: errorMsg,
        duration: 5000
      })
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.language-switcher {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  min-height: 48px;
  padding: 0 16px;
  border-radius: 18px;
  border: 1px solid var(--border-primary);
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 16px 36px -28px rgba(34, 62, 109, 0.34);
  transition: all 0.2s ease;
}

.language-switcher:hover {
  background: rgba(255, 255, 255, 0.92);
  border-color: rgba(140, 183, 255, 0.86);
  transform: translateY(-1px);
}

.lang-text {
  font-size: 14px;
  font-weight: 700;
  color: var(--theme-text);
}

.dark .language-switcher {
  background: rgba(16, 28, 50, 0.82);
  box-shadow: 0 16px 36px -28px rgba(0, 0, 0, 0.34);
}

.dark .language-switcher:hover {
  background: rgba(20, 35, 61, 0.96);
}
</style>
