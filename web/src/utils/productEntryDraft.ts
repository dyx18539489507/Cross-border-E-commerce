/**
 * 模块说明：数字丝路商品录入草稿状态。
 * 业务场景：商品信息来自手动录入或丝路 Agent 结果，需要跨多步骤页面持续传递到合规分析和项目创建接口。
 * 核心职责：维护统一 ProductEntryDraft，同时兼容旧创建项目草稿字段，保证新旧页面能读取同一份商品上下文。
 */
import type { AgentInput, AgentResult } from '@/types/agent'
import type { CreateDramaRequest } from '@/types/drama'
import type { ProductAttachment, ProductEntryDraft, ProductEntrySource } from '@/types/product'

export type { ProductContext, ProductEntryDraft, ProductEntrySource } from '@/types/product'

export interface ProductEntryBasicInfo {
  title: string
  category: string
  categoryPrimary?: string
  categorySecondary?: string
  brand: string
  productImage?: string
}

export interface ProductEntryTargetMarket {
  marketCode: string
  marketName: string
  marketEmoji: string
  platform: string
  marketingGoals: string[]
}

export interface ProductEntryDetails {
  description: string
  coreSellingPoints: string
  weight: string
  dimensions: string
  keywords: string
  material: string
  audience: string
  scenarios: string
  priceRange: string
  specifications: string
  notes: string
  hasSensitiveClaims: boolean
  attachmentNames: string[]
  attachments: ProductAttachment[]
}

type LegacyProductEntryFlowDraft = {
  basicInfo?: Partial<ProductEntryBasicInfo>
  targetMarket?: Partial<ProductEntryTargetMarket>
  productDetails?: Partial<ProductEntryDetails>
}

type ProductEntryDraftPatch = Partial<ProductEntryDraft> & LegacyProductEntryFlowDraft

export const PRODUCT_ENTRY_DRAFT_KEY = 'drama:create:product-entry:context'
export const PRODUCT_ENTRY_BASIC_DRAFT_KEY = 'drama:create:product-entry:basic'
export const PRODUCT_ENTRY_FLOW_DRAFT_KEY = 'drama:create:product-entry:draft'

const nowIso = () => new Date().toISOString()

const readJSON = <T>(key: string): T | null => {
  if (typeof window === 'undefined') {
    return null
  }

  const raw = window.sessionStorage.getItem(key)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as T
  } catch {
    window.sessionStorage.removeItem(key)
    return null
  }
}

const writeJSON = (key: string, value: unknown) => {
  if (typeof window === 'undefined') {
    return
  }

  try {
    window.sessionStorage.setItem(key, JSON.stringify(value))
  } catch {
    if (key === PRODUCT_ENTRY_DRAFT_KEY && value && typeof value === 'object') {
      // 商品图片可能是较大的 Data URL，超出 sessionStorage 限制时先保留文本字段，避免录入流程整体丢失。
      const fallback = { ...(value as ProductEntryDraft), productImage: '' }
      window.sessionStorage.setItem(key, JSON.stringify(fallback))
    }
  }
}

const toText = (value: unknown) => {
  return typeof value === 'string' ? value.trim() : ''
}

const splitTextList = (value: string) => {
  return value
    .split(/[\/,，;；、\n]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

const toStringList = (value: unknown): string[] => {
  if (Array.isArray(value)) {
    return value.map((item) => String(item).trim()).filter(Boolean)
  }
  if (typeof value === 'string') {
    return splitTextList(value)
  }
  return []
}

const uniqueList = (items: unknown[]) => {
  return Array.from(new Set(items.flatMap((item) => toStringList(item)))).filter(Boolean)
}

const compactLines = (items: unknown[]) => {
  return items.map(toText).filter(Boolean).join('\n')
}

const compactSentenceList = (items: unknown[]) => {
  return items.map(toText).filter(Boolean)
}

const normalizeAttachments = (value: unknown): ProductAttachment[] => {
  if (!Array.isArray(value)) {
    return []
  }

  return value
    .map((item): ProductAttachment | null => {
      const raw = item && typeof item === 'object' ? item as Partial<ProductAttachment> : null
      if (!raw) return null
      const name = toText(raw.name)
      const url = toText(raw.url)
      if (!name || !url) return null
      return {
        name,
        url,
        size: Number(raw.size || 0),
        mimeType: toText(raw.mimeType),
        type: toText(raw.type),
        category: toText(raw.category) || undefined
      }
    })
    .filter((item): item is ProductAttachment => item !== null)
}

const parseCategoryParts = (category: string) => {
  const [categoryPrimary = '', categorySecondary = ''] = category.split('/').map((item) => item.trim())
  return { categoryPrimary, categorySecondary }
}

const hasDraftContent = (draft: ProductEntryDraft) => {
  return Boolean(
    draft.productName ||
      draft.category ||
      draft.brand ||
      draft.productImage ||
      draft.description ||
      draft.coreSellingPoints.length ||
      draft.materialSpec ||
      draft.usageScenario ||
      draft.targetMarket ||
      draft.targetPlatform ||
      draft.targetAudience ||
      draft.marketingGoal.length ||
      draft.budgetPreference ||
      draft.agentSummary ||
      draft.complianceHints?.length ||
      draft.localizationHints?.length ||
      draft.attachments?.length
  )
}

export const createEmptyProductDraft = (source: ProductEntrySource = 'manual'): ProductEntryDraft => {
  const timestamp = nowIso()
  return {
    source,
    productName: '',
    category: '',
    brand: '',
    productImage: '',
    description: '',
    coreSellingPoints: [],
    materialSpec: '',
    usageScenario: '',
    targetMarket: '',
    targetPlatform: '',
    targetAudience: '',
    marketingGoal: [],
    budgetPreference: '',
    agentSummary: '',
    complianceHints: [],
    localizationHints: [],
    attachments: [],
    createdAt: timestamp,
    updatedAt: timestamp
  }
}

/**
 * 功能：把任意历史草稿或当前草稿归一成 ProductEntryDraft。
 * 参数：value 可能是新版商品草稿、旧 basicInfo/targetMarket/productDetails 草稿或局部 patch。
 * 返回：可被商品录入、合规分析和项目创建共用的统一草稿；无法识别时返回 null。
 */
const normalizeProductDraft = (value: unknown): ProductEntryDraft | null => {
  if (!value || typeof value !== 'object') {
    return null
  }

  const raw = value as Partial<ProductEntryDraft> & {
    title?: string
    basicInfo?: Partial<ProductEntryBasicInfo>
    targetMarket?: unknown
    productDetails?: Partial<ProductEntryDetails>
  }
  const timestamp = nowIso()
  const rawTargetMarket = raw.targetMarket
  const legacyTargetMarket =
    rawTargetMarket && typeof rawTargetMarket === 'object'
      ? (rawTargetMarket as Partial<ProductEntryTargetMarket>)
      : null
  const targetMarketText =
    typeof rawTargetMarket === 'string' ? rawTargetMarket : toText(legacyTargetMarket?.marketName)
  const targetPlatformText = toText(raw.targetPlatform) || toText(legacyTargetMarket?.platform)
  const details = raw.productDetails || {}
  const basic = raw.basicInfo || {}

  const coreSellingPoints =
    toStringList(raw.coreSellingPoints).length > 0
      ? toStringList(raw.coreSellingPoints)
      : toStringList(details.coreSellingPoints || details.keywords)
  const marketingGoal =
    toStringList(raw.marketingGoal).length > 0
      ? toStringList(raw.marketingGoal)
      : toStringList(legacyTargetMarket?.marketingGoals)

  return {
    source: raw.source === 'agent' ? 'agent' : 'manual',
    productName: toText(raw.productName) || toText(raw.title) || toText(basic.title),
    category: toText(raw.category) || toText(basic.category),
    brand: toText(raw.brand) || toText(basic.brand),
    productImage: toText(raw.productImage) || toText(basic.productImage),
    description: toText(raw.description) || toText(details.description),
    coreSellingPoints,
    materialSpec: toText(raw.materialSpec) || toText(details.material) || toText(details.specifications),
    usageScenario: toText(raw.usageScenario) || toText(details.scenarios),
    targetMarket: targetMarketText,
    targetPlatform: targetPlatformText,
    targetAudience: toText(raw.targetAudience) || toText(details.audience),
    marketingGoal,
    budgetPreference: toText(raw.budgetPreference) || toText(details.priceRange),
    agentSummary: toText(raw.agentSummary),
    complianceHints: toStringList(raw.complianceHints),
    localizationHints: toStringList(raw.localizationHints),
    attachments: normalizeAttachments(raw.attachments),
    createdAt: toText(raw.createdAt) || timestamp,
    updatedAt: toText(raw.updatedAt) || timestamp
  }
}

const getLegacyProductDraft = (): ProductEntryDraft | null => {
  const basic = readJSON<Partial<ProductEntryBasicInfo>>(PRODUCT_ENTRY_BASIC_DRAFT_KEY) || {}
  const flow = readJSON<LegacyProductEntryFlowDraft>(PRODUCT_ENTRY_FLOW_DRAFT_KEY) || {}
  const merged = normalizeProductDraft({
    source: 'manual',
    basicInfo: {
      ...basic,
      ...(flow.basicInfo || {})
    },
    targetMarket: flow.targetMarket,
    productDetails: flow.productDetails
  })

  return merged && hasDraftContent(merged) ? merged : null
}

const persistLegacyDrafts = (draft: ProductEntryDraft) => {
  // 旧页面仍读取 basic/flow 两份草稿，写回它们可以让数字丝路新录入流和旧创建流保持兼容。
  const { categoryPrimary, categorySecondary } = parseCategoryParts(draft.category)
  const basicInfo: ProductEntryBasicInfo = {
    title: draft.productName,
    category: draft.category,
    categoryPrimary,
    categorySecondary,
    brand: draft.brand,
    productImage: draft.productImage
  }
  const targetMarket: ProductEntryTargetMarket = {
    marketCode: '',
    marketName: draft.targetMarket,
    marketEmoji: '',
    platform: draft.targetPlatform,
    marketingGoals: [...draft.marketingGoal]
  }
  const productDetails: ProductEntryDetails = {
    description: draft.description,
    coreSellingPoints: draft.coreSellingPoints.join('、'),
    weight: '',
    dimensions: '',
    keywords: '',
    material: draft.materialSpec,
    audience: draft.targetAudience,
    scenarios: draft.usageScenario,
    priceRange: draft.budgetPreference,
    specifications: '',
    notes: [...(draft.complianceHints || []), ...(draft.localizationHints || [])].join('；'),
    hasSensitiveClaims: false,
    attachmentNames: draft.attachments.map((item) => item.name),
    attachments: [...draft.attachments]
  }

  writeJSON(PRODUCT_ENTRY_BASIC_DRAFT_KEY, basicInfo)
  writeJSON(PRODUCT_ENTRY_FLOW_DRAFT_KEY, {
    basicInfo,
    targetMarket,
    productDetails
  })
}

const legacyPatchToProductPatch = (patch: ProductEntryDraftPatch): Partial<ProductEntryDraft> => {
  const next: Partial<ProductEntryDraft> = { ...patch }

  if (patch.basicInfo) {
    next.productName = patch.basicInfo.title ?? next.productName
    next.category = patch.basicInfo.category ?? next.category
    next.brand = patch.basicInfo.brand ?? next.brand
    next.productImage = patch.basicInfo.productImage ?? next.productImage
  }

  if (patch.targetMarket && typeof patch.targetMarket === 'object') {
    const legacyTargetMarket = patch.targetMarket as Partial<ProductEntryTargetMarket>
    next.targetMarket = legacyTargetMarket.marketName ?? next.targetMarket
    next.targetPlatform = legacyTargetMarket.platform ?? next.targetPlatform
    next.marketingGoal = legacyTargetMarket.marketingGoals ?? next.marketingGoal
  }

  if (patch.productDetails) {
    next.description = patch.productDetails.description ?? next.description
    next.materialSpec = patch.productDetails.material ?? next.materialSpec
    next.targetAudience = patch.productDetails.audience ?? next.targetAudience
    next.usageScenario = patch.productDetails.scenarios ?? next.usageScenario
    next.budgetPreference = patch.productDetails.priceRange ?? next.budgetPreference
    next.coreSellingPoints = toStringList(patch.productDetails.coreSellingPoints || patch.productDetails.keywords)
    next.attachments = patch.productDetails.attachments ?? next.attachments
  }

  delete (next as ProductEntryDraftPatch).basicInfo
  delete (next as ProductEntryDraftPatch).productDetails
  return next
}

export const getProductDraft = (): ProductEntryDraft | null => {
  const stored = normalizeProductDraft(readJSON<ProductEntryDraft>(PRODUCT_ENTRY_DRAFT_KEY))
  if (stored) {
    return stored
  }

  return getLegacyProductDraft()
}

export const saveProductDraft = (patch: ProductEntryDraftPatch): ProductEntryDraft => {
  const normalizedPatch = legacyPatchToProductPatch(patch)
  const current = getProductDraft() || createEmptyProductDraft(normalizedPatch.source || 'manual')
  const next = normalizeProductDraft({
    ...current,
    ...normalizedPatch,
    source: normalizedPatch.source || current.source,
    createdAt: current.createdAt,
    updatedAt: nowIso()
  }) || createEmptyProductDraft(normalizedPatch.source || current.source)

  writeJSON(PRODUCT_ENTRY_DRAFT_KEY, next)
  persistLegacyDrafts(next)
  return next
}

export const clearProductDraft = () => {
  if (typeof window === 'undefined') {
    return
  }

  window.sessionStorage.removeItem(PRODUCT_ENTRY_DRAFT_KEY)
  window.sessionStorage.removeItem(PRODUCT_ENTRY_BASIC_DRAFT_KEY)
  window.sessionStorage.removeItem(PRODUCT_ENTRY_FLOW_DRAFT_KEY)
}

export const startManualProductDraft = () => {
  clearProductDraft()
  return saveProductDraft(createEmptyProductDraft('manual'))
}

export const readProductEntryBasicInfo = (): Partial<ProductEntryBasicInfo> => {
  const draft = getProductDraft()
  if (!draft) {
    return {}
  }

  const { categoryPrimary, categorySecondary } = parseCategoryParts(draft.category)
  return {
    title: draft.productName,
    category: draft.category,
    categoryPrimary,
    categorySecondary,
    brand: draft.brand,
    productImage: draft.productImage
  }
}

export const saveProductEntryBasicInfo = (basicInfo: ProductEntryBasicInfo) => {
  saveProductDraft({
    productName: basicInfo.title,
    category: basicInfo.category,
    brand: basicInfo.brand,
    productImage: basicInfo.productImage || ''
  })
}

export const readProductEntryDraft = (): ProductEntryDraft => {
  return getProductDraft() || createEmptyProductDraft('manual')
}

export const writeProductEntryDraft = (patch: ProductEntryDraftPatch) => {
  return saveProductDraft(patch)
}

/**
 * 功能：把丝路 Agent 结果转换为商品录入草稿。
 * 参数：result 为 Agent 完整方案；input 为用户启动 Agent 时的原始商品和图片上下文。
 * 返回：source=agent 的 ProductEntryDraft，携带合规提示、本地化提示和后续合规分析所需字段。
 */
export const agentResultToProductDraft = (result: AgentResult, input?: AgentInput | null): ProductEntryDraft => {
  const recognized = (result.recognizedInfo || {}) as Partial<AgentResult['recognizedInfo']>
  const overview = (result.overview || {}) as Partial<AgentResult['overview']>
  const compliance = (result.compliance || {}) as Partial<AgentResult['compliance']>
  const localization = (result.localization || {}) as Partial<AgentResult['localization']>
  const promotion = (result.promotion || {}) as Partial<AgentResult['promotion']>
  const agentMessage = (result.agentMessage || {}) as Partial<AgentResult['agentMessage']>
  const timestamp = nowIso()

  // Agent 的合规模块不直接等同最终合规结论，这里只沉淀为“提示”，交由合规页再次基于目标市场校验。
  const complianceHints = uniqueList([
    compliance.summary,
    compliance.riskTags,
    compliance.missingInfo,
    compliance.suggestions,
    compliance.forbiddenExpressions,
    compliance.saferExpressions
  ])
  const localizationHints = uniqueList([
    localization.direction,
    localization.reason,
    localization.tone,
    localization.keywords,
    localization.sceneSuggestions
  ])
  const targetPlatform =
    toText(recognized.targetPlatform) || toStringList(promotion.platforms)[0] || toText(input?.targetPlatform)
  const coreSellingPoints = toStringList(recognized.coreSellingPoints).length
    ? toStringList(recognized.coreSellingPoints)
    : toStringList(input?.coreSellingPoints)

  return {
    source: 'agent',
    productName: toText(recognized.productName) || toText(input?.productName),
    category: toText(recognized.category) || toText(input?.category),
    brand: '',
    productImage: toText(input?.imageDataUrl),
    description: compactLines([
      recognized.imageUnderstanding,
      overview.marketStrategy
    ]),
    coreSellingPoints,
    materialSpec: toText(input?.materialSpec),
    usageScenario: toText(input?.usageScenario) || toStringList(localization.sceneSuggestions).join('、'),
    targetMarket: toText(recognized.targetMarket) || toText(input?.targetMarket),
    targetPlatform,
    targetAudience: toText(recognized.targetAudience) || toText(input?.targetAudience),
    marketingGoal: toStringList(promotion.focusMetrics),
    budgetPreference: '',
    agentSummary: compactLines([
      agentMessage.summary,
      overview.complianceRiskLevel ? `合规风险：${overview.complianceRiskLevel}` : '',
      overview.marketStrategy,
      overview.recommendedVideoStyle ? `推荐视频形式：${overview.recommendedVideoStyle}` : '',
      overview.recommendedDigitalHuman ? `推荐数字人：${overview.recommendedDigitalHuman}` : ''
    ]),
    complianceHints,
    localizationHints,
    attachments: [],
    createdAt: timestamp,
    updatedAt: timestamp
  }
}

export const saveAgentResultAsProductDraft = (result: AgentResult, input?: AgentInput | null) => {
  clearProductDraft()
  return saveProductDraft(agentResultToProductDraft(result, input))
}

/**
 * 功能：把商品录入草稿适配为后端现有项目创建/合规预检请求。
 * 参数：draft 默认为当前商品草稿，也可传入指定草稿。
 * 返回：CreateDramaRequest；商品名缺失时返回 null，避免创建无业务主体的合规任务。
 */
export const buildCreateDramaDraftFromProductEntry = (
  draft: ProductEntryDraft | null = getProductDraft()
): CreateDramaRequest | null => {
  if (!draft || !draft.productName.trim()) {
    return null
  }

  const descriptionParts = compactSentenceList([
    draft.description,
    draft.usageScenario ? `使用场景: ${draft.usageScenario}` : '',
    draft.targetAudience ? `目标人群: ${draft.targetAudience}` : '',
    draft.targetPlatform ? `目标平台: ${draft.targetPlatform}` : '',
    draft.attachments.length ? `已上传素材: ${draft.attachments.map((item) => `${item.name} ${item.url}`).join('；')}` : '',
    draft.agentSummary ? `Agent 摘要: ${draft.agentSummary}` : ''
  ])
  const sellingPoints = draft.coreSellingPoints.length
    ? draft.coreSellingPoints.join('、')
    : draft.marketingGoal.join('、')
  const tags = [draft.brand, draft.targetPlatform, draft.budgetPreference].filter(Boolean).join(', ')

  return {
    title: draft.productName.trim(),
    description:
      descriptionParts.join('\n') ||
      [draft.category, draft.brand, draft.targetMarket, draft.targetPlatform].filter(Boolean).join(' / ') ||
      draft.productName.trim(),
    target_country: draft.targetMarket ? [draft.targetMarket] : [],
    material_composition: draft.materialSpec,
    marketing_selling_points: sellingPoints,
    genre: draft.category || undefined,
    tags: tags || undefined
  }
}
