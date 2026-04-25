<template>
  <div class="system-settings">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('settings.systemLanguage') }}</span>
        </div>
      </template>
      
      <el-form label-width="120px">
        <el-form-item :label="$t('settings.currentLanguage')">
          <el-radio-group 
            v-model="currentLanguage" 
            @change="handleLanguageChange"
            :disabled="loading"
          >
            <el-radio label="zh">{{ $t('languageSwitcher.options.zh') }}</el-radio>
            <el-radio label="en">{{ $t('languageSwitcher.options.en') }}</el-radio>
          </el-radio-group>
          <div v-if="loading" style="margin-top: 8px; color: var(--el-color-primary);">
            <el-icon class="is-loading"><Loading /></el-icon>
            {{ $t('languageSwitcher.switching') }}
          </div>
        </el-form-item>
        
        <el-form-item>
          <el-alert
            :title="$t('settings.languageSwitchNotice')"
            type="warning"
            :closable="false"
            show-icon
          >
            <template #default>
              <p>{{ $t('settings.languageSwitchDesc') }}</p>
              <ul>
                <li>{{ $t('settings.languageSwitchItem1') }}</li>
                <li>{{ $t('settings.languageSwitchItem2') }}</li>
                <li>{{ $t('settings.languageSwitchItem3') }}</li>
              </ul>
            </template>
          </el-alert>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { settingsAPI } from '@/api/settings'
import { useI18n } from 'vue-i18n'
import { setLanguage, toBackendLanguage } from '@/locales'

const { t } = useI18n()
const currentLanguage = ref<'zh' | 'en'>('zh')
const loading = ref(false)

const loadCurrentLanguage = async () => {
  try {
    const res = await settingsAPI.getLanguage()
    currentLanguage.value = res?.language as 'zh' | 'en'
    setLanguage(res?.language || 'zh')
    console.log('Current language loaded:', res?.language)
  } catch (error) {
    console.error('Failed to load language:', error)
    ElMessage.error(t('languageSwitcher.messages.loadFailed'))
  }
}

const handleLanguageChange = async (value: 'zh' | 'en') => {
  const backendLang = toBackendLanguage(value)
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
    console.log('Updating language to:', value)
    
    const res = await settingsAPI.updateLanguage(value)
    console.log('Language update response:', res)
    
    setLanguage(value)
    
    ElMessage.success({
      message: res?.message || (
        backendLang === 'zh'
          ? t('languageSwitcher.messages.switchedToZh')
          : t('languageSwitcher.messages.switchedToEn')
      ),
      duration: 3000
    })
  } catch (error: any) {
    console.error('Language update error:', error)
    
    if (error !== 'cancel') {
      let errorMessage = t('languageSwitcher.messages.unknownError')
      if (error?.message) {
        errorMessage = error.message
      } else if (error?.response?.data?.error?.message) {
        errorMessage = error.response.data.error.message
      } else if (typeof error === 'string') {
        errorMessage = error
      }
      
      const errorMsg = currentLanguage.value === 'zh'
        ? t('languageSwitcher.messages.switchFailedZh', { error: errorMessage })
        : t('languageSwitcher.messages.switchFailedEn', { error: errorMessage })
      
      ElMessage.error({
        message: errorMsg,
        duration: 5000
      })
    }
    
    // 恢复原来的选择
    await loadCurrentLanguage()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadCurrentLanguage()
})
</script>

<style scoped>
.system-settings {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

:deep(.el-alert ul) {
  margin-top: 10px;
  padding-left: 20px;
}

:deep(.el-alert li) {
  margin: 5px 0;
}
</style>
