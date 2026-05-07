import type { CreateDramaRequest } from '@/types/drama'

export interface ProductEntryBasicInfo {
  title: string
  category: string
  categoryPrimary?: string
  categorySecondary?: string
  brand: string
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
}

export interface ProductEntryDraft {
  basicInfo: Partial<ProductEntryBasicInfo>
  targetMarket: Partial<ProductEntryTargetMarket>
  productDetails: Partial<ProductEntryDetails>
}

export const PRODUCT_ENTRY_BASIC_DRAFT_KEY = 'drama:create:product-entry:basic'
export const PRODUCT_ENTRY_FLOW_DRAFT_KEY = 'drama:create:product-entry:draft'

const readJSON = <T>(key: string): Partial<T> | null => {
  if (typeof window === 'undefined') {
    return null
  }

  const raw = window.sessionStorage.getItem(key)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as Partial<T>
  } catch {
    window.sessionStorage.removeItem(key)
    return null
  }
}

const writeJSON = (key: string, value: unknown) => {
  if (typeof window === 'undefined') {
    return
  }

  window.sessionStorage.setItem(key, JSON.stringify(value))
}

const normalizeStringArray = (value: unknown) => {
  if (!Array.isArray(value)) {
    return []
  }

  return value.map((item) => String(item).trim()).filter(Boolean)
}

export const readProductEntryBasicInfo = (): Partial<ProductEntryBasicInfo> => {
  const basicInfo = readJSON<ProductEntryBasicInfo>(PRODUCT_ENTRY_BASIC_DRAFT_KEY)
  return basicInfo || {}
}

export const saveProductEntryBasicInfo = (basicInfo: ProductEntryBasicInfo) => {
  writeJSON(PRODUCT_ENTRY_BASIC_DRAFT_KEY, basicInfo)
  writeProductEntryDraft({ basicInfo })
}

export const readProductEntryDraft = (): ProductEntryDraft => {
  const draft = readJSON<ProductEntryDraft>(PRODUCT_ENTRY_FLOW_DRAFT_KEY) || {}
  const legacyBasicInfo = readProductEntryBasicInfo()

  return {
    basicInfo: {
      ...legacyBasicInfo,
      ...(draft.basicInfo || {})
    },
    targetMarket: draft.targetMarket || {},
    productDetails: draft.productDetails || {}
  }
}

export const writeProductEntryDraft = (partial: Partial<ProductEntryDraft>) => {
  const current = readProductEntryDraft()
  const next: ProductEntryDraft = {
    basicInfo: {
      ...current.basicInfo,
      ...(partial.basicInfo || {})
    },
    targetMarket: {
      ...current.targetMarket,
      ...(partial.targetMarket || {})
    },
    productDetails: {
      ...current.productDetails,
      ...(partial.productDetails || {})
    }
  }

  writeJSON(PRODUCT_ENTRY_FLOW_DRAFT_KEY, next)
}

export const buildCreateDramaDraftFromProductEntry = (draft = readProductEntryDraft()): CreateDramaRequest | null => {
  const title = typeof draft.basicInfo.title === 'string' ? draft.basicInfo.title.trim() : ''
  const category = typeof draft.basicInfo.category === 'string' ? draft.basicInfo.category.trim() : ''
  const brand = typeof draft.basicInfo.brand === 'string' ? draft.basicInfo.brand.trim() : ''
  const marketName =
    typeof draft.targetMarket.marketName === 'string' ? draft.targetMarket.marketName.trim() : ''
  const platform = typeof draft.targetMarket.platform === 'string' ? draft.targetMarket.platform.trim() : ''
  const details = draft.productDetails

  if (!title) {
    return null
  }

  const descriptionParts = [
    details.description,
    details.scenarios ? `使用场景: ${details.scenarios}` : '',
    details.specifications ? `商品规格: ${details.specifications}` : '',
    details.notes ? `注意事项: ${details.notes}` : '',
    details.hasSensitiveClaims ? '涉及敏感功效描述: 是' : '涉及敏感功效描述: 否'
  ]
    .map((item) => (typeof item === 'string' ? item.trim() : ''))
    .filter(Boolean)

  const marketingGoals = normalizeStringArray(draft.targetMarket.marketingGoals)
  const keywordTags = typeof details.keywords === 'string' ? details.keywords.trim() : ''
  const tags = [brand, platform, keywordTags].filter(Boolean).join(', ')

  return {
    title,
    description:
      descriptionParts.join('\n') || [category, brand, platform].filter(Boolean).join(' / ') || title,
    target_country: marketName ? [marketName] : [],
    material_composition: typeof details.material === 'string' ? details.material.trim() : '',
    marketing_selling_points: marketingGoals.length > 0 ? marketingGoals.join('、') : keywordTags,
    genre: category || undefined,
    tags: tags || undefined
  }
}
