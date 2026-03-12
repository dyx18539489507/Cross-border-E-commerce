<template>
  <!-- Drama List Page - Refactored with modern minimalist design -->
  <!-- 短剧列表页面 - 使用现代简约设计重构 -->
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <!-- App Header / 应用头部 -->
      <AppHeader :fixed="false">
        <template #left>
          <div class="page-title">
            <h1>{{ $t('drama.title') }}</h1>
            <span class="subtitle">{{ $t('drama.totalProjects', { count: total }) }}</span>
          </div>
        </template>
        <template #right>
          <el-button type="primary" @click="handleCreate" class="header-btn primary">
            <el-icon>
              <Plus />
            </el-icon>
            <span class="btn-text">{{ $t('drama.createNew') }}</span>
          </el-button>
        </template>
      </AppHeader>

      <!-- Project Grid / 项目网格 -->
      <div v-loading="loading" class="projects-grid" :class="{ 'is-empty': !loading && dramas.length === 0 }">
        <!-- Empty state / 空状态 -->
        <EmptyState v-if="!loading && dramas.length === 0" :title="$t('drama.empty')" :icon="Film">
          <el-button type="primary" @click="handleCreate">
            <el-icon>
              <Plus />
            </el-icon>
            {{ $t('drama.createNew') }}
          </el-button>
        </EmptyState>

        <!-- Project Cards / 项目卡片列表 -->
        <ProjectCard v-for="drama in dramas" :key="drama.id" :title="drama.title" :description="drama.description"
          :updated-at="drama.updated_at" :episode-count="drama.total_episodes || 0" @click="viewDrama(drama.id)">
          <template #actions>
            <ActionButton :icon="Edit" :tooltip="$t('common.edit')" @click="editDrama(drama.id)" />
            <el-popconfirm :title="$t('drama.deleteConfirm')" :confirm-button-text="$t('common.confirm')"
              :cancel-button-text="$t('common.cancel')" @confirm="deleteDrama(drama.id)">
              <template #reference>
                <el-button :icon="Delete" class="action-button danger" link />
              </template>
            </el-popconfirm>
          </template>
        </ProjectCard>
      </div>

      <!-- Edit Dialog / 编辑对话框 -->
      <el-dialog v-model="editDialogVisible" :title="$t('drama.editProject')" width="640px"
        :close-on-click-modal="false" class="edit-dialog dialog-form-safe">
        <el-form
          ref="editFormRef"
          :model="editForm"
          :rules="editRules"
          label-position="top"
          v-loading="editLoading"
          class="edit-form long-form form-enter-flow"
          @submit.prevent="saveEdit"
          @keydown.enter="handleFormEnterNavigation"
        >
          <el-form-item :label="$t('drama.projectName')" prop="title" required>
            <el-input
              v-model="editForm.title"
              :placeholder="$t('drama.projectNamePlaceholder')"
              size="large"
              maxlength="50"
              show-word-limit
            />
          </el-form-item>
          <el-form-item :label="$t('drama.projectDesc')" prop="description" required>
            <el-input
              v-model="editForm.description"
              type="textarea"
              :rows="4"
              :placeholder="$t('drama.projectDescPlaceholder')"
              maxlength="500"
              show-word-limit
              resize="none"
            />
          </el-form-item>

          <el-form-item :label="$t('drama.targetCountry')" prop="target_country" required>
            <el-select
              v-model="editForm.target_country"
              size="large"
              multiple
              filterable
              :reserve-keyword="false"
              :filter-method="handleEditCountryFilter"
              @change="handleEditCountryChange"
              @visible-change="handleEditCountryVisibleChange"
              :placeholder="$t('drama.targetCountryPlaceholder')"
              :class="['country-select', { 'has-value': (editForm.target_country?.length || 0) > 0 }]"
            >
              <el-option
                v-for="country in filteredEditCountries"
                :key="country.code"
                :label="country.label"
                :value="country.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="$t('drama.materialComposition')" prop="material_composition">
            <el-input
              v-model="editForm.material_composition"
              type="textarea"
              :rows="3"
              :placeholder="$t('drama.materialCompositionPlaceholder')"
              maxlength="200"
              show-word-limit
              resize="none"
            />
          </el-form-item>

          <el-form-item :label="$t('drama.marketingSellingPoints')" prop="marketing_selling_points">
            <el-input
              v-model="editForm.marketing_selling_points"
              type="textarea"
              :rows="3"
              :placeholder="$t('drama.marketingSellingPointsPlaceholder')"
              maxlength="200"
              show-word-limit
              resize="none"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <div class="dialog-footer">
            <el-button @click="editDialogVisible = false" size="large">{{ $t('common.cancel') }}</el-button>
            <el-button type="primary" @click="saveEdit" :loading="editLoading" size="large">
              {{ $t('common.save') }}
            </el-button>
          </div>
        </template>
      </el-dialog>

      <!-- Create Drama Dialog / 创建短剧弹窗 -->
      <CreateDramaDialog v-model="createDialogVisible" @created="loadDramas" />

    </div>

    <div v-if="total === 0" class="page-beian">
      <a
        class="page-beian-link"
        href="https://beian.miit.gov.cn"
        target="_blank"
        rel="noopener noreferrer"
      >
        豫ICP备2026007932号-1
      </a>
    </div>

    <!-- Sticky Pagination / 吸底分页器 -->
    <div v-if="total > 0" class="pagination-sticky">
      <div class="pagination-inner">
        <div class="pagination-info">
          <span class="pagination-total">{{ $t('drama.totalProjects', { count: total }) }}</span>
        </div>
        <a
          class="pagination-beian"
          href="https://beian.miit.gov.cn"
          target="_blank"
          rel="noopener noreferrer"
        >
          豫ICP备2026007932号-1
        </a>
        <div v-if="total > 0" class="pagination-actions">
          <div class="pagination-controls">
            <el-pagination v-model:current-page="queryParams.page" v-model:page-size="queryParams.page_size"
              :total="total" :page-sizes="[12, 24, 36, 48]" :pager-count="5" layout="prev, pager, next"
              @size-change="loadDramas" @current-change="loadDramas" />
          </div>
          <div class="pagination-size">
            <span class="size-label">{{ $t('common.perPage') }}</span>
            <el-select v-model="queryParams.page_size" size="small" class="size-select" @change="loadDramas">
              <el-option :value="12" label="12" />
              <el-option :value="24" label="24" />
              <el-option :value="36" label="36" />
              <el-option :value="48" label="48" />
            </el-select>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  Plus,
  Film,
  Setting,
  Edit,
  View,
  Delete,
  InfoFilled
} from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import type { Drama, DramaListQuery } from '@/types/drama'
import { AppHeader, ProjectCard, ActionButton, CreateDramaDialog, EmptyState } from '@/components/common'
import { handleFormEnterNavigation } from '@/utils/formFocus'
import { ALL_COUNTRIES } from '@/constants/countries'

const router = useRouter()
const loading = ref(false)
const dramas = ref<Drama[]>([])
const total = ref(0)

const queryParams = ref<DramaListQuery>({
  page: 1,
  page_size: 12
})

// Create dialog state / 创建弹窗状态
const createDialogVisible = ref(false)

// Load drama list / 加载短剧列表
const loadDramas = async () => {
  loading.value = true
  try {
    const res = await dramaAPI.list(queryParams.value)
    dramas.value = res.items || []
    total.value = res.pagination?.total || 0
  } catch (error: any) {
    ElMessage.error(error.message || '加载失败')
  } finally {
    loading.value = false
  }
}

// Navigation handlers / 导航处理
const handleCreate = () => createDialogVisible.value = true
const viewDrama = (id: string) => router.push(`/dramas/${id}`)

// Edit dialog state / 编辑对话框状态
const editDialogVisible = ref(false)
const editLoading = ref(false)
const editFormRef = ref<FormInstance>()
const editCountryKeyword = ref('')
const editForm = ref({
  id: '',
  title: '',
  description: '',
  target_country: [] as string[],
  material_composition: '',
  marketing_selling_points: ''
})

const filteredEditCountries = computed(() => {
  const keyword = editCountryKeyword.value.trim().toLowerCase()
  if (!keyword) {
    return ALL_COUNTRIES
  }
  return ALL_COUNTRIES.filter((country) => country.searchText.includes(keyword))
})

const countryCodeSet = new Set(ALL_COUNTRIES.map((item) => item.value))

const normalizeTargetCountries = (value: string | string[] | undefined): string[] => {
  const rawList = Array.isArray(value)
    ? value
    : String(value || '')
      .split(',')
      .map((item) => item.trim())
      .filter((item) => item.length > 0)

  const normalized: string[] = []
  for (const item of rawList) {
    const trimmed = item.trim()
    if (!trimmed) continue

    let code = trimmed.toUpperCase()
    if (!countryCodeSet.has(code)) {
      const bracketMatch = trimmed.match(/\(([A-Za-z]{2})\)/)
      const tailCodeMatch = trimmed.match(/\b([A-Za-z]{2})\b$/)
      code = (bracketMatch?.[1] || tailCodeMatch?.[1] || '').toUpperCase()
    }

    if (countryCodeSet.has(code) && !normalized.includes(code)) {
      normalized.push(code)
    }
  }
  return normalized
}

const handleEditCountryFilter = (keyword: string) => {
  editCountryKeyword.value = keyword
}

const handleEditCountryVisibleChange = (visible: boolean) => {
  if (!visible) {
    editCountryKeyword.value = ''
  }
}

const handleEditCountryChange = () => {
  editCountryKeyword.value = ''
}

const editRules: FormRules = {
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

// Open edit dialog / 打开编辑对话框
const editDrama = async (id: string) => {
  editLoading.value = true
  editDialogVisible.value = true
  try {
    const drama = await dramaAPI.get(id)
    editForm.value = {
      id: drama.id,
      title: drama.title,
      description: drama.description || '',
      target_country: normalizeTargetCountries(drama.target_country as string | string[] | undefined),
      material_composition: drama.material_composition || '',
      marketing_selling_points: drama.marketing_selling_points || ''
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载失败')
    editDialogVisible.value = false
  } finally {
    editLoading.value = false
  }
}

// Save edit changes / 保存编辑更改
const saveEdit = async () => {
  if (!editFormRef.value) return

  const valid = await editFormRef.value.validate().then(() => true).catch(() => false)
  if (!valid) return

  editLoading.value = true
  try {
    await dramaAPI.update(editForm.value.id, {
      title: editForm.value.title.trim(),
      description: editForm.value.description.trim(),
      target_country: editForm.value.target_country.map((item) => item.trim()).filter((item) => item.length > 0),
      material_composition: editForm.value.material_composition.trim(),
      marketing_selling_points: editForm.value.marketing_selling_points.trim()
    })
    ElMessage.success('保存成功')
    editDialogVisible.value = false
    loadDramas()
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    editLoading.value = false
  }
}

// Delete drama / 删除短剧
const deleteDrama = async (id: string) => {
  try {
    await dramaAPI.delete(id)
    ElMessage.success('删除成功')
    loadDramas()
  } catch (error: any) {
    ElMessage.error(error.message || '删除失败')
  }
}

onMounted(() => {
  loadDramas()
})
</script>

<style scoped>
/* ========================================
   Page Layout / 页面布局 - 紧凑边距
   ======================================== */
.page-container {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  /* padding: var(--space-2) var(--space-3); */
  transition: background var(--transition-normal);
}

@media (min-width: 768px) {
  .page-container {
    /* padding: var(--space-3) var(--space-4); */
  }
}

@media (min-width: 1024px) {
  .page-container {
    /* padding: var(--space-4) var(--space-5); */
  }
}

.content-wrapper {
  flex: 1 0 auto;
  margin: 0 auto;
  width: 100%;
}

/* ========================================
   Page Title / 页面标题
   ======================================== */
.page-title {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.page-title h1 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.3;
}

.page-title .subtitle {
  font-size: 0.8125rem;
  color: var(--text-muted);
}

/* ========================================
   Header Buttons / 头部按钮
   ======================================== */
.header-btn {
  border-radius: var(--radius-lg);
  font-weight: 500;
}

.header-btn.primary {
  background: linear-gradient(135deg, var(--accent) 0%, #0284c7 100%);
  border: none;
  box-shadow: 0 4px 14px rgba(14, 165, 233, 0.35);
}

.header-btn.primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(14, 165, 233, 0.45);
}

@media (max-width: 640px) {
  .btn-text {
    display: none;
  }

  .header-btn {
    padding: 0.5rem 0.75rem;
  }
}

/* ========================================
   Projects Grid / 项目网格 - 紧凑间距
   ======================================== */
.projects-grid {
  padding: 12px;
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: var(--space-2);
  margin-bottom: var(--space-4);
  min-height: 300px;
  padding-bottom: var(--space-4);
}

@media (min-width: 480px) {
  .projects-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 640px) {
  .projects-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: var(--space-2);
  }
}

@media (min-width: 900px) {
  .projects-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: var(--space-3);
  }
}

@media (min-width: 1200px) {
  .projects-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}

@media (min-width: 1500px) {
  .projects-grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}

.projects-grid.is-empty {
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ========================================
   Sticky Pagination / 吸底分页器
   ======================================== */
.pagination-sticky {
  position: sticky;
  bottom: 0;
  z-index: 40;
  margin-top: auto;
  width: 100%;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(16px);
  border-top: 1px solid var(--border-primary);
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.05);
  padding-bottom: max(env(safe-area-inset-bottom), 0px);
  transition: transform var(--transition-fast), opacity var(--transition-fast);
}

.dark .pagination-sticky {
  background: rgba(10, 15, 26, 0.9);
  border-top: 1px solid var(--border-primary);
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.3);
}

.pagination-inner {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  margin: 0 auto;
  padding: var(--space-3) var(--space-4);
  gap: var(--space-4);
}

.page-beian {
  flex-shrink: 0;
  display: flex;
  justify-content: center;
  padding: 8px 16px calc(10px + env(safe-area-inset-bottom));
}

.page-beian-link {
  color: var(--text-secondary);
  font-size: 13px;
  line-height: 1.3;
  text-align: center;
  text-decoration: none;
  transition: color var(--transition-fast);
}

.page-beian-link:hover {
  color: var(--accent);
}

@media (min-width: 768px) {
  .pagination-inner {
    padding: var(--space-3) var(--space-6);
  }
}

.pagination-info {
  display: none;
}

@media (min-width: 768px) {
  .pagination-info {
    display: block;
  }
}

.pagination-total {
  font-size: 0.8125rem;
  color: var(--text-muted);
  font-weight: 500;
}

.pagination-beian {
  justify-self: center;
  color: var(--text-secondary);
  font-size: 13px;
  line-height: 1.3;
  text-align: center;
  text-decoration: none;
  white-space: nowrap;
  transition: color var(--transition-fast);
}

.pagination-beian:hover {
  color: var(--accent);
}

.pagination-actions {
  justify-self: end;
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.pagination-controls {
  display: flex;
}

.pagination-size {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.size-label {
  font-size: 0.8125rem;
  color: var(--text-muted);
  display: none;
}

@media (min-width: 768px) {
  .size-label {
    display: block;
  }
}

.size-select {
  width: 4.5rem;
}

.size-select :deep(.el-input__wrapper) {
  height: 2rem;
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

/* ========================================
   Edit Dialog / 编辑对话框
   ======================================== */
.edit-dialog :deep(.el-dialog) {
  border-radius: var(--radius-xl);
}

.edit-dialog :deep(.el-dialog__header) {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-primary);
  margin-right: 0;
}

.edit-dialog :deep(.el-dialog__title) {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.edit-dialog :deep(.el-dialog__body) {
  padding: 1.5rem;
  max-height: 70vh;
  overflow-y: auto;
}

.edit-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.edit-form :deep(.el-input__wrapper),
.edit-form :deep(.el-textarea__inner),
.edit-form :deep(.el-select__wrapper) {
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
  transition: all var(--transition-fast);
}

.edit-form :deep(.el-input__wrapper:hover),
.edit-form :deep(.el-textarea__inner:hover),
.edit-form :deep(.el-select__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--border-secondary) inset;
}

.edit-form :deep(.el-input__wrapper.is-focus),
.edit-form :deep(.el-textarea__inner:focus),
.edit-form :deep(.el-select__wrapper.is-focused) {
  box-shadow: 0 0 0 2px var(--accent) inset;
}

.edit-form :deep(.el-input__inner),
.edit-form :deep(.el-textarea__inner),
.edit-form :deep(.el-select__selected-item) {
  color: var(--text-primary);
}

.edit-form :deep(.el-input__inner::placeholder),
.edit-form :deep(.el-textarea__inner::placeholder) {
  color: var(--text-muted);
}

.edit-form :deep(.el-input__count) {
  color: var(--text-muted);
  background: transparent;
}

.edit-form :deep(.el-form-item.is-error .el-input__wrapper),
.edit-form :deep(.el-form-item.is-error .el-textarea__inner),
.edit-form :deep(.el-form-item.is-error .el-select__wrapper) {
  box-shadow: 0 0 0 1px var(--el-color-danger) inset !important;
}

.edit-form :deep(.el-form-item.is-error .el-input__wrapper.is-focus),
.edit-form :deep(.el-form-item.is-error .el-select__wrapper.is-focused),
.edit-form :deep(.el-form-item.is-error .el-textarea__inner:focus) {
  box-shadow: 0 0 0 2px var(--el-color-danger) inset !important;
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

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

@media (max-width: 768px) {
  .projects-grid {
    padding-bottom: var(--space-4);
  }

  .pagination-inner {
    display: flex;
    justify-content: center;
    padding: var(--space-2) var(--space-3);
    gap: var(--space-2);
    flex-wrap: wrap;
  }

  .pagination-beian {
    width: 100%;
    order: 2;
  }

  .pagination-actions {
    width: 100%;
    justify-content: center;
    gap: var(--space-2);
  }

  .pagination-controls {
    width: 100%;
    justify-content: center;
  }

  .pagination-size {
    display: none;
  }

  .dialog-footer {
    flex-direction: column-reverse;
    width: 100%;
  }

  .dialog-footer .el-button {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .projects-grid {
    padding: 8px;
    padding-bottom: var(--space-3);
  }

  .pagination-inner {
    padding: 8px;
  }

  .pagination-controls :deep(.el-pagination) {
    justify-content: center;
  }
}

/* Delete button style */
.action-button.danger {
  padding: 0.5rem;
  color: var(--text-muted);
}

.action-button.danger:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}
</style>
