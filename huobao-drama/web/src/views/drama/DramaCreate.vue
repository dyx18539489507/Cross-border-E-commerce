<template>
  <!-- Drama Create Page / 创建短剧页面 -->
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <!-- Header / 头部 -->
      <AppHeader :fixed="false" :show-logo="false">
        <template #left>
          <el-button text @click="goBack" class="back-btn">
            <el-icon><ArrowLeft /></el-icon>
            <span>返回</span>
          </el-button>
          <div class="page-title">
            <h1>创建新项目</h1>
            <span class="subtitle">填写基本信息并进行合规校验</span>
          </div>
        </template>
      </AppHeader>

      <!-- Form Card / 表单卡片 -->
      <div class="form-card">
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          class="create-form long-form form-enter-flow"
          @submit.prevent="handleSubmit"
          @keydown.enter="handleFormEnterNavigation"
        >
          <el-form-item label="项目标题" prop="title" required>
            <el-input
              v-model="form.title"
              placeholder="给你的短剧起个名字"
              size="large"
              maxlength="50"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="项目描述" prop="description" required>
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="4"
              placeholder="简要描述你的短剧内容、风格或创意"
              maxlength="500"
              show-word-limit
              resize="none"
            />
          </el-form-item>

          <el-form-item label="目标国家" prop="target_country" required>
            <el-select
              v-model="form.target_country"
              size="large"
              multiple
              filterable
              :reserve-keyword="false"
              :filter-method="handleCountryFilter"
              @change="handleCountryChange"
              @visible-change="handleCountryVisibleChange"
              placeholder="请选择目标国家"
              :class="['country-select', { 'has-value': (form.target_country?.length || 0) > 0 }]"
            >
              <el-option
                v-for="country in filteredCountries"
                :key="country.code"
                :label="country.label"
                :value="country.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item label="材质/成分" prop="material_composition">
            <el-input
              v-model="form.material_composition"
              type="textarea"
              :rows="3"
              placeholder="请输入产品材质、成分或主要原料"
              maxlength="200"
              show-word-limit
              resize="none"
            />
          </el-form-item>

          <el-form-item label="宣传卖点" prop="marketing_selling_points">
            <el-input
              v-model="form.marketing_selling_points"
              type="textarea"
              :rows="3"
              placeholder="请输入宣传卖点（如功能优势、适用场景等）"
              maxlength="200"
              show-word-limit
              resize="none"
            />
          </el-form-item>

          <div class="form-actions">
            <el-button size="large" @click="goBack">取消</el-button>
            <el-button
              type="primary"
              size="large"
              :loading="loading"
              @click="handleSubmit"
            >
              <el-icon v-if="!loading"><ArrowRight /></el-icon>
              下一步
            </el-button>
          </div>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { dramaAPI } from '@/api/drama'
import { ALL_COUNTRIES } from '@/constants/countries'
import type { ComplianceResult, CreateDramaRequest } from '@/types/drama'
import { handleFormEnterNavigation } from '@/utils/formFocus'
import {
  buildCreateDramaPayload,
  getComplianceRiskMeta,
  localizeSuggestedCategories,
  normalizeComplianceResult
} from '@/utils/compliance'
import { AppHeader } from '@/components/common'

const router = useRouter()
const { t } = useI18n()
const formRef = ref<FormInstance>()
const loading = ref(false)
const countryKeyword = ref('')

const form = reactive<CreateDramaRequest>({
  title: '',
  description: '',
  target_country: [],
  material_composition: '',
  marketing_selling_points: ''
})

const filteredCountries = computed(() => {
  const keyword = countryKeyword.value.trim().toLowerCase()
  if (!keyword) {
    return ALL_COUNTRIES
  }
  return ALL_COUNTRIES.filter((country) => country.searchText.includes(keyword))
})

const handleCountryFilter = (keyword: string) => {
  countryKeyword.value = keyword
}

const handleCountryVisibleChange = (visible: boolean) => {
  if (!visible) {
    countryKeyword.value = ''
  }
}

const handleCountryChange = () => {
  countryKeyword.value = ''
}

const rules: FormRules = {
  title: [
    { required: true, message: '请输入项目标题', trigger: 'blur' },
    { min: 1, max: 50, message: '标题长度在 1 到 50 个字符', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入项目描述', trigger: 'blur' },
    { min: 1, max: 500, message: '描述长度在 1 到 500 个字符', trigger: 'blur' }
  ],
  target_country: [
    { type: 'array', required: true, min: 1, message: '请选择目标国家', trigger: 'change' }
  ],
  material_composition: [
    { max: 200, message: '材质/成分长度不能超过 200 个字符', trigger: 'blur' }
  ],
  marketing_selling_points: [
    { max: 200, message: '宣传卖点长度不能超过 200 个字符', trigger: 'blur' }
  ]
}

const escapeHtml = (value: string) =>
  value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')

const renderList = (items: string[]) => {
  if (!items.length) {
    return '<li>无</li>'
  }
  return items.map((item) => `<li>${escapeHtml(item)}</li>`).join('')
}

const showComplianceReport = async (compliance: ComplianceResult) => {
  const riskMeta = getComplianceRiskMeta(compliance)
  const levelLabel = compliance.level_label || riskMeta.text
  const riskLabel = `${levelLabel}（${riskMeta.range}）`
  const localizedCategories = localizeSuggestedCategories(compliance.suggested_categories || [])

  const html = `
    <div style="line-height: 1.6;">
      <p><strong>风险评分：</strong><span style="color:${riskMeta.color};font-weight:700;">${compliance.score}</span></p>
      <p><strong>风险等级：</strong><span style="color:${riskMeta.color};font-weight:700;">${riskLabel}</span></p>
      <p><strong>评估结论：</strong>${escapeHtml(compliance.summary || '无')}</p>
      <p><strong>不合规点：</strong></p>
      <ul>${renderList(compliance.non_compliance_points || [])}</ul>
      <p><strong>整改建议：</strong></p>
      <ul>${renderList(compliance.rectification_suggestions || [])}</ul>
      <p><strong>建议类目：</strong></p>
      <ul>${renderList(localizedCategories)}</ul>
    </div>
  `

  await ElMessageBox.alert(html, t('drama.complianceReportTitle'), {
    confirmButtonText: t('common.confirm'),
    dangerouslyUseHTMLString: true
  })
}

const tryShowComplianceFromError = async (error: any): Promise<boolean> => {
  const compliance = normalizeComplianceResult(error?.details?.compliance)
  if (!compliance) {
    return false
  }
  await showComplianceReport(compliance)
  return true
}

// Submit form / 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  const valid = await formRef.value.validate().then(() => true).catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const payload = buildCreateDramaPayload(form)
    const result = await dramaAPI.create(payload)
    const dramaId = String(result.drama.id)
    await showComplianceReport(result.compliance)

    ElMessage.success('创建成功')
    router.push(`/dramas/${dramaId}`)
  } catch (error: any) {
    const hasCompliance = await tryShowComplianceFromError(error)
    if (hasCompliance) {
      ElMessage.warning(error.message || '风险过高，请先整改后再提交')
      return
    }
    ElMessage.error(error.message || '创建失败')
  } finally {
    loading.value = false
  }
}

// Go back / 返回上一页
const goBack = () => {
  router.back()
}

</script>

<style scoped>
/* ========================================
   Page Layout / 页面布局 - 紧凑边距
   ======================================== */
.page-container {
  min-height: 100vh;
  background-color: var(--bg-primary);
  padding: var(--space-2) var(--space-3);
  transition: background-color var(--transition-normal);
}

@media (min-width: 768px) {
  .page-container {
    padding: var(--space-3) var(--space-4);
  }
}

.content-wrapper {
  max-width: 640px;
  margin: 0 auto;
}

/* ========================================
   Form Card / 表单卡片
   ======================================== */
.form-card {
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-card);
}

/* ========================================
   Form Styles / 表单样式 - 紧凑内边距
   ======================================== */
.create-form {
  padding: var(--space-4);
}

.create-form :deep(.el-form-item) {
  margin-bottom: var(--space-4);
}

.country-select {
  width: 100%;
}

.country-select :deep(.el-select__placeholder) {
  color: #a8b5c6;
}

.country-select.has-value :deep(.el-select__placeholder) {
  color: var(--text-primary);
}

/* ========================================
   Form Actions / 表单操作区
   ======================================== */
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  padding-top: var(--space-4);
  border-top: 1px solid var(--border-primary);
  margin-top: var(--space-2);
}

.form-actions .el-button {
  min-width: 100px;
}

@media (max-width: 768px) {
  .page-container {
    padding: var(--space-2);
  }

  .create-form {
    padding: var(--space-3);
  }

  .form-actions {
    position: sticky;
    bottom: max(env(safe-area-inset-bottom), 0px);
    z-index: 5;
    flex-direction: column-reverse;
    align-items: stretch;
    gap: var(--space-2);
    background: var(--bg-card);
    margin-top: var(--space-4);
    padding: var(--space-3) 0 max(var(--space-2), env(safe-area-inset-bottom)) 0;
  }

  .form-actions .el-button {
    width: 100%;
    min-width: 0;
  }
}
</style>
