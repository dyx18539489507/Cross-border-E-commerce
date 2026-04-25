<template>
  <div class="page-container compliance-analysis-page">
    <header class="compliance-header">
      <div class="compliance-header__inner">
        <div class="compliance-header__left">
          <button type="button" class="brand-link" @click="router.push('/')">
            <span class="brand-link__mark">
              <img :src="brandIcon" alt="" />
            </span>
            <span class="brand-link__name">{{ t('app.name') }}</span>
          </button>

          <nav class="compliance-nav" aria-label="主导航">
            <button
              v-for="item in navItems"
              :key="item.label"
              type="button"
              class="compliance-nav__item"
              :class="{ 'compliance-nav__item--active': item.active }"
              @click="handleNavClick(item.path)"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>

        <div class="compliance-header__right">
          <button type="button" class="header-icon-button" aria-label="通知">
            <img :src="bellIcon" alt="" />
            <span class="header-icon-button__dot"></span>
          </button>
        </div>
      </div>
    </header>

    <main class="compliance-main">
      <div class="compliance-shell">
        <section class="compliance-hero">
          <h1>合规分析报告</h1>
          <p>基于全球法规数据库的智能合规检测结果</p>
        </section>

        <section class="product-summary-card">
          <div class="product-summary-card__info">
            <div class="product-summary-card__emoji">📱</div>
            <div class="product-summary-card__copy">
              <h2>{{ reportProduct.title }}</h2>
              <p>{{ productSubtitle }}</p>
            </div>
          </div>

          <div class="product-summary-card__time">
            <span>检测时间</span>
            <strong>{{ reportMeta.detectedAt }}</strong>
          </div>
        </section>

        <section class="risk-overview-card">
          <div class="risk-overview-card__copy">
            <div class="risk-overview-card__title-row">
              <img :src="riskShieldIcon" alt="" class="risk-overview-card__title-icon" />
              <h3>综合风险评估</h3>
            </div>
            <p>基于200+法规条款的智能分析</p>

            <div class="risk-pill">
              <img :src="riskWarningIcon" alt="" />
              <span>中等风险</span>
            </div>
          </div>

          <div class="risk-score-ring">
            <div class="risk-score-ring__art"></div>
            <div class="risk-score-ring__center">
              <strong>{{ reportMeta.score }}</strong>
              <span>/ 100</span>
            </div>
          </div>
        </section>

        <section class="compliance-details">
          <h3>详细检测项</h3>

          <article
            v-for="section in reportSections"
            :key="section.title"
            class="detail-card"
          >
            <header class="detail-card__header">
              <div class="detail-card__title">
                <img :src="sectionDocIcon" alt="" />
                <h4>{{ section.title }}</h4>
              </div>

              <div class="detail-card__status">
                <div class="detail-card__risk-level">
                  <span>风险等级</span>
                  <strong>{{ section.riskLevel }}</strong>
                </div>
                <span class="detail-card__badge" :class="`detail-card__badge--${section.tone}`">
                  {{ section.badge }}
                </span>
              </div>
            </header>

            <div class="detail-card__body">
              <div
                v-for="item in section.items"
                :key="item.title"
                class="detail-item"
              >
                <img
                  :src="getItemIcon(item.icon)"
                  alt=""
                  class="detail-item__icon"
                  :class="`detail-item__icon--${item.icon}`"
                />
                <div class="detail-item__copy">
                  <strong>{{ item.title }}</strong>
                  <span>{{ item.description }}</span>
                </div>
              </div>
            </div>
          </article>
        </section>

        <section class="ai-advice-card">
          <div class="ai-advice-card__title">
            <img :src="aiSuggestionIcon" alt="" />
            <h3>AI 优化建议</h3>
          </div>

          <ol class="ai-advice-list">
            <li v-for="item in aiSuggestions" :key="item">
              <span class="ai-advice-list__index">{{ aiSuggestions.indexOf(item) + 1 }}</span>
              <span class="ai-advice-list__text">{{ item }}</span>
            </li>
          </ol>
        </section>

        <section class="compliance-actions">
          <button
            type="button"
            class="compliance-actions__primary"
            :disabled="creatingFlow"
            @click="handleContinue"
          >
            <span>{{ creatingFlow ? '正在创建项目...' : '继续生成脚本与分镜' }}</span>
            <img :src="ctaArrowIcon" alt="" />
          </button>

          <button type="button" class="compliance-actions__secondary" @click="handleExportPdf">
            下载完整报告 PDF
          </button>

          <button type="button" class="compliance-actions__ghost" @click="router.push('/dramas')">
            返回工作台
          </button>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { dramaAPI } from '@/api/drama'
import aiSuggestionIcon from '@/assets/compliance-analysis/ai-suggestion.svg'
import itemInfoIcon from '@/assets/compliance-analysis/item-info.svg'
import itemWarningIcon from '@/assets/compliance-analysis/item-warning.svg'
import riskShieldIcon from '@/assets/compliance-analysis/risk-shield.svg'
import riskWarningIcon from '@/assets/compliance-analysis/risk-warning.svg'
import sectionDocIcon from '@/assets/compliance-analysis/section-doc.svg'
import bellIcon from '@/assets/product-entry/bell.svg'
import brandIcon from '@/assets/product-entry/brand-icon.svg'
import ctaArrowIcon from '@/assets/product-entry/arrow-right.svg'
import { buildCreateDramaPayload } from '@/utils/compliance'
import { clearCreateDramaDraft, peekCreateDramaDraft } from '@/utils/createDramaDraft'
import { buildEpisodeStagePath, saveEpisodeWorkflowContext } from '@/utils/episodeWorkflowContext'

type DetailTone = 'warning' | 'success' | 'info'

interface ProductEntryDraft {
  title: string
  category: string
  brand: string
}

interface ReportSection {
  title: string
  riskLevel: string
  badge: string
  tone: DetailTone
  items: Array<{
    title: string
    description: string
    icon: DetailTone
  }>
}

const PRODUCT_ENTRY_DRAFT_KEY = 'drama:create:product-entry:basic'

const router = useRouter()
const { t } = useI18n()
const creatingFlow = ref(false)

const navItems = [
  { label: '工作台', path: '/dramas', active: false },
  { label: '商品录入', path: '/dramas/create', active: false },
  { label: '合规分析', path: '', active: true },
  { label: '脚本/分镜', path: '', active: false },
  { label: '内容创作', path: '', active: false },
  { label: '视频剪辑', path: '', active: false },
  { label: '数据分析', path: '/analytics', active: false }
] as const

const reportMeta = reactive({
  detectedAt: '2026-04-18 14:30',
  score: 72
})

const reportProduct = reactive({
  title: '智能手表 Pro',
  category: '电子产品',
  market: '🇺🇸 美国'
})

const reportSections: ReportSection[] = [
  {
    title: '产品安全认证',
    riskLevel: '中',
    badge: '⚠ 需注意',
    tone: 'warning',
    items: [
      {
        title: 'FCC认证',
        description: '需要FCC认证，适用于所有无线电子设备',
        icon: 'warning'
      },
      {
        title: 'UL认证',
        description: '建议获取UL认证以提升产品信誉',
        icon: 'info'
      }
    ]
  },
  {
    title: '材料与成分',
    riskLevel: '低',
    badge: '✓ 通过',
    tone: 'success',
    items: [
      {
        title: 'RoHS指令',
        description: '材料符合欧盟RoHS有害物质限制',
        icon: 'success'
      },
      {
        title: 'REACH法规',
        description: '未检测到REACH高度关注物质',
        icon: 'success'
      }
    ]
  },
  {
    title: '标签与包装',
    riskLevel: '中',
    badge: '⚠ 需注意',
    tone: 'warning',
    items: [
      {
        title: '能效标签',
        description: '需要添加能效等级标签',
        icon: 'warning'
      },
      {
        title: '警告标识',
        description: '已包含必要的警告信息',
        icon: 'success'
      }
    ]
  },
  {
    title: '知识产权',
    riskLevel: '低',
    badge: 'ℹ 建议',
    tone: 'info',
    items: [
      {
        title: '商标检查',
        description: '建议进行商标查重以避免侵权风险',
        icon: 'info'
      },
      {
        title: '专利检查',
        description: '建议检查是否存在相关专利',
        icon: 'info'
      }
    ]
  }
]

const aiSuggestions = [
  '申请FCC认证，预计费用$2,000-$5,000，周期4-6周',
  '添加能效标签，可通过我们的合作实验室快速获取',
  '进行商标查重，避免潜在的知识产权纠纷',
  '考虑获取UL认证以提升品牌竞争力'
]

const productSubtitle = computed(() => `${reportProduct.category} · 目标市场: ${reportProduct.market}`)

const restoreDraft = () => {
  if (typeof window === 'undefined') {
    return
  }

  const raw = window.sessionStorage.getItem(PRODUCT_ENTRY_DRAFT_KEY)
  if (!raw) {
    return
  }

  try {
    const draft = JSON.parse(raw) as Partial<ProductEntryDraft>
    if (typeof draft.title === 'string' && draft.title.trim()) {
      reportProduct.title = draft.title.trim()
    }
    if (typeof draft.category === 'string' && draft.category.trim()) {
      reportProduct.category = draft.category.trim()
    }
  } catch {
    window.sessionStorage.removeItem(PRODUCT_ENTRY_DRAFT_KEY)
  }
}

const getItemIcon = (tone: DetailTone) => {
  if (tone === 'warning') {
    return itemWarningIcon
  }
  if (tone === 'success') {
    return aiSuggestionIcon
  }
  return itemInfoIcon
}

const handleNavClick = (path: string) => {
  if (!path) {
    return
  }

  router.push(path)
}

const handleContinue = async () => {
  if (creatingFlow.value) return

  const draft = peekCreateDramaDraft()
  if (!draft) {
    ElMessage.warning('未找到商品录入信息，请先返回商品录入页完善资料')
    return
  }

  creatingFlow.value = true
  try {
    const payload = buildCreateDramaPayload(draft)
    const complianceResult = await dramaAPI.checkCompliance(payload)

    reportMeta.score = complianceResult.compliance.score
    reportMeta.detectedAt = formatDateTime(new Date())

    if (complianceResult.compliance.level === 'red') {
      ElMessage.error('当前商品合规风险过高，请先根据报告调整后再继续')
      return
    }

    const created = await dramaAPI.create({
      ...payload,
      compliance_token: complianceResult.compliance_token
    })

    const dramaId = String(created.drama.id)
    await dramaAPI.saveEpisodes(dramaId, [
      {
        episode_number: 1,
        title: '第1集',
        description: payload.description,
        script_content: ''
      }
    ])

    const refreshedDrama = await dramaAPI.get(dramaId)
    const firstEpisode = refreshedDrama.episodes?.find((episode) => episode.episode_number === 1)

    if (!firstEpisode?.id) {
      throw new Error('首集创建成功但未获取到章节信息')
    }

    saveEpisodeWorkflowContext({
      dramaId,
      episodeId: String(firstEpisode.id),
      episodeNumber: firstEpisode.episode_number
    })
    clearCreateDramaDraft()

    ElMessage.success('项目已创建，正在进入脚本与分镜阶段')
    await router.push(
      buildEpisodeStagePath('script', {
        dramaId,
        episodeId: String(firstEpisode.id),
        episodeNumber: firstEpisode.episode_number
      })
    )
  } catch (error: any) {
    ElMessage.error(error?.message || '创建业务流程失败，请稍后重试')
  } finally {
    creatingFlow.value = false
  }
}

const splitWrappedLines = (
  ctx: CanvasRenderingContext2D,
  text: string,
  maxWidth: number
) => {
  const blocks = text.split(/\n/)
  const lines: string[] = []

  for (const block of blocks) {
    if (!block) {
      lines.push('')
      continue
    }

    let line = ''
    for (const char of block) {
      const candidate = line + char
      if (!line || ctx.measureText(candidate).width <= maxWidth) {
        line = candidate
      } else {
        lines.push(line)
        line = char
      }
    }

    if (line) {
      lines.push(line)
    }
  }

  return lines.length ? lines : ['']
}

const textToUint8 = (value: string) => new TextEncoder().encode(value)

const concatUint8Arrays = (parts: Uint8Array[]) => {
  const totalLength = parts.reduce((sum, item) => sum + item.length, 0)
  const merged = new Uint8Array(totalLength)
  let offset = 0

  for (const item of parts) {
    merged.set(item, offset)
    offset += item.length
  }

  return merged
}

const buildPdfBlobFromJpegDataUrl = (
  jpegDataUrl: string,
  imageWidth: number,
  imageHeight: number
) => {
  const base64 = jpegDataUrl.split(',')[1] || ''
  const binary = atob(base64)
  const jpegBytes = new Uint8Array(binary.length)

  for (let index = 0; index < binary.length; index += 1) {
    jpegBytes[index] = binary.charCodeAt(index)
  }

  const pageWidth = 595.28
  const pageHeight = Number(((pageWidth * imageHeight) / imageWidth).toFixed(2))
  const contentStream = `q\n${pageWidth.toFixed(2)} 0 0 ${pageHeight.toFixed(2)} 0 0 cm\n/Im0 Do\nQ\n`

  const objects: Uint8Array[] = []
  objects.push(textToUint8('1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n'))
  objects.push(textToUint8('2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n'))
  objects.push(
    textToUint8(
      `3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ${pageWidth.toFixed(2)} ${pageHeight.toFixed(2)}] /Resources << /XObject << /Im0 4 0 R >> >> /Contents 5 0 R >>\nendobj\n`
    )
  )
  objects.push(
    concatUint8Arrays([
      textToUint8(
        `4 0 obj\n<< /Type /XObject /Subtype /Image /Width ${Math.round(imageWidth)} /Height ${Math.round(imageHeight)} /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length ${jpegBytes.length} >>\nstream\n`
      ),
      jpegBytes,
      textToUint8('\nendstream\nendobj\n')
    ])
  )
  objects.push(
    textToUint8(
      `5 0 obj\n<< /Length ${textToUint8(contentStream).length} >>\nstream\n${contentStream}endstream\nendobj\n`
    )
  )

  const header = textToUint8('%PDF-1.4\n%\xFF\xFF\xFF\xFF\n')
  const offsets: number[] = [0]
  let currentOffset = header.length

  for (const objectBytes of objects) {
    offsets.push(currentOffset)
    currentOffset += objectBytes.length
  }

  let xref = 'xref\n0 6\n0000000000 65535 f \n'
  for (let index = 1; index <= 5; index += 1) {
    xref += `${String(offsets[index]).padStart(10, '0')} 00000 n \n`
  }

  const xrefBytes = textToUint8(xref)
  const trailer = textToUint8(`trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n${currentOffset}\n%%EOF`)
  const pdfBytes = concatUint8Arrays([header, ...objects, xrefBytes, trailer])
  return new Blob([pdfBytes], { type: 'application/pdf' })
}

const downloadBlob = (blob: Blob, filename: string) => {
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  anchor.style.display = 'none'
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
  URL.revokeObjectURL(url)
}

const handleExportPdf = () => {
  const canvas = document.createElement('canvas')
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    ElMessage.error('PDF 导出失败，请稍后重试')
    return
  }

  const fontFamily = '"PingFang SC", "Microsoft YaHei", "Segoe UI", Arial, sans-serif'
  const canvasWidth = 1320
  const marginX = 72
  const contentWidth = canvasWidth - marginX * 2
  const lineHeight = 34

  const titleFont = `700 46px ${fontFamily}`
  const headingFont = `700 28px ${fontFamily}`
  const bodyFont = `500 22px ${fontFamily}`
  const metaFont = `500 20px ${fontFamily}`

  ctx.font = bodyFont
  const suggestionLines = aiSuggestions.map((item) => splitWrappedLines(ctx, item, contentWidth - 36))
  const sectionLines = reportSections.map((section) => ({
    ...section,
    items: section.items.map((item) => ({
      ...item,
      titleLines: splitWrappedLines(ctx, item.title, contentWidth - 60),
      descriptionLines: splitWrappedLines(ctx, item.description, contentWidth - 60)
    }))
  }))

  let totalHeight = 120
  totalHeight += 60
  totalHeight += 64
  totalHeight += 120
  totalHeight += 56
  totalHeight += 80
  totalHeight += 72

  for (const section of sectionLines) {
    totalHeight += 70
    totalHeight += section.items.reduce(
      (sum, item) => sum + item.titleLines.length * lineHeight + item.descriptionLines.length * lineHeight + 42,
      0
    )
    totalHeight += 34
  }

  totalHeight += 56
  totalHeight += suggestionLines.reduce((sum, lines) => sum + lines.length * lineHeight + 16, 0)
  totalHeight += 80

  canvas.width = canvasWidth
  canvas.height = Math.max(1800, Math.ceil(totalHeight))
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, canvas.width, canvas.height)

  let y = 86
  ctx.fillStyle = '#0a2463'
  ctx.font = titleFont
  ctx.fillText('合规分析报告', marginX, y)

  y += 52
  ctx.fillStyle = '#45556c'
  ctx.font = metaFont
  ctx.fillText('基于全球法规数据库的智能合规检测结果', marginX, y)

  y += 64
  ctx.fillStyle = '#0a2463'
  ctx.font = headingFont
  ctx.fillText(reportProduct.title, marginX, y)
  y += 36
  ctx.fillStyle = '#45556c'
  ctx.font = bodyFont
  ctx.fillText(productSubtitle.value, marginX, y)
  y += 36
  ctx.fillText(`检测时间：${reportMeta.detectedAt}`, marginX, y)

  y += 72
  ctx.fillStyle = '#0a2463'
  ctx.font = headingFont
  ctx.fillText('综合风险评估', marginX, y)
  y += 42
  ctx.fillStyle = '#f97316'
  ctx.fillText(`风险评分：${reportMeta.score} / 100`, marginX, y)
  y += 36
  ctx.fillText('风险等级：中等风险', marginX, y)

  y += 60
  for (const section of sectionLines) {
    ctx.fillStyle = '#0a2463'
    ctx.font = headingFont
    ctx.fillText(`${section.title}（${section.badge}）`, marginX, y)
    y += 44

    for (const item of section.items) {
      ctx.fillStyle = item.icon === 'warning' ? '#f97316' : item.icon === 'success' ? '#10b981' : '#06b6d4'
      ctx.beginPath()
      ctx.arc(marginX + 10, y - 8, 7, 0, Math.PI * 2)
      ctx.fill()

      ctx.fillStyle = '#0a2463'
      ctx.font = bodyFont
      for (const line of item.titleLines) {
        ctx.fillText(line, marginX + 34, y)
        y += lineHeight
      }

      ctx.fillStyle = '#45556c'
      for (const line of item.descriptionLines) {
        ctx.fillText(line, marginX + 34, y)
        y += lineHeight
      }

      y += 12
    }

    y += 18
  }

  ctx.fillStyle = '#0a2463'
  ctx.font = headingFont
  ctx.fillText('AI 优化建议', marginX, y)
  y += 44

  ctx.fillStyle = '#314158'
  ctx.font = bodyFont
  suggestionLines.forEach((lines, index) => {
    ctx.fillStyle = '#7c3aed'
    ctx.fillText(`${index + 1}.`, marginX, y)
    ctx.fillStyle = '#314158'
    lines.forEach((line) => {
      ctx.fillText(line, marginX + 28, y)
      y += lineHeight
    })
    y += 10
  })

  const jpegDataUrl = canvas.toDataURL('image/jpeg', 0.92)
  const pdfBlob = buildPdfBlobFromJpegDataUrl(jpegDataUrl, canvas.width, canvas.height)
  downloadBlob(pdfBlob, `合规校验报告_${reportMeta.detectedAt.replace(/[: ]/g, '-')}.pdf`)
  ElMessage.success('PDF 报告已下载')
}

onMounted(() => {
  restoreDraft()
})
</script>

<style scoped>
.page-container.compliance-analysis-page {
  --compliance-heading-font: 'Urbanist', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  --compliance-body-font: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  min-height: var(--app-vh, 100vh);
  padding: 0 !important;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%) !important;
  color: #0a2463;
  font-family: var(--compliance-body-font);
}

.compliance-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 20;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}

.compliance-header__inner {
  width: min(100%, 1075px);
  height: 64px;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.compliance-header__left {
  display: flex;
  align-items: center;
  gap: 32px;
  min-width: 0;
  flex: 1;
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
  width: 36px;
  height: 36px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background:
    linear-gradient(
      135deg,
      #0a2463 0%,
      #093873 7.1429%,
      #074c83 14.2857%,
      #066193 21.4286%,
      #0575a3 28.5714%,
      #048ab3 35.7143%,
      #05a0c3 42.8571%,
      #06b6d4 50%,
      #3aa8d9 57.1429%,
      #4f99dd 64.2857%,
      #5e8ae1 71.4286%,
      #6879e4 78.5714%,
      #7168e7 85.7143%,
      #7753ea 92.8571%,
      #7c3aed 100%
    );
}

.brand-link__mark img {
  width: 20px;
  height: 20px;
  display: block;
}

.brand-link__name {
  font-family: var(--compliance-body-font);
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  color: #0a2463;
  white-space: nowrap;
}

.compliance-nav {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;
}

.compliance-nav::-webkit-scrollbar {
  display: none;
}

.compliance-nav__item {
  font-family: var(--compliance-body-font);
  height: 32px;
  padding: 0 12px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: #45556c;
  font-size: 14px;
  line-height: 20px;
  font-weight: 400;
  white-space: nowrap;
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.compliance-nav__item:hover {
  color: #0a2463;
  background: rgba(241, 245, 249, 0.9);
}

.compliance-nav__item--active {
  color: #0a2463;
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.1) 0%, rgba(124, 58, 237, 0.1) 100%);
}

.compliance-header__right {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
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
}

.header-icon-button img {
  width: 20px;
  height: 20px;
  display: block;
}

.header-icon-button__dot {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #f97316;
}

.compliance-main {
  width: min(100%, 1075px);
  margin: 0 auto;
  padding: 104px 32px 48px;
}

.compliance-shell {
  display: flex;
  flex-direction: column;
  gap: 32px;
  width: min(100%, 1011px);
  margin: 0 auto;
}

.compliance-hero {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.compliance-hero h1 {
  font-family: var(--compliance-heading-font);
  font-size: 30px;
  line-height: 36px;
  font-weight: 700;
  color: #0a2463;
}

.compliance-hero p {
  font-family: var(--compliance-body-font);
  font-size: 16px;
  line-height: 24px;
  color: #45556c;
}

.product-summary-card,
.detail-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
}

.product-summary-card {
  min-height: 114px;
  padding: 25px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.product-summary-card__info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.product-summary-card__emoji {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background:
    linear-gradient(
      135deg,
      #06b6d4 0%,
      #2aafd6 7.1429%,
      #3aa8d9 14.2857%,
      #46a0db 21.4286%,
      #4f99dd 28.5714%,
      #5791df 35.7143%,
      #5e8ae1 42.8571%,
      #6382e2 50%,
      #6879e4 57.1429%,
      #6d71e6 64.2857%,
      #7168e7 71.4286%,
      #745ee9 78.5714%,
      #7753ea 85.7143%,
      #7a48ec 92.8571%,
      #7c3aed 100%
    );
  font-size: 30px;
  line-height: 36px;
  color: #0f172a;
}

.product-summary-card__copy {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.product-summary-card__copy h2 {
  font-family: var(--compliance-heading-font);
  font-size: 20px;
  line-height: 28px;
  font-weight: 700;
  color: #0a2463;
}

.product-summary-card__copy p {
  font-family: var(--compliance-body-font);
  font-size: 16px;
  line-height: 24px;
  color: #45556c;
}

.product-summary-card__time {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.product-summary-card__time span {
  font-family: var(--compliance-body-font);
  font-size: 14px;
  line-height: 20px;
  color: #62748e;
}

.product-summary-card__time strong {
  font-family: var(--compliance-body-font);
  font-size: 16px;
  line-height: 24px;
  font-weight: 500;
  color: #0a2463;
}

.risk-overview-card {
  min-height: 196px;
  padding: 34px;
  border: 2px solid rgba(249, 115, 22, 0.2);
  border-radius: 16px;
  background: linear-gradient(169.028deg, #fff7ed 0%, #fef2f2 100%);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.risk-overview-card__copy {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.risk-overview-card__title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.risk-overview-card__title-icon {
  width: 32px;
  height: 32px;
  display: block;
}

.risk-overview-card__copy h3 {
  font-family: var(--compliance-heading-font);
  font-size: 24px;
  line-height: 32px;
  font-weight: 700;
  color: #0a2463;
}

.risk-overview-card__copy p {
  font-family: var(--compliance-body-font);
  font-size: 16px;
  line-height: 24px;
  color: #45556c;
}

.risk-pill {
  width: 124px;
  height: 40px;
  border-radius: 16px;
  background: #ffffff;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
}

.risk-pill img {
  width: 20px;
  height: 20px;
  display: block;
}

.risk-pill span {
  font-family: var(--compliance-body-font);
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  color: #f97316;
}

.risk-score-ring {
  position: relative;
  width: 128px;
  height: 128px;
  flex-shrink: 0;
}

.risk-score-ring__art {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: conic-gradient(from -90deg, #f97316 0 72%, #e2e8f0 72% 100%);
}

.risk-score-ring__art::after {
  content: '';
  position: absolute;
  inset: 12px;
  border-radius: 50%;
  background: #fff7ed;
}

.risk-score-ring__center {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.risk-score-ring__center strong {
  font-family: var(--compliance-body-font);
  font-size: 30px;
  line-height: 36px;
  font-weight: 700;
  color: #0a2463;
}

.risk-score-ring__center span {
  font-family: var(--compliance-body-font);
  font-size: 12px;
  line-height: 16px;
  color: #62748e;
}

.compliance-details {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.compliance-details > h3,
.ai-advice-card__title h3 {
  font-family: var(--compliance-heading-font);
  font-size: 20px;
  line-height: 28px;
  font-weight: 700;
  color: #0a2463;
}

.detail-card {
  overflow: hidden;
}

.detail-card__header {
  min-height: 93px;
  padding: 24px;
  border-bottom: 1px solid #e2e8f0;
  background: linear-gradient(90deg, #f8fafc 0%, #ffffff 100%);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.detail-card__title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.detail-card__title img {
  width: 24px;
  height: 24px;
  display: block;
}

.detail-card__title h4 {
  font-family: var(--compliance-heading-font);
  font-size: 18px;
  line-height: 28px;
  font-weight: 600;
  color: #0a2463;
}

.detail-card__status {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.detail-card__risk-level {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: flex-end;
}

.detail-card__risk-level span {
  font-family: var(--compliance-body-font);
  font-size: 12px;
  line-height: 16px;
  color: #62748e;
}

.detail-card__risk-level strong {
  font-family: var(--compliance-body-font);
  font-size: 16px;
  line-height: 24px;
  font-weight: 500;
  color: #0a2463;
}

.detail-card__badge {
  min-width: 72px;
  height: 36px;
  padding: 0 16px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: var(--compliance-body-font);
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.detail-card__badge--warning {
  background: #ffedd4;
  color: #f97316;
}

.detail-card__badge--success {
  background: #dcfce7;
  color: #10b981;
}

.detail-card__badge--info {
  background: #dbeafe;
  color: #06b6d4;
}

.detail-card__body {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-item {
  min-height: 80px;
  border-radius: 16px;
  background: #f8fafc;
  padding: 16px;
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.detail-item__icon {
  width: 20px;
  height: 20px;
  display: block;
  flex-shrink: 0;
}

.detail-item__copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item__copy strong {
  font-family: var(--compliance-body-font);
  font-size: 16px;
  line-height: 24px;
  font-weight: 500;
  color: #0a2463;
}

.detail-item__copy span {
  font-family: var(--compliance-body-font);
  font-size: 14px;
  line-height: 20px;
  color: #45556c;
}

.ai-advice-card {
  min-height: 258px;
  padding: 33px;
  border: 1px solid rgba(6, 182, 212, 0.2);
  border-radius: 16px;
  background: linear-gradient(165.684deg, rgba(6, 182, 212, 0.05) 0%, rgba(124, 58, 237, 0.05) 100%);
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.ai-advice-card__title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.ai-advice-card__title img {
  width: 28px;
  height: 28px;
  display: block;
}

.ai-advice-list {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ai-advice-list li {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.ai-advice-list__index {
  width: 24px;
  height: 24px;
  border-radius: 999px;
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: var(--compliance-body-font);
  font-size: 12px;
  line-height: 16px;
  font-weight: 700;
  color: #ffffff;
  flex-shrink: 0;
  margin-top: 2px;
}

.ai-advice-list__text {
  font-family: var(--compliance-body-font);
  font-size: 16px;
  line-height: 24px;
  color: #314158;
}

.compliance-actions {
  display: flex;
  align-items: stretch;
  gap: 16px;
}

.compliance-actions button {
  border: none;
  border-radius: 16px;
  font-family: var(--compliance-body-font);
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background-color 0.2s ease;
}

.compliance-actions button:hover {
  transform: translateY(-1px);
}

.compliance-actions__primary {
  flex: 1;
  min-height: 60px;
  padding: 18px 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #ffffff;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
}

.compliance-actions__primary img {
  width: 20px;
  height: 20px;
  display: block;
}

.compliance-actions__secondary {
  width: 182px;
  min-height: 60px;
  background: #ffffff;
  border: 2px solid #e2e8f0 !important;
  color: #0a2463;
}

.compliance-actions__ghost {
  width: 128px;
  min-height: 60px;
  background: #f1f5f9;
  color: #0a2463;
}

@media (max-width: 1080px) {
  .compliance-main {
    padding-left: 24px;
    padding-right: 24px;
  }
}

@media (max-width: 920px) {
  .detail-card__header,
  .product-summary-card,
  .risk-overview-card {
    flex-direction: column;
    align-items: flex-start;
  }

  .product-summary-card__time,
  .detail-card__risk-level {
    align-items: flex-start;
  }

  .detail-card__status {
    width: 100%;
    justify-content: space-between;
  }

  .compliance-actions {
    flex-direction: column;
  }

  .compliance-actions__secondary,
  .compliance-actions__ghost {
    width: 100%;
  }
}

@media (max-width: 720px) {
  .compliance-header__inner {
    height: auto;
    min-height: 64px;
    padding: 10px 16px;
    align-items: flex-start;
    flex-direction: column;
  }

  .compliance-header__left,
  .compliance-header__right {
    width: 100%;
  }

  .compliance-header__right {
    justify-content: space-between;
  }

  .compliance-main {
    padding: 132px 16px 40px;
  }

  .compliance-shell {
    gap: 24px;
  }

  .compliance-hero h1 {
    font-size: 28px;
    line-height: 34px;
  }

  .product-summary-card,
  .risk-overview-card,
  .ai-advice-card {
    padding: 24px 20px;
  }

  .detail-card__header,
  .detail-card__body {
    padding: 20px;
  }
}
</style>
