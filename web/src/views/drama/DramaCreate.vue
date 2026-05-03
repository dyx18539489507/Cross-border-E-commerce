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
                  <div ref="categoryCascaderRef" class="category-cascader">
                    <button
                      id="product-category"
                      type="button"
                      class="category-cascader__trigger"
                      :class="{
                        'category-cascader__trigger--placeholder': !resolvedCategoryLabel,
                        'category-cascader__trigger--error': Boolean(errors.category),
                        'category-cascader__trigger--open': isCategoryCascaderOpen
                      }"
                      aria-haspopup="listbox"
                      :aria-expanded="isCategoryCascaderOpen"
                      @click="toggleCategoryCascader"
                    >
                      <span>{{ resolvedCategoryLabel || '请选择商品品类' }}</span>
                      <img :src="chevronDownIcon" alt="" class="category-cascader__icon" />
                    </button>

                    <div v-if="isCategoryCascaderOpen" class="category-cascader__panel">
                      <div class="category-cascader__search">
                        <input
                          ref="categorySearchInputRef"
                          v-model="categorySearchKeyword"
                          type="search"
                          class="category-cascader__search-input"
                          placeholder="搜索一级或二级品类"
                          aria-label="搜索商品品类"
                          @click.stop
                          @keydown.stop
                        />
                      </div>

                      <div v-if="normalizedCategorySearch" class="category-cascader__results" role="listbox">
                        <button
                          v-for="result in categorySearchResults"
                          :key="`${result.primary}/${result.secondary}`"
                          type="button"
                          class="category-cascader__result"
                          :class="{
                            'category-cascader__result--active':
                              form.categoryPrimary === result.primary && form.categorySecondary === result.secondary
                          }"
                          @click="selectCategoryResult(result)"
                        >
                          <span class="category-cascader__result-primary">{{ result.primary }}</span>
                          <strong class="category-cascader__result-secondary">{{ result.secondary }}</strong>
                        </button>

                        <p v-if="categorySearchResults.length === 0" class="category-cascader__empty">
                          未找到匹配品类
                        </p>
                      </div>

                      <div v-else class="category-cascader__columns">
                        <div class="category-cascader__column category-cascader__column--primary" role="listbox">
                          <button
                            v-for="group in categoryGroups"
                            :key="group.label"
                            type="button"
                            class="category-cascader__option"
                            :class="{ 'category-cascader__option--active': form.categoryPrimary === group.label }"
                            @click="selectCategoryPrimary(group.label)"
                          >
                            {{ group.label }}
                          </button>
                        </div>

                        <div class="category-cascader__column category-cascader__column--secondary" role="listbox">
                          <p v-if="!form.categoryPrimary" class="category-cascader__empty">请先选择一级品类</p>
                          <template v-else>
                            <button
                              v-for="option in subcategoryOptions"
                              :key="option"
                              type="button"
                              class="category-cascader__option"
                              :class="{ 'category-cascader__option--active': form.categorySecondary === option }"
                              @click="selectCategorySecondary(option)"
                            >
                              {{ option }}
                            </button>
                          </template>
                        </div>
                      </div>
                    </div>
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
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
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
  categoryPrimary?: string
  categorySecondary?: string
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

interface CategorySearchResult {
  primary: string
  secondary: string
}

const PRODUCT_ENTRY_DRAFT_KEY = 'drama:create:product-entry:basic'
const MAX_UPLOAD_SIZE = 5 * 1024 * 1024
const ACCEPTED_IMAGE_TYPES = new Set(['image/jpeg', 'image/png'])

const router = useRouter()
const brandLogo = '/logo_circle.png'
const fileInputRef = ref<HTMLInputElement | null>(null)
const categorySearchInputRef = ref<HTMLInputElement | null>(null)
const notificationRef = ref<HTMLElement | null>(null)
const categoryCascaderRef = ref<HTMLElement | null>(null)
const imagePreviewUrl = ref('')
const imageName = ref('')
const isDragOver = ref(false)
const showNotifications = ref(false)
const isCategoryCascaderOpen = ref(false)
const categorySearchKeyword = ref('')

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
  categoryPrimary: '',
  categorySecondary: '',
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

const categoryGroups = [
  {
    label: '消费电子',
    children: ['智能穿戴', '智能手表', '手机配件', '电脑周边', '音频设备', '摄影摄像', '游戏外设', '充电储能']
  },
  {
    label: '家用电器',
    children: ['厨房小家电', '空气炸锅', '咖啡设备', '清洁电器', '个护电器', '环境电器', '食品接触小家电']
  },
  {
    label: '智能家居',
    children: ['智能安防', '智能照明', '智能插座', '智能门锁', '传感器', '智能窗帘']
  },
  {
    label: '家居生活',
    children: ['厨房餐厨', '收纳整理', '家纺布艺', '灯具照明', '卫浴用品', '家装五金']
  },
  {
    label: '美妆个护',
    children: ['护肤', '彩妆', '美发造型', '身体护理', '口腔护理', '美容仪器', '香氛']
  },
  {
    label: '服饰配件',
    children: ['女装', '男装', '鞋靴', '箱包', '饰品配件', '运动服饰', '内衣家居服']
  },
  {
    label: '运动户外',
    children: ['健身器材', '户外露营', '骑行装备', '球类运动', '瑜伽训练', '运动护具']
  },
  {
    label: '母婴玩具',
    children: ['婴童用品', '益智玩具', '毛绒玩具', '童车童床', '喂养用品', '儿童服饰']
  },
  {
    label: '玩具模型',
    children: ['积木拼插', '模型手办', '遥控玩具', '桌游卡牌', '科学玩具']
  },
  {
    label: '宠物用品',
    children: ['宠物喂养', '宠物清洁', '宠物玩具', '宠物出行']
  },
  {
    label: '食品饮料',
    children: ['休闲零食', '冲调饮品', '健康食品', '烘焙原料', '调味品', '即食食品']
  },
  {
    label: '汽摩配件',
    children: ['车载电子', '汽车内饰', '清洁养护', '骑行配件', '维修工具', '安全应急']
  },
  {
    label: '办公文具',
    children: ['办公设备', '书写工具', '纸品本册', '桌面收纳', '会议用品', '打印耗材']
  },
  {
    label: '工业工具',
    children: ['手动工具', '电动工具', '测量仪器', '劳保用品', '包装耗材', '五金配件']
  },
  {
    label: '电子元器件',
    children: ['连接器', '传感元件', '开发板', '电源模块', '线缆线束', '显示模组']
  },
  {
    label: '商业设备',
    children: ['收银设备', '条码扫描', '展示货架', '餐饮设备', '仓储设备', '广告标识']
  },
  {
    label: '包装印刷',
    children: ['纸箱纸盒', '快递包装', '食品包装', '礼品包装', '标签贴纸', '包装袋']
  },
  {
    label: '家装建材',
    children: ['墙地面材料', '门窗五金', '厨卫五金', '电工电料', '装饰材料', '防水密封']
  },
  {
    label: '珠宝饰品',
    children: ['时尚首饰', '手表', '眼镜配件', '发饰']
  },
  {
    label: '箱包旅行',
    children: ['旅行箱', '双肩包', '收纳包', '钱包卡包', '旅行配件', '防盗包']
  },
  {
    label: '园艺工具',
    children: ['园艺工具', '户外照明', '花盆容器', '灌溉设备', '草坪维护']
  },
  {
    label: '医疗健康',
    children: ['家用检测', '康复护理', '按摩理疗', '健康监测', '辅助器具', '消毒防护']
  },
  {
    label: '乐器音响',
    children: ['乐器配件', '录音设备', '扬声器', '调音设备']
  },
  {
    label: '图书教育',
    children: ['教辅材料', '早教用品', '教学教具', '科学实验']
  },
  {
    label: '鞋服箱配',
    children: ['功能鞋靴', '户外鞋服', '帽袜围巾', '服装辅料']
  },
  {
    label: '节庆礼品',
    children: ['节日装饰', '派对用品', '创意礼品', '贺卡包装', '婚庆用品', '定制礼品']
  },
  {
    label: '艺术手工',
    children: ['绘画用品', '针织缝纫', '陶艺材料', '纸艺材料', '手账文创']
  },
  {
    label: '安防监控',
    children: ['监控摄像', '门禁考勤', '报警器', '可视门铃', '消防用品', '安全标识']
  },
  {
    label: '绿色能源',
    children: ['太阳能设备', '储能电源', '节能照明', '充电桩配件', '户外电源', '电池配件']
  },
  {
    label: '农牧用品',
    children: ['农用工具', '灌溉配件', '养殖设备', '种植耗材', '温室配件']
  }
]

const selectedCategoryGroup = computed(() => {
  return categoryGroups.find((group) => group.label === form.categoryPrimary)
})

const subcategoryOptions = computed(() => selectedCategoryGroup.value?.children || [])

const normalizeCategorySearchText = (value: string) => value.trim().toLocaleLowerCase('zh-CN')

const normalizedCategorySearch = computed(() => normalizeCategorySearchText(categorySearchKeyword.value))

const categorySearchResults = computed<CategorySearchResult[]>(() => {
  const keyword = normalizedCategorySearch.value

  if (!keyword) {
    return []
  }

  return categoryGroups.flatMap((group) => {
    const primaryMatches = normalizeCategorySearchText(group.label).includes(keyword)

    return group.children
      .filter((secondary) => {
        const fullLabel = normalizeCategorySearchText(`${group.label} ${secondary}`)
        return primaryMatches || fullLabel.includes(keyword)
      })
      .map((secondary) => ({
        primary: group.label,
        secondary
      }))
  })
})

const getResolvedCategory = () => {
  return [form.categoryPrimary, form.categorySecondary].filter(Boolean).join(' / ')
}

const resolvedCategoryLabel = computed(() => getResolvedCategory())

const getCompatibleDraft = (): CreateDramaRequest => ({
  title: form.title.trim(),
  description: [getResolvedCategory(), form.brand.trim()].filter(Boolean).join(' / '),
  target_country: [],
  material_composition: '',
  marketing_selling_points: '',
  genre: getResolvedCategory() || undefined,
  tags: form.brand.trim() || undefined
})

const persistStepDraft = () => {
  if (typeof window === 'undefined') {
    return
  }

  const draft: ProductEntryDraft = {
    title: form.title,
    category: getResolvedCategory(),
    categoryPrimary: form.categoryPrimary,
    categorySecondary: form.categorySecondary,
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
    const savedCategory = typeof draft.category === 'string' ? draft.category : ''
    const [savedPrimary = '', savedSecondary = ''] = savedCategory.split('/').map((item) => item.trim())
    form.categoryPrimary = typeof draft.categoryPrimary === 'string' ? draft.categoryPrimary : savedPrimary
    form.categorySecondary = typeof draft.categorySecondary === 'string' ? draft.categorySecondary : savedSecondary
    form.category = getResolvedCategory()
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

const focusCategorySearchInput = () => {
  void nextTick(() => {
    categorySearchInputRef.value?.focus()
  })
}

const toggleCategoryCascader = () => {
  isCategoryCascaderOpen.value = !isCategoryCascaderOpen.value

  if (isCategoryCascaderOpen.value) {
    categorySearchKeyword.value = ''
    focusCategorySearchInput()
  }
}

const selectCategoryPrimary = (label: string) => {
  if (form.categoryPrimary === label) {
    return
  }

  form.categoryPrimary = label
  form.categorySecondary = ''
  form.category = getResolvedCategory()
  clearFieldError('category')
  persistStepDraft()
}

const selectCategorySecondary = (option: string) => {
  form.categorySecondary = option
  form.category = getResolvedCategory()
  clearFieldError('category')
  persistStepDraft()
  categorySearchKeyword.value = ''
  isCategoryCascaderOpen.value = false
}

const selectCategoryResult = (result: CategorySearchResult) => {
  form.categoryPrimary = result.primary
  selectCategorySecondary(result.secondary)
}

const validateVisibleFields = () => {
  let valid = true

  if (!form.title.trim()) {
    errors.title = '请输入商品名称'
    valid = false
  }

  if (!form.categoryPrimary.trim() || !form.categorySecondary.trim()) {
    errors.category = '请选择完整的一级和二级商品品类'
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
  const target = event.target as Node

  if (showNotifications.value && notificationRef.value && !notificationRef.value.contains(target)) {
    showNotifications.value = false
  }

  if (isCategoryCascaderOpen.value && categoryCascaderRef.value && !categoryCascaderRef.value.contains(target)) {
    isCategoryCascaderOpen.value = false
    categorySearchKeyword.value = ''
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
  padding: 96px 25.5px 28px;
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
  margin: 28px auto 0;
  display: flex;
  align-items: flex-start;
}

.product-entry-step {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
  flex: 1 1 0;
}

.product-entry-step--last {
  flex: 0 0 auto;
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
  width: auto;
  min-width: 0;
  height: 2px;
  background: #e2e8f0;
  flex: 1 1 auto;
  margin-right: 16px;
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
  margin: 36px auto 0;
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
  gap: 20px;
  padding: 28px 33px 0;
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
.category-cascader__trigger {
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
.category-cascader__trigger:hover {
  border-color: #cad5e2;
}

.field-block__control:focus,
.category-cascader__trigger:focus-visible,
.category-cascader__trigger--open {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.08);
  background: #ffffff;
}

.field-block__control--error,
.category-cascader__trigger--error {
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

.category-cascader {
  position: relative;
  width: 100%;
}

.category-cascader__trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 48px 12px 16px;
  color: #0f172a;
  font-size: 16px;
  font-weight: 400;
  line-height: normal;
  text-align: left;
  cursor: pointer;
  transition:
    border-color 180ms ease,
    box-shadow 180ms ease,
    background-color 180ms ease;
}

.category-cascader__trigger span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.category-cascader__trigger--placeholder {
  color: rgba(15, 23, 42, 0.5);
}

.category-cascader__icon {
  position: absolute;
  top: 50%;
  right: 16px;
  width: 16px;
  height: 16px;
  transform: translateY(-50%);
  pointer-events: none;
}

.category-cascader__panel {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  z-index: 20;
  display: flex;
  flex-direction: column;
  border: 1px solid #dbe4ef;
  border-radius: 16px;
  background: #ffffff;
  box-shadow:
    0 16px 36px rgba(15, 23, 42, 0.12),
    0 4px 12px rgba(15, 23, 42, 0.08);
  overflow: hidden;
}

.category-cascader__search {
  padding: 10px;
  border-bottom: 1px solid #e2e8f0;
  background: #ffffff;
}

.category-cascader__search-input {
  width: 100%;
  height: 38px;
  border: 1px solid #dbe4ef;
  border-radius: 12px;
  background: #f8fafc;
  padding: 8px 12px;
  color: #0f172a;
  font-size: 14px;
  line-height: 20px;
  transition:
    border-color 180ms ease,
    box-shadow 180ms ease,
    background-color 180ms ease;
}

.category-cascader__search-input::placeholder {
  color: #90a1b9;
}

.category-cascader__search-input:focus {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.08);
  background: #ffffff;
}

.category-cascader__columns {
  min-height: 0;
  display: grid;
  grid-template-columns: 168px minmax(0, 1fr);
  flex: 1 1 auto;
  align-items: stretch;
}

.category-cascader__column {
  padding: 8px;
}

.category-cascader__column--primary {
  max-height: 286px;
  overflow-y: auto;
}

.category-cascader__column--primary {
  border-right: 1px solid #e2e8f0;
  background: #f8fafc;
}

.category-cascader__column--secondary {
  overflow: visible;
}

.category-cascader__option {
  width: 100%;
  min-height: 36px;
  border: none;
  border-radius: 10px;
  background: transparent;
  padding: 8px 10px;
  color: #314158;
  font-size: 14px;
  font-weight: 400;
  line-height: 20px;
  text-align: left;
  cursor: pointer;
  transition:
    background-color 160ms ease,
    color 160ms ease;
}

.category-cascader__option:hover,
.category-cascader__option--active {
  background: rgba(124, 58, 237, 0.08);
  color: #0a2463;
  font-weight: 600;
}

.category-cascader__results {
  max-height: 286px;
  padding: 8px;
  overflow-y: auto;
}

.category-cascader__result {
  width: 100%;
  min-height: 46px;
  border: none;
  border-radius: 12px;
  background: transparent;
  padding: 7px 10px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  gap: 2px;
  text-align: left;
  cursor: pointer;
  transition:
    background-color 160ms ease,
    color 160ms ease;
}

.category-cascader__result:hover,
.category-cascader__result--active {
  background: rgba(124, 58, 237, 0.08);
}

.category-cascader__result-primary {
  color: #64748b;
  font-size: 12px;
  line-height: 16px;
}

.category-cascader__result-secondary {
  color: #0f172a;
  font-size: 14px;
  font-weight: 600;
  line-height: 20px;
}

.category-cascader__empty {
  margin: 0;
  padding: 12px 10px;
  color: #90a1b9;
  font-size: 14px;
  line-height: 20px;
}

.upload-input {
  display: none;
}

.upload-zone {
  min-height: 164px;
  width: 100%;
  border: 2px dashed #cad5e2;
  border-radius: 16px;
  background: transparent;
  padding: 20px 32px;
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
  margin: 24px 33px 28px;
  padding-top: 24px;
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
    display: grid;
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

  .category-cascader__panel {
    position: fixed;
    top: auto;
    right: 16px;
    bottom: calc(16px + env(safe-area-inset-bottom));
    left: 16px;
    max-height: min(78vh, 560px);
    border-radius: 20px;
    overflow: hidden;
  }

  .category-cascader__search {
    position: sticky;
    top: 0;
    z-index: 1;
    padding: 12px;
  }

  .category-cascader__search-input {
    height: 44px;
    font-size: 16px;
  }

  .category-cascader__columns {
    grid-template-columns: minmax(118px, 0.42fr) minmax(0, 1fr);
    max-height: calc(min(78vh, 560px) - 69px);
    overflow: hidden;
  }

  .category-cascader__column {
    max-height: none;
    padding: 10px;
  }

  .category-cascader__column--primary {
    display: block;
    max-height: calc(min(78vh, 560px) - 69px);
    overflow-y: auto;
    border-right: 1px solid #e2e8f0;
    border-bottom: none;
    background: #f8fafc;
    scrollbar-width: none;
  }

  .category-cascader__column--primary::-webkit-scrollbar {
    display: none;
  }

  .category-cascader__column--secondary {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    align-content: start;
    gap: 8px;
    max-height: calc(min(78vh, 560px) - 69px);
    overflow-y: auto;
  }

  .category-cascader__option {
    min-height: 42px;
    border-radius: 12px;
    padding: 10px 12px;
  }

  .category-cascader__results {
    max-height: calc(min(78vh, 560px) - 69px);
  }

  .category-cascader__column--primary .category-cascader__option + .category-cascader__option {
    margin-top: 8px;
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
