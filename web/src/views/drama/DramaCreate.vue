<template>
  <div class="product-entry-page">
    <header class="product-entry-header">
      <div class="product-entry-header__inner">
        <div class="product-entry-header__left">
          <button type="button" class="brand-link" aria-label="返回首页" @click="router.push('/')">
            <span class="brand-link__mark">
              <img :src="brandLogo" alt="" />
            </span>
            <span class="brand-link__copy">
              <strong>数字丝路</strong>
              <small>Digital Silk Road</small>
            </span>
          </button>

          <nav class="product-entry-nav" aria-label="主导航">
            <button
              v-for="item in navItems"
              :key="item.label"
              type="button"
              class="product-entry-nav__item"
              :class="{ 'product-entry-nav__item--active': item.active }"
              :style="{ width: item.width }"
              :aria-current="item.active ? 'page' : undefined"
              @click="handleNavClick(item.path)"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>

        <div class="product-entry-header__right">
          <div ref="notificationRef" class="notification-center">
            <button
              type="button"
              class="header-icon-button"
              aria-label="通知"
              :aria-expanded="showNotifications"
              aria-haspopup="dialog"
              @click="toggleNotifications"
            >
              <img :src="bellIcon" alt="" />
              <span v-if="unreadNotificationCount" class="header-icon-button__dot">
                {{ unreadNotificationCount }}
              </span>
            </button>

            <section v-if="showNotifications" class="notification-popover" aria-label="消息通知">
              <div class="notification-popover__head">
                <div>
                  <strong>消息通知</strong>
                  <span>{{ unreadNotificationCount }} 条未读消息</span>
                </div>
                <button type="button" class="notification-popover__link" @click="markAllNotificationsRead">
                  全部已读
                </button>
              </div>

              <div class="notification-list">
                <article
                  v-for="notice in notifications"
                  :key="notice.id"
                  class="notification-item"
                  :class="{ 'notification-item--unread': !notice.read }"
                >
                  <span class="notification-item__status" aria-hidden="true"></span>
                  <div class="notification-item__body">
                    <div class="notification-item__title-row">
                      <strong>{{ notice.title }}</strong>
                      <span>{{ notice.time }}</span>
                    </div>
                    <p>{{ notice.content }}</p>
                    <div class="notification-item__actions">
                      <button type="button" @click="openNotification(notice.id)">
                        查看
                      </button>
                      <button type="button" @click="dismissNotification(notice.id)">
                        忽略
                      </button>
                    </div>
                  </div>
                </article>

                <p v-if="notifications.length === 0" class="notification-empty">暂无新的消息</p>
              </div>
            </section>
          </div>
        </div>
      </div>
    </header>

    <main class="product-entry-main">
      <div class="product-entry-shell">
        <div class="product-entry-layout">
          <section class="product-entry-head">
            <h1 class="product-entry-head__title">商品信息录入</h1>
            <p class="product-entry-head__subtitle">填写商品基本信息，开启智能合规检测与内容生成流程</p>
          </section>

          <section class="product-entry-steps" aria-label="步骤进度">
            <div
              v-for="(step, index) in steps"
              :key="step.label"
              class="product-entry-step"
              :class="{ 'product-entry-step--last': index === steps.length - 1 }"
            >
              <div class="product-entry-step__lead">
                <span
                  class="product-entry-step__icon"
                  :class="{ 'product-entry-step__icon--active': step.active }"
                >
                  <img :src="step.icon" alt="" />
                </span>
                <span
                  class="product-entry-step__label"
                  :class="{ 'product-entry-step__label--active': step.active }"
                >
                  {{ step.label }}
                </span>
              </div>

              <span v-if="index !== steps.length - 1" class="product-entry-step__line" aria-hidden="true">
                <span class="product-entry-step__line-fill"></span>
              </span>
            </div>
          </section>

          <section class="product-entry-card">
            <div class="product-entry-card__body">
              <div class="field-block">
                <label class="field-block__label" for="product-name">
                  商品名称
                  <span class="field-block__required">*</span>
                </label>
                <input
                  id="product-name"
                  v-model.trim="form.title"
                  type="text"
                  class="field-block__control"
                  :class="{ 'field-block__control--error': Boolean(errors.title) }"
                  placeholder="例如: 智能手表 Pro Max"
                  maxlength="50"
                  @input="handleTextInput('title')"
                />
                <p v-if="errors.title" class="field-block__error">{{ errors.title }}</p>
              </div>

              <div class="product-entry-grid">
                <div class="field-block">
                  <label class="field-block__label" for="product-category">
                    品类
                    <span class="field-block__required">*</span>
                  </label>
                  <div class="select-shell" :class="{ 'select-shell--error': Boolean(errors.category) }">
                    <select
                      id="product-category"
                      v-model="form.category"
                      class="select-shell__control"
                      :class="{ 'select-shell__control--placeholder': !form.category }"
                      @change="handleCategoryChange"
                    >
                      <option value="" disabled>请选择商品品类</option>
                      <option v-for="option in categoryOptions" :key="option" :value="option">
                        {{ option }}
                      </option>
                    </select>
                    <img :src="chevronDownIcon" alt="" class="select-shell__icon" />
                  </div>
                  <p v-if="errors.category" class="field-block__error">{{ errors.category }}</p>
                </div>

                <div class="field-block">
                  <label class="field-block__label" for="product-brand">品牌</label>
                  <input
                    id="product-brand"
                    v-model.trim="form.brand"
                    type="text"
                    class="field-block__control"
                    placeholder="品牌名称"
                    maxlength="50"
                    @input="persistStepDraft"
                  />
                </div>
              </div>

              <div class="field-block">
                <label class="field-block__label">商品图片</label>
                <input
                  ref="fileInputRef"
                  type="file"
                  class="upload-input"
                  accept="image/png,image/jpeg"
                  @change="handleFileChange"
                />

                <button
                  type="button"
                  class="upload-zone"
                  :class="{
                    'upload-zone--dragging': isDragOver,
                    'upload-zone--filled': Boolean(imagePreviewUrl)
                  }"
                  @click="openFileDialog"
                  @keydown.enter.prevent="openFileDialog"
                  @keydown.space.prevent="openFileDialog"
                  @dragenter.prevent="isDragOver = true"
                  @dragover.prevent="isDragOver = true"
                  @dragleave.prevent="isDragOver = false"
                  @drop.prevent="handleDrop"
                >
                  <template v-if="imagePreviewUrl">
                    <div class="upload-zone__preview">
                      <img :src="imagePreviewUrl" alt="商品预览" class="upload-zone__preview-image" />
                      <div class="upload-zone__preview-copy">
                        <strong>{{ imageName || '已选择商品图片' }}</strong>
                        <span>已完成图片上传，可点击替换或移除当前图片</span>
                        <div class="upload-zone__preview-actions">
                          <span class="upload-zone__preview-link">点击替换</span>
                          <span
                            class="upload-zone__preview-link upload-zone__preview-link--danger"
                            @click.stop="removeImage"
                          >
                            移除图片
                          </span>
                        </div>
                      </div>
                    </div>
                  </template>

                  <template v-else>
                    <div class="upload-zone__empty">
                      <img :src="uploadIcon" alt="" class="upload-zone__icon" />
                      <span class="upload-zone__title">点击上传或拖拽图片到此处</span>
                      <span class="upload-zone__hint">支持 JPG、PNG 格式，最大 5MB</span>
                    </div>
                  </template>
                </button>
              </div>
            </div>

            <div class="product-entry-card__footer">
              <button type="button" class="footer-button footer-button--ghost" disabled>
                上一步
              </button>

              <button type="button" class="footer-button footer-button--primary" @click="handleNextStep">
                <span>下一步</span>
                <img :src="arrowRightIcon" alt="" />
              </button>
            </div>
          </section>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { saveCreateDramaDraft } from '@/utils/createDramaDraft'
import type { CreateDramaRequest } from '@/types/drama'
import arrowRightIcon from '@/assets/figma/product-entry/arrow-right.svg'
import bellIcon from '@/assets/figma/product-entry/bell.svg'
import chevronDownIcon from '@/assets/figma/product-entry/chevron-down.svg'
import stepBasicIcon from '@/assets/figma/product-entry/step-basic.svg'
import stepCompleteIcon from '@/assets/figma/product-entry/step-complete.svg'
import stepDetailIcon from '@/assets/figma/product-entry/step-detail.svg'
import stepMarketIcon from '@/assets/figma/product-entry/step-market.svg'
import uploadIcon from '@/assets/figma/product-entry/upload.svg'

interface ProductEntryDraft {
  title: string
  category: string
  brand: string
}

interface HeaderNotification {
  id: number
  title: string
  content: string
  time: string
  read: boolean
  path?: string
}

const PRODUCT_ENTRY_DRAFT_KEY = 'drama:create:product-entry:basic'
const MAX_UPLOAD_SIZE = 5 * 1024 * 1024
const ACCEPTED_IMAGE_TYPES = new Set(['image/jpeg', 'image/png'])

const router = useRouter()
const brandLogo = '/logo_circle.png'
const fileInputRef = ref<HTMLInputElement | null>(null)
const notificationRef = ref<HTMLElement | null>(null)
const imagePreviewUrl = ref('')
const imageName = ref('')
const isDragOver = ref(false)
const showNotifications = ref(false)

const notifications = ref<HeaderNotification[]>([
  {
    id: 1,
    title: '商品信息待完善',
    content: '补充品牌或商品图后，AI 合规分析会给出更准确的准入风险。',
    time: '刚刚',
    read: false
  },
  {
    id: 2,
    title: '合规检测准备就绪',
    content: '基本信息保存后，可进入目标市场选择并启动合规分析。',
    time: '10 分钟前',
    read: false,
    path: '/compliance'
  },
  {
    id: 3,
    title: '素材规范提醒',
    content: '商品图片建议使用清晰主图，支持 JPG、PNG，单张不超过 5MB。',
    time: '今天',
    read: true
  }
])

const unreadNotificationCount = computed(() => notifications.value.filter((notice) => !notice.read).length)

const form = reactive({
  title: '',
  category: '',
  brand: ''
})

const errors = reactive({
  title: '',
  category: ''
})

const navItems = [
  { label: '工作台', path: '/dramas', active: false, width: '66px' },
  { label: '商品录入', path: '/dramas/create', active: true, width: '80px' },
  { label: '合规分析', path: '/compliance', active: false, width: '80px' },
  { label: '脚本/分镜', path: '/workspace/script', active: false, width: '92px' },
  { label: '内容创作', path: '/workspace/content', active: false, width: '80px' },
  { label: '视频剪辑', path: '/workspace/timeline', active: false, width: '80px' },
  { label: '数据分析', path: '/analytics', active: false, width: '80px' }
] as const

const steps = [
  { label: '基本信息', icon: stepBasicIcon, active: true },
  { label: '目标市场', icon: stepMarketIcon, active: false },
  { label: '商品详情', icon: stepDetailIcon, active: false },
  { label: '完成', icon: stepCompleteIcon, active: false }
] as const

const categoryOptions = [
  '消费电子',
  '家居家电',
  '运动户外',
  '美妆个护',
  '母婴玩具',
  '宠物用品'
]

const getCompatibleDraft = (): CreateDramaRequest => ({
  title: form.title.trim(),
  description: [form.category.trim(), form.brand.trim()].filter(Boolean).join(' / '),
  target_country: [],
  material_composition: '',
  marketing_selling_points: '',
  genre: form.category.trim() || undefined,
  tags: form.brand.trim() || undefined
})

const persistStepDraft = () => {
  if (typeof window === 'undefined') {
    return
  }

  const draft: ProductEntryDraft = {
    title: form.title,
    category: form.category,
    brand: form.brand
  }

  window.sessionStorage.setItem(PRODUCT_ENTRY_DRAFT_KEY, JSON.stringify(draft))
  saveCreateDramaDraft(getCompatibleDraft())
}

const restoreStepDraft = () => {
  if (typeof window === 'undefined') {
    return
  }

  const raw = window.sessionStorage.getItem(PRODUCT_ENTRY_DRAFT_KEY)
  if (!raw) {
    return
  }

  try {
    const draft = JSON.parse(raw) as Partial<ProductEntryDraft>
    form.title = typeof draft.title === 'string' ? draft.title : ''
    form.category = typeof draft.category === 'string' ? draft.category : ''
    form.brand = typeof draft.brand === 'string' ? draft.brand : ''
  } catch {
    window.sessionStorage.removeItem(PRODUCT_ENTRY_DRAFT_KEY)
  }
}

const revokeImagePreview = () => {
  if (!imagePreviewUrl.value) {
    return
  }

  URL.revokeObjectURL(imagePreviewUrl.value)
  imagePreviewUrl.value = ''
}

const clearFieldError = (field: keyof typeof errors) => {
  errors[field] = ''
}

const handleTextInput = (field: keyof typeof errors) => {
  clearFieldError(field)
  persistStepDraft()
}

const handleCategoryChange = () => {
  clearFieldError('category')
  persistStepDraft()
}

const validateVisibleFields = () => {
  let valid = true

  if (!form.title.trim()) {
    errors.title = '请输入商品名称'
    valid = false
  }

  if (!form.category.trim()) {
    errors.category = '请选择商品品类'
    valid = false
  }

  return valid
}

const applySelectedFile = (file: File) => {
  if (!ACCEPTED_IMAGE_TYPES.has(file.type)) {
    ElMessage.error('仅支持 JPG、PNG 格式的图片')
    return
  }

  if (file.size > MAX_UPLOAD_SIZE) {
    ElMessage.error('图片大小不能超过 5MB')
    return
  }

  revokeImagePreview()
  imagePreviewUrl.value = URL.createObjectURL(file)
  imageName.value = file.name
}

const openFileDialog = () => {
  fileInputRef.value?.click()
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  const [file] = target.files || []
  if (!file) {
    return
  }

  applySelectedFile(file)
  target.value = ''
}

const handleDrop = (event: DragEvent) => {
  isDragOver.value = false
  const [file] = Array.from(event.dataTransfer?.files || [])
  if (!file) {
    return
  }

  applySelectedFile(file)
}

const removeImage = () => {
  imageName.value = ''
  revokeImagePreview()
}

const handleNextStep = () => {
  if (!validateVisibleFields()) {
    ElMessage.warning('请先完善必填信息')
    return
  }

  persistStepDraft()
  router.push('/compliance')
}

const handleNavClick = (path: string) => {
  if (!path) {
    return
  }

  router.push(path)
}

const toggleNotifications = () => {
  showNotifications.value = !showNotifications.value
}

const markAllNotificationsRead = () => {
  notifications.value = notifications.value.map((notice) => ({ ...notice, read: true }))
}

const dismissNotification = (id: number) => {
  notifications.value = notifications.value.filter((notice) => notice.id !== id)
}

const openNotification = (id: number) => {
  const notice = notifications.value.find((item) => item.id === id)
  if (!notice) {
    return
  }

  notice.read = true
  showNotifications.value = false

  if (notice.path) {
    router.push(notice.path)
  }
}

const handleDocumentClick = (event: MouseEvent) => {
  if (!showNotifications.value || !notificationRef.value) {
    return
  }

  if (!notificationRef.value.contains(event.target as Node)) {
    showNotifications.value = false
  }
}

onMounted(() => {
  restoreStepDraft()
  document.addEventListener('click', handleDocumentClick)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick)
  revokeImagePreview()
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600;700&family=Noto+Sans+SC:wght@400;500;700&family=Urbanist:wght@700&display=swap');

.product-entry-page {
  min-height: 100vh;
  width: 100%;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
  color: #0a2463;
  overflow-x: hidden;
}

.product-entry-page,
.product-entry-page :is(button, input, select) {
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

.product-entry-header {
  position: fixed;
  inset: 0 0 auto;
  z-index: 30;
  height: 65px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}

.product-entry-header__inner {
  width: min(100%, 1075px);
  height: 64px;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.product-entry-header__left {
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  align-items: center;
  gap: 32px;
}

.brand-link {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  flex-shrink: 0;
}

.brand-link__mark {
  width: 44px;
  height: 44px;
  padding: 4px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(226, 232, 240, 0.92);
  box-shadow: 0 12px 28px -18px rgba(15, 23, 42, 0.34);
}

.brand-link__mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: 999px;
  display: block;
}

.brand-link__copy {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.brand-link__copy strong {
  color: #0a2463;
  font-size: 16px;
  font-weight: 700;
  line-height: 22px;
  white-space: nowrap;
}

.brand-link__copy small {
  color: #62748e;
  font-size: 11px;
  line-height: 14px;
  white-space: nowrap;
}

.product-entry-nav {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 4px;
  overflow-x: auto;
  scrollbar-width: none;
}

.product-entry-nav::-webkit-scrollbar {
  display: none;
}

.product-entry-nav__item {
  height: 32px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: #45556c;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  cursor: pointer;
  transition:
    background-color 180ms ease,
    color 180ms ease,
    transform 180ms ease;
  white-space: nowrap;
}

.product-entry-nav__item:hover {
  color: #0a2463;
  background: rgba(241, 245, 249, 0.92);
}

.product-entry-nav__item--active {
  color: #0a2463;
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.1) 0%, rgba(124, 58, 237, 0.1) 100%);
}

.product-entry-header__right {
  width: 188px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex: 0 0 auto;
}

.notification-center {
  position: relative;
  margin-left: auto;
}

.header-icon-button {
  position: relative;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 12px;
  background: transparent;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 180ms ease;
}

.header-icon-button:hover {
  background: rgba(241, 245, 249, 0.92);
}

.header-icon-button img {
  width: 20px;
  height: 20px;
  display: block;
}

.header-icon-button__dot {
  position: absolute;
  top: 2px;
  left: 22px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 999px;
  background: #f97316;
  border: 2px solid #ffffff;
  color: #ffffff;
  font-size: 10px;
  font-weight: 700;
  line-height: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 16px rgba(249, 115, 22, 0.22);
}

.notification-popover {
  position: absolute;
  top: calc(100% + 14px);
  right: 0;
  width: 340px;
  border: 1px solid rgba(226, 232, 240, 0.96);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.98);
  box-shadow:
    0 24px 48px rgba(15, 23, 42, 0.14),
    0 8px 18px rgba(15, 23, 42, 0.08);
  overflow: hidden;
}

.notification-popover::before {
  content: '';
  position: absolute;
  top: -7px;
  right: 18px;
  width: 14px;
  height: 14px;
  background: #ffffff;
  border-top: 1px solid rgba(226, 232, 240, 0.96);
  border-left: 1px solid rgba(226, 232, 240, 0.96);
  transform: rotate(45deg);
}

.notification-popover__head {
  position: relative;
  z-index: 1;
  padding: 18px 18px 14px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid #eef2f7;
}

.notification-popover__head div {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.notification-popover__head strong {
  color: #0a2463;
  font-size: 16px;
  font-weight: 700;
  line-height: 22px;
}

.notification-popover__head span {
  color: #64748b;
  font-size: 12px;
  line-height: 18px;
}

.notification-popover__link {
  border: none;
  padding: 2px 0;
  background: transparent;
  color: #2563eb;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
  cursor: pointer;
  white-space: nowrap;
}

.notification-popover__link:hover {
  color: #0a2463;
}

.notification-list {
  position: relative;
  z-index: 1;
  max-height: 360px;
  overflow-y: auto;
}

.notification-item {
  display: grid;
  grid-template-columns: 8px 1fr;
  gap: 10px;
  padding: 14px 18px;
  border-bottom: 1px solid #f1f5f9;
  background: #ffffff;
}

.notification-item:last-child {
  border-bottom: none;
}

.notification-item--unread {
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.08) 0%, rgba(124, 58, 237, 0.06) 100%);
}

.notification-item__status {
  width: 8px;
  height: 8px;
  margin-top: 7px;
  border-radius: 999px;
  background: #cbd5e1;
}

.notification-item--unread .notification-item__status {
  background: #f97316;
}

.notification-item__body {
  min-width: 0;
}

.notification-item__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.notification-item__title-row strong {
  min-width: 0;
  color: #0a2463;
  font-size: 14px;
  font-weight: 700;
  line-height: 20px;
}

.notification-item__title-row span {
  color: #90a1b9;
  font-size: 12px;
  line-height: 18px;
  white-space: nowrap;
}

.notification-item p {
  margin: 4px 0 0;
  color: #45556c;
  font-size: 13px;
  line-height: 20px;
}

.notification-item__actions {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.notification-item__actions button {
  border: none;
  padding: 0;
  background: transparent;
  color: #2563eb;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
  cursor: pointer;
}

.notification-item__actions button:last-child {
  color: #64748b;
}

.notification-item__actions button:hover {
  color: #0a2463;
}

.notification-empty {
  margin: 0;
  padding: 28px 18px;
  color: #64748b;
  font-size: 14px;
  line-height: 22px;
  text-align: center;
}


.product-entry-main {
  width: 100%;
}

.product-entry-shell {
  width: min(100%, 1075px);
  margin: 0 auto;
}

.product-entry-layout {
  padding: 104px 25.5px 48px;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
}

.product-entry-head {
  width: 960px;
  margin: 0 auto;
}

.product-entry-head__title {
  margin: 0;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 30px;
  font-weight: 700;
  line-height: 36px;
}

.product-entry-head__subtitle {
  margin: 8px 0 0;
  color: #45556c;
  font-size: 16px;
  font-weight: 400;
  line-height: 24px;
}

.product-entry-steps {
  width: 960px;
  height: 76px;
  margin: 32px auto 0;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.product-entry-step {
  display: flex;
  align-items: center;
  gap: 16px;
  padding-right: 16px;
}

.product-entry-step--last {
  padding-right: 0;
}

.product-entry-step__lead {
  width: 56px;
  height: 76px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
}

.product-entry-step__icon {
  width: 48px;
  height: 48px;
  border-radius: 999px;
  background: #e2e8f0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
}

.product-entry-step__icon--active {
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  box-shadow:
    0 10px 15px 0 rgba(0, 0, 0, 0.1),
    0 4px 6px 0 rgba(0, 0, 0, 0.1);
}

.product-entry-step__icon img {
  width: 24px;
  height: 24px;
  display: block;
}

.product-entry-step__label {
  color: #90a1b9;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  white-space: nowrap;
}

.product-entry-step__label--active {
  color: #0a2463;
}

.product-entry-step__line {
  width: 152px;
  height: 2px;
  background: #e2e8f0;
  flex: 0 0 auto;
  overflow: hidden;
}

.product-entry-step__line-fill {
  display: block;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, #06b6d4 0%, #6382e2 50%, #7c3aed 100%);
}

.product-entry-card {
  width: 960px;
  min-height: 607px;
  margin: 48px auto 0;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #ffffff;
  box-shadow:
    0 10px 15px 0 rgba(0, 0, 0, 0.1),
    0 4px 6px 0 rgba(0, 0, 0, 0.1);
}

.product-entry-card__body {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 33px 33px 0;
}

.field-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-block__label {
  color: #0a2463;
  font-size: 16px;
  font-weight: 500;
  line-height: 24px;
}

.field-block__required {
  color: #fb2c36;
}

.field-block__control,
.select-shell {
  width: 100%;
  height: 50px;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #f8fafc;
}

.field-block__control {
  padding: 12px 16px;
  color: #0f172a;
  font-size: 16px;
  font-weight: 400;
  line-height: normal;
  transition:
    border-color 180ms ease,
    box-shadow 180ms ease,
    background-color 180ms ease;
}

.field-block__control::placeholder {
  color: rgba(15, 23, 42, 0.5);
}

.field-block__control:hover,
.select-shell:hover {
  border-color: #cad5e2;
}

.field-block__control:focus,
.select-shell:focus-within {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.08);
  background: #ffffff;
}

.field-block__control--error,
.select-shell--error {
  border-color: #fb7185;
}

.field-block__error {
  color: #e11d48;
  font-size: 13px;
  line-height: 18px;
}

.product-entry-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 24px;
}

.select-shell {
  position: relative;
  display: flex;
  align-items: center;
  overflow: hidden;
}

.select-shell__control {
  width: 100%;
  height: 100%;
  padding: 12px 48px 12px 16px;
  border: none;
  background: transparent;
  color: #0f172a;
  font-size: 16px;
  font-weight: 400;
  line-height: normal;
  appearance: none;
  cursor: pointer;
}

.select-shell__control:focus {
  outline: none;
}

.select-shell__control--placeholder {
  color: rgba(15, 23, 42, 0.5);
}

.select-shell__icon {
  position: absolute;
  top: 50%;
  right: 16px;
  width: 16px;
  height: 16px;
  transform: translateY(-50%);
  pointer-events: none;
}

.upload-input {
  display: none;
}

.upload-zone {
  min-height: 184px;
  width: 100%;
  border: 2px dashed #cad5e2;
  border-radius: 16px;
  background: transparent;
  padding: 24px 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  cursor: pointer;
  transition:
    border-color 180ms ease,
    background-color 180ms ease,
    transform 180ms ease,
    box-shadow 180ms ease;
}

.upload-zone:hover {
  border-color: #b6c6da;
  background: rgba(248, 250, 252, 0.62);
}

.upload-zone:focus-visible {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.08);
}

.upload-zone--dragging {
  border-color: #7c3aed;
  background: rgba(124, 58, 237, 0.04);
}

.upload-zone--filled {
  border-style: solid;
  background: #f8fafc;
}

.upload-zone__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.upload-zone__icon {
  width: 48px;
  height: 48px;
  display: block;
}

.upload-zone__title {
  margin-top: 16px;
  color: #45556c;
  font-size: 16px;
  font-weight: 400;
  line-height: 24px;
}

.upload-zone__hint {
  margin-top: 4px;
  color: #90a1b9;
  font-size: 14px;
  font-weight: 400;
  line-height: 20px;
}

.upload-zone__preview {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 20px;
}

.upload-zone__preview-image {
  width: 148px;
  height: 104px;
  border-radius: 14px;
  object-fit: cover;
  border: 1px solid #d7e0eb;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.08);
  flex-shrink: 0;
}

.upload-zone__preview-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  text-align: left;
}

.upload-zone__preview-copy strong {
  color: #0a2463;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

.upload-zone__preview-copy span {
  color: #64748b;
  font-size: 14px;
  font-weight: 400;
  line-height: 20px;
}

.upload-zone__preview-actions {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 18px;
}

.upload-zone__preview-link {
  color: #0a2463;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
}

.upload-zone__preview-link--danger {
  color: #dc2626;
}

.product-entry-card__footer {
  margin: 32px 33px 32px;
  padding-top: 33px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.footer-button {
  height: 48px;
  border: none;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
  cursor: pointer;
  transition:
    transform 180ms ease,
    box-shadow 180ms ease,
    opacity 180ms ease;
}

.footer-button img {
  width: 20px;
  height: 20px;
  display: block;
}

.footer-button--ghost {
  width: 96px;
  background: #f1f5f9;
  color: #0a2463;
  opacity: 0.5;
  cursor: not-allowed;
}

.footer-button--primary {
  width: 124px;
  color: #ffffff;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
}

.footer-button--primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 24px rgba(99, 102, 241, 0.24);
}

.footer-button--primary:active {
  transform: translateY(0);
}

@media (max-width: 1120px) {
  .product-entry-header__inner,
  .product-entry-shell {
    width: 100%;
  }

  .product-entry-layout {
    padding-inline: 20px;
  }

  .product-entry-head,
  .product-entry-steps,
  .product-entry-card {
    width: 100%;
  }
}

@media (max-width: 900px) {
  .product-entry-header {
    height: auto;
  }

  .product-entry-header__inner {
    height: auto;
    padding-block: 12px;
    align-items: flex-start;
    flex-direction: column;
  }

  .product-entry-header__left,
  .product-entry-header__right {
    width: 100%;
  }

  .product-entry-header__right {
    justify-content: flex-end;
  }

  .notification-popover {
    right: 0;
  }

  .product-entry-layout {
    padding-top: 140px;
  }

  .product-entry-steps {
    height: auto;
    gap: 20px;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .product-entry-step {
    padding-right: 0;
  }

  .product-entry-step__line {
    display: none;
  }

  .product-entry-grid {
    grid-template-columns: 1fr;
  }

  .upload-zone__preview {
    flex-direction: column;
    align-items: flex-start;
  }
}

@media (max-width: 640px) {
  .product-entry-layout {
    padding: 148px 16px 24px;
  }

  .product-entry-head__title {
    font-size: 28px;
    line-height: 34px;
  }

  .product-entry-card__body {
    padding: 24px 20px 0;
  }

  .product-entry-card__footer {
    margin: 28px 20px 24px;
    padding-top: 24px;
    flex-direction: column-reverse;
    gap: 12px;
  }

  .footer-button {
    width: 100%;
  }

  .upload-zone {
    padding: 24px 20px;
  }

  .notification-center {
    position: static;
  }

  .notification-popover {
    position: fixed;
    top: 78px;
    right: 16px;
    left: 16px;
    width: auto;
  }

  .notification-popover::before {
    right: 22px;
  }
}
</style>
