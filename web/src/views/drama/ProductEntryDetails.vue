<!--
/**
 * 模块说明：数字丝路商品录入第三步。
 * 业务场景：在市场平台确定后，补充商品描述、材质、卖点、人群、场景和敏感功效信息。
 * 核心职责：把合规分析需要的细节写入商品草稿，并在提交后进入合规分析页。
 */
-->
<template>
  <ProductEntryFlowShell active-step="details">
    <section class="product-entry-card product-entry-card--details">
      <div class="product-entry-card__body">
        <div v-if="summaryItems.length" class="draft-summary" aria-label="已录入基础信息">
          <div v-for="item in summaryItems" :key="item.label" class="draft-summary__item">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
          </div>
        </div>

        <div class="field-block">
          <label class="field-block__label" for="product-description">
            商品描述
            <span class="field-block__required">*</span>
          </label>
          <textarea
            id="product-description"
            v-model.trim="form.description"
            class="field-block__control field-block__control--textarea"
            :class="{ 'field-block__control--error': Boolean(errors.description) }"
            placeholder="请详细描述您的商品，包括功能特点、材质、规格等信息..."
            maxlength="600"
            @input="handleFieldInput('description')"
          ></textarea>
          <p v-if="errors.description" class="field-block__error">{{ errors.description }}</p>
        </div>

        <div class="field-block">
          <label class="field-block__label" for="product-selling-points">核心卖点</label>
          <input
            id="product-selling-points"
            v-model.trim="form.coreSellingPoints"
            type="text"
            class="field-block__control"
            placeholder="例如: 便携、易清洗、适合户外使用"
            @input="persistDraft"
          />
        </div>

        <div class="product-entry-grid">
          <div class="field-block">
            <label class="field-block__label" for="product-weight">重量 (kg)</label>
            <input
              id="product-weight"
              v-model.trim="form.weight"
              type="text"
              inputmode="decimal"
              class="field-block__control"
              placeholder="0.00"
              @input="persistDraft"
            />
          </div>

          <div class="field-block">
            <label class="field-block__label" for="product-dimensions">尺寸 (cm)</label>
            <input
              id="product-dimensions"
              v-model.trim="form.dimensions"
              type="text"
              class="field-block__control"
              placeholder="长 x 宽 x 高"
              @input="persistDraft"
            />
          </div>
        </div>

        <div class="field-block">
          <label class="field-block__label" for="product-keywords">关键词 (用于SEO优化)</label>
          <input
            id="product-keywords"
            v-model.trim="form.keywords"
            type="text"
            class="field-block__control"
            placeholder="用逗号分隔多个关键词"
            @input="persistDraft"
          />
        </div>

        <div class="supplement-panel">
          <div class="supplement-panel__head">
            <h2>补充详情</h2>
            <span>用于后续合规分析与内容生成</span>
          </div>

          <div class="product-entry-grid">
            <div class="field-block">
              <label class="field-block__label" for="product-material">材质 / 成分</label>
              <input
                id="product-material"
                v-model.trim="form.material"
                type="text"
                class="field-block__control"
                placeholder="例如: ABS 塑料、硅胶表带"
                @input="persistDraft"
              />
            </div>

            <div class="field-block">
              <label class="field-block__label" for="product-audience">适用人群</label>
              <input
                id="product-audience"
                v-model.trim="form.audience"
                type="text"
                class="field-block__control"
                placeholder="例如: 户外运动人群、办公用户"
                @input="persistDraft"
              />
            </div>
          </div>

          <div class="product-entry-grid">
            <div class="field-block">
              <label class="field-block__label" for="product-price">价格区间</label>
              <input
                id="product-price"
                v-model.trim="form.priceRange"
                type="text"
                class="field-block__control"
                placeholder="例如: 29-49 USD"
                @input="persistDraft"
              />
            </div>

            <div class="field-block">
              <label class="field-block__label" for="product-specs">商品规格</label>
              <input
                id="product-specs"
                v-model.trim="form.specifications"
                type="text"
                class="field-block__control"
                placeholder="颜色、容量、套装、型号等"
                @input="persistDraft"
              />
            </div>
          </div>

          <div class="field-block">
            <label class="field-block__label" for="product-scenes">使用场景</label>
            <textarea
              id="product-scenes"
              v-model.trim="form.scenarios"
              class="field-block__control field-block__control--textarea field-block__control--short"
              placeholder="例如: 日常通勤、健身记录、跨境直播间展示等"
              maxlength="260"
              @input="persistDraft"
            ></textarea>
          </div>

          <div class="field-block">
            <label class="field-block__label" for="product-notes">注意事项</label>
            <textarea
              id="product-notes"
              v-model.trim="form.notes"
              class="field-block__control field-block__control--textarea field-block__control--short"
              placeholder="输入包装、运输、认证、禁用语等需要提示的信息"
              maxlength="260"
              @input="persistDraft"
            ></textarea>
          </div>

          <label class="sensitive-toggle">
            <input v-model="form.hasSensitiveClaims" type="checkbox" @change="persistDraft" />
            <span class="sensitive-toggle__box" aria-hidden="true"></span>
            <span class="sensitive-toggle__copy">
              <strong>涉及敏感功效描述</strong>
              <small>例如治疗、瘦身、杀菌、防疾病等需要重点合规审核的表述</small>
            </span>
          </label>
        </div>

        <div class="field-block">
          <label class="field-block__label">商品图片 / 补充素材</label>
          <input ref="fileInputRef" type="file" class="upload-input" multiple @change="handleFileChange" />
          <button type="button" class="upload-zone" @click="openFileDialog">
            <img :src="uploadIcon" alt="" class="upload-zone__icon" />
            <span class="upload-zone__copy">
              <strong>{{ isUploadingAttachments ? '素材上传中...' : '上传商品图片或素材' }}</strong>
              <small>上传后保存为后端素材 URL，供合规分析和内容生成继续复用</small>
            </span>
          </button>
          <div v-if="attachmentDisplayItems.length" class="attachment-list">
            <span v-for="item in attachmentDisplayItems" :key="item.key">{{ item.label }}</span>
          </div>
        </div>
      </div>

      <div class="product-entry-card__footer">
        <button type="button" class="footer-button footer-button--ghost" @click="goPrevious">上一步</button>

        <button type="button" class="footer-button footer-button--primary footer-button--submit" @click="submitEntry">
          <span>开始合规分析</span>
          <img :src="arrowRightIcon" alt="" />
        </button>
      </div>
    </section>
  </ProductEntryFlowShell>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import ProductEntryFlowShell from '@/components/product-entry/ProductEntryFlowShell.vue'
import arrowRightIcon from '@/assets/figma/product-entry/arrow-right.svg'
import uploadIcon from '@/assets/figma/product-entry/upload.svg'
import { notificationAPI } from '@/api/notification'
import { uploadAPI } from '@/api/upload'
import { saveCreateDramaDraft } from '@/utils/createDramaDraft'
import {
  buildCreateDramaDraftFromProductEntry,
  readProductEntryDraft,
  writeProductEntryDraft
} from '@/utils/productEntryDraft'
import type { ProductEntryDetails } from '@/utils/productEntryDraft'
import type { ProductAttachment } from '@/types/product'

const router = useRouter()
const fileInputRef = ref<HTMLInputElement | null>(null)
const isUploadingAttachments = ref(false)
const basicSummary = reactive({
  title: '',
  category: '',
  market: ''
})

const form = reactive<ProductEntryDetails>({
  description: '',
  coreSellingPoints: '',
  weight: '',
  dimensions: '',
  keywords: '',
  material: '',
  audience: '',
  scenarios: '',
  priceRange: '',
  specifications: '',
  notes: '',
  hasSensitiveClaims: false,
  attachmentNames: [],
  attachments: []
})

const errors = reactive({
  description: ''
})

const summaryItems = computed(() => {
  return [
    { label: '商品名称', value: basicSummary.title },
    { label: '商品品类', value: basicSummary.category },
    { label: '目标市场', value: basicSummary.market }
  ].filter((item) => item.value)
})

const attachmentDisplayItems = computed(() => {
  if (form.attachments.length) {
    return form.attachments.map((item) => ({
      key: item.url,
      label: item.name
    }))
  }

  return form.attachmentNames.map((name) => ({
    key: name,
    label: name
  }))
})

const persistDraft = () => {
  const sellingPointSource = form.coreSellingPoints || form.keywords
  // 详情页字段会直接影响合规风险和脚本内容，按输入实时写入草稿，减少用户离开页面后的信息损失。
  const patch = {
    description: form.description,
    coreSellingPoints: sellingPointSource
      .split(/[\/,，;；、\n]/)
      .map((item) => item.trim())
      .filter(Boolean),
    materialSpec: form.material,
    targetAudience: form.audience,
    usageScenario: form.scenarios,
    budgetPreference: form.priceRange,
    attachments: [...form.attachments]
  }

  writeProductEntryDraft({
    ...patch,
    ...(form.notes
      ? form.notes
          .split(/[；;\n]/)
          .map((item) => item.trim())
          .filter(Boolean)
          .length
        ? {
            complianceHints: form.notes
              .split(/[；;\n]/)
              .map((item) => item.trim())
              .filter(Boolean)
          }
        : {}
      : {})
  })

  const createDraft = buildCreateDramaDraftFromProductEntry()
  if (createDraft) {
    saveCreateDramaDraft(createDraft)
  }
}

const handleFieldInput = (field: keyof typeof errors) => {
  errors[field] = ''
  persistDraft()
}

const restoreDraft = () => {
  const draft = readProductEntryDraft()

  // 顶部摘要只展示前两步关键上下文，帮助用户确认当前详情是补给哪个商品和市场。
  basicSummary.title = draft.productName
  basicSummary.category = draft.category
  basicSummary.market = draft.targetMarket

  form.description = draft.description
  form.coreSellingPoints = draft.coreSellingPoints.join('、')
  form.weight = ''
  form.dimensions = ''
  form.keywords = ''
  form.material = draft.materialSpec
  form.audience = draft.targetAudience
  form.scenarios = draft.usageScenario
  form.priceRange = draft.budgetPreference
  form.specifications = ''
  form.notes = (draft.complianceHints || []).join('；')
  form.hasSensitiveClaims = false
  form.attachments = [...(draft.attachments || [])]
  form.attachmentNames = form.attachments.map((item) => item.name)
}

const validateStep = () => {
  let valid = true

  if (!form.description.trim()) {
    errors.description = '请输入商品描述'
    valid = false
  }

  return valid
}

const openFileDialog = () => {
  fileInputRef.value?.click()
}

const normalizeUploadedAttachment = (file: File, uploaded: Awaited<ReturnType<typeof uploadAPI.uploadFile>>): ProductAttachment => ({
  name: uploaded.filename || file.name,
  url: uploaded.url,
  size: uploaded.size || file.size,
  mimeType: uploaded.content_type || file.type,
  type: uploaded.asset_type || 'image',
  category: uploaded.category
})

const uploadAttachments = async (files: File[]) => {
  if (files.length === 0 || isUploadingAttachments.value) return

  isUploadingAttachments.value = true
  try {
    const availableSlots = Math.max(0, 8 - form.attachments.length)
    const nextFiles = files.slice(0, availableSlots)
    if (nextFiles.length < files.length) {
      ElMessage.warning('最多保留 8 个补充素材')
    }

    const results = await Promise.allSettled(
      nextFiles.map(async (file) => normalizeUploadedAttachment(
        file,
        await uploadAPI.uploadFile(file, { category: 'product-attachments' })
      ))
    )
    const uploaded = results
      .filter((item): item is PromiseFulfilledResult<ProductAttachment> => item.status === 'fulfilled')
      .map((item) => item.value)

    if (uploaded.length) {
      const byUrl = new Map<string, ProductAttachment>()
      ;[...form.attachments, ...uploaded].forEach((item) => byUrl.set(item.url, item))
      form.attachments = Array.from(byUrl.values()).slice(0, 8)
      form.attachmentNames = form.attachments.map((item) => item.name)
      persistDraft()
      ElMessage.success(`已上传 ${uploaded.length} 个素材`)
    }

    if (uploaded.length < nextFiles.length) {
      ElMessage.warning('部分素材上传失败，请检查文件类型或大小后重试')
    }
  } finally {
    isUploadingAttachments.value = false
  }
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = Array.from(target.files || [])
  void uploadAttachments(files)

  target.value = ''
}

const goPrevious = () => {
  persistDraft()
  router.push('/product-entry/market')
}

const submitEntry = () => {
  if (!validateStep()) {
    ElMessage.warning('请先完善必填信息')
    return
  }

  persistDraft()
  // 录入完成通知用于工作台提示，不作为合规分析前置条件，避免通知服务异常阻断主流程。
  void notificationAPI
    .create({
      type: 'product_entry_submitted',
      title: '商品信息已提交',
      content: basicSummary.title
        ? `「${basicSummary.title}」的商品信息已录入完成。`
        : '商品信息已录入完成。',
      path: '/compliance',
      metadata: {
        market: basicSummary.market,
        category: basicSummary.category
      }
    })
    .catch(() => {})

  router.push('/compliance')
}

onMounted(restoreDraft)
</script>

<style scoped>
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
  gap: 24px;
  padding: 33px 33px 0;
}

.draft-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.draft-summary__item {
  min-width: 0;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: linear-gradient(155deg, rgba(6, 182, 212, 0.05) 0%, rgba(124, 58, 237, 0.05) 100%);
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.draft-summary__item span {
  color: #62748e;
  font-size: 12px;
  line-height: 16px;
}

.draft-summary__item strong {
  min-width: 0;
  overflow: hidden;
  color: #0a2463;
  font-size: 15px;
  font-weight: 700;
  line-height: 22px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.field-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-block__label {
  color: #0a2463;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

.field-block__required {
  color: #fb2c36;
}

.field-block__control {
  width: 100%;
  height: 50px;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #f8fafc;
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

.field-block__control:hover {
  border-color: #cad5e2;
}

.field-block__control:focus {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.08);
  background: #ffffff;
}

.field-block__control--textarea {
  min-height: 154px;
  resize: vertical;
  line-height: 24px;
}

.field-block__control--short {
  min-height: 96px;
}

.field-block__control--error {
  border-color: #fb7185;
}

.field-block__error {
  margin: 0;
  color: #e11d48;
  font-size: 13px;
  line-height: 18px;
}

.product-entry-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 24px;
}

.supplement-panel {
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: linear-gradient(180deg, rgba(248, 250, 252, 0.74), #ffffff);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.supplement-panel__head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
}

.supplement-panel__head h2 {
  margin: 0;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 18px;
  font-weight: 700;
  line-height: 28px;
}

.supplement-panel__head span {
  color: #90a1b9;
  font-size: 13px;
  line-height: 20px;
}

.sensitive-toggle {
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #ffffff;
  padding: 14px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  transition:
    border-color 180ms ease,
    box-shadow 180ms ease;
}

.sensitive-toggle:hover {
  border-color: #cad5e2;
}

.sensitive-toggle input {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.sensitive-toggle__box {
  width: 22px;
  height: 22px;
  border: 2px solid #cbd5e1;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  transition:
    border-color 180ms ease,
    background 180ms ease;
}

.sensitive-toggle input:checked + .sensitive-toggle__box {
  border-color: #06b6d4;
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
}

.sensitive-toggle input:checked + .sensitive-toggle__box::after {
  content: '';
  width: 8px;
  height: 4px;
  border-bottom: 2px solid #ffffff;
  border-left: 2px solid #ffffff;
  transform: rotate(-45deg) translate(1px, -1px);
}

.sensitive-toggle__copy {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.sensitive-toggle__copy strong {
  color: #0a2463;
  font-size: 15px;
  font-weight: 700;
  line-height: 22px;
}

.sensitive-toggle__copy small {
  color: #62748e;
  font-size: 13px;
  line-height: 20px;
}

.upload-input {
  display: none;
}

.upload-zone {
  min-height: 96px;
  border: 2px dashed #cad5e2;
  border-radius: 16px;
  background: transparent;
  padding: 20px 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  text-align: left;
  cursor: pointer;
  transition:
    border-color 180ms ease,
    background-color 180ms ease,
    box-shadow 180ms ease;
}

.upload-zone:hover {
  border-color: #06b6d4;
  background: rgba(248, 250, 252, 0.76);
}

.upload-zone:focus-visible,
.footer-button:focus-visible {
  outline: none;
  box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.09);
}

.upload-zone__icon {
  width: 48px;
  height: 48px;
  display: block;
  flex: 0 0 auto;
}

.upload-zone__copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.upload-zone__copy strong {
  color: #45556c;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

.upload-zone__copy small {
  color: #90a1b9;
  font-size: 14px;
  line-height: 20px;
}

.attachment-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.attachment-list span {
  max-width: 240px;
  overflow: hidden;
  border-radius: 999px;
  background: #f1f5f9;
  padding: 6px 12px;
  color: #45556c;
  font-size: 13px;
  line-height: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-entry-card__footer {
  margin: 28px 33px;
  padding-top: 28px;
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
}

.footer-button--ghost:hover {
  background: #e8eef6;
}

.footer-button--primary {
  width: 108px;
  color: #ffffff;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
}

.footer-button--primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 24px rgba(99, 102, 241, 0.24);
}

.footer-button:active {
  transform: translateY(0);
}

@media (max-width: 1120px) {
  .product-entry-card {
    width: 100%;
  }
}

@media (max-width: 760px) {
  .draft-summary,
  .product-entry-grid {
    grid-template-columns: 1fr;
  }

  .supplement-panel__head {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 640px) {
  .product-entry-card__body {
    padding: 24px 20px 0;
  }

  .supplement-panel {
    padding: 16px;
  }

  .upload-zone {
    align-items: flex-start;
    flex-direction: column;
  }

  .product-entry-card__footer {
    margin: 24px 20px;
    flex-direction: column-reverse;
    gap: 12px;
  }

  .footer-button,
  .footer-button--ghost,
  .footer-button--primary {
    width: 100%;
  }
}
</style>
