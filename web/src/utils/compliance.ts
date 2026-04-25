import type { ComplianceResult, ComplianceRiskLevel, CreateDramaRequest } from '@/types/drama'

export interface ComplianceRiskMeta {
  key: ComplianceRiskLevel
  badge: string
  text: string
  range: string
  color: string
}

type ComplianceLocale = 'zh-CN' | 'en-US'

const normalizeComplianceLocale = (language?: string): ComplianceLocale =>
  String(language || '').toLowerCase().startsWith('en') ? 'en-US' : 'zh-CN'

const getDefaultCompliance = (language?: string): ComplianceResult => {
  const locale = normalizeComplianceLocale(language)
  return {
    score: 0,
    level: 'green',
    level_label: locale === 'en-US' ? 'Low' : '低',
    summary: locale === 'en-US' ? 'No compliance assessment result was returned' : '未获取到合规评估结果',
    non_compliance_points: [],
    rectification_suggestions: [],
    suggested_categories: []
  }
}

const ZH_RISK_META_MAP: Record<ComplianceRiskLevel, ComplianceRiskMeta> = {
  green: {
    key: 'green',
    badge: '低风险',
    text: '低',
    range: '0-29',
    color: '#22c55e'
  },
  yellow: {
    key: 'yellow',
    badge: '中等风险',
    text: '中',
    range: '30-59',
    color: '#f59e0b'
  },
  orange: {
    key: 'orange',
    badge: '高风险',
    text: '高',
    range: '60-79',
    color: '#f97316'
  },
  red: {
    key: 'red',
    badge: '禁止',
    text: '禁止',
    range: '>=80',
    color: '#ef4444'
  }
}

const EN_RISK_META_MAP: Record<ComplianceRiskLevel, ComplianceRiskMeta> = {
  green: {
    key: 'green',
    badge: 'Low Risk',
    text: 'Low',
    range: '0-29',
    color: '#22c55e'
  },
  yellow: {
    key: 'yellow',
    badge: 'Medium Risk',
    text: 'Medium',
    range: '30-59',
    color: '#f59e0b'
  },
  orange: {
    key: 'orange',
    badge: 'High Risk',
    text: 'High',
    range: '60-79',
    color: '#f97316'
  },
  red: {
    key: 'red',
    badge: 'Blocked',
    text: 'Blocked',
    range: '>=80',
    color: '#ef4444'
  }
}

const getRiskMetaMap = (language?: string) =>
  normalizeComplianceLocale(language) === 'en-US' ? EN_RISK_META_MAP : ZH_RISK_META_MAP

const clampScore = (score: number): number => {
  if (score < 0) return 0
  if (score > 100) return 100
  return score
}

const isRiskLevel = (value: unknown): value is ComplianceRiskLevel => {
  return value === 'green' || value === 'yellow' || value === 'orange' || value === 'red'
}

const getLevelByScore = (score: number): ComplianceRiskLevel => {
  if (score >= 80) return 'red'
  if (score >= 60) return 'orange'
  if (score >= 30) return 'yellow'
  return 'green'
}

const toStringList = (value: unknown): string[] => {
  if (!Array.isArray(value)) return []
  return value
    .map((item) => (typeof item === 'string' ? item.trim() : ''))
    .filter((item) => item.length > 0)
}

const hasChinese = (value: string): boolean => /[\u4e00-\u9fa5]/.test(value)
const hasEnglishLetters = (value: string): boolean => /[A-Za-z]/.test(value)

const cleanCategorySegment = (value: string): string =>
  value
    .replace(/[_|]+/g, ' ')
    .replace(/[()]/g, ' ')
    .replace(/[,&/]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()

const normalizeCategoryKey = (value: string): string =>
  cleanCategorySegment(value)
    .toLowerCase()
    .replace(/\s*&\s*/g, ' & ')
    .replace(/\s+/g, ' ')

const GENERIC_CATEGORY_KEYS = new Set([
  'other',
  'others',
  'other related categories',
  'other related category',
  'other categories',
  'miscellaneous',
  '其他',
  '其他类目',
  '其他分类',
  '其他相关类目'
])

const EXACT_CATEGORY_MAP: Record<string, string> = {
  'pet supplies > feeders & waterers': '宠物用品 > 喂食器与饮水器',
  'electronics > smart home > pet cameras & feeders': '电子产品 > 智能家居 > 宠物摄像头与喂食器',
  'electronics > smart home': '电子产品 > 智能家居',
  'pet supplies': '宠物用品',
  'home & kitchen': '家居厨房',
  'home kitchen': '家居厨房',
  'home decor': '家居装饰',
  'kitchen & dining': '厨房餐饮',
  'kitchen dining': '厨房餐饮',
  'kitchen & table linens': '厨房与桌布',
  'household supplies': '家居日用',
  'storage & organization': '收纳整理',
  'furniture': '家具',
  'patio, lawn & garden': '庭院、草坪与花园',
  'patio lawn garden': '庭院、草坪与花园',
  'health & personal care': '健康与个护',
  'health personal care': '健康与个护',
  'beauty & personal care': '美妆个护',
  'beauty personal care': '美妆个护',
  'sports & outdoors': '运动与户外',
  'sports outdoors': '运动与户外',
  'outdoor': '户外',
  'outdoors': '户外',
  'camping & hiking': '露营与徒步',
  'camping hiking': '露营与徒步',
  'outdoor recreation': '户外休闲',
  'camp furniture': '露营家具',
  'outdoor furniture': '户外家具',
  'outdoor chairs': '户外椅',
  'outdoor chair': '户外椅',
  'chairs': '椅子',
  'chair': '椅子',
  'tents': '帐篷',
  'toys & games': '玩具与游戏',
  'baby products': '母婴用品',
  'automotive': '汽车用品',
  'office products': '办公用品',
  'fashion': '时尚服饰',
  'clothing': '服装',
  'electronics': '电子产品',
  'smart home': '智能家居'
}

const SEGMENT_CATEGORY_MAP: Record<string, string> = {
  'pet supplies': '宠物用品',
  'feeder': '喂食器',
  'feeders': '喂食器',
  'feeders & waterers': '喂食器与饮水器',
  'feeders waterers': '喂食器与饮水器',
  'waterer': '饮水器',
  'waterers': '饮水器',
  'pet cameras & feeders': '宠物摄像头与喂食器',
  'electronics': '电子产品',
  'smart home': '智能家居',
  'home': '家居',
  'kitchen': '厨房',
  'home decor': '家居装饰',
  'home & kitchen': '家居厨房',
  'home kitchen': '家居厨房',
  'kitchen & dining': '厨房餐饮',
  'kitchen dining': '厨房餐饮',
  'household supplies': '家居日用',
  'storage & organization': '收纳整理',
  'furniture': '家具',
  'patio, lawn & garden': '庭院、草坪与花园',
  'patio lawn garden': '庭院、草坪与花园',
  'garden': '园艺',
  'health & personal care': '健康与个护',
  'health personal care': '健康与个护',
  'beauty & personal care': '美妆个护',
  'beauty personal care': '美妆个护',
  'sports & outdoors': '运动与户外',
  'sports outdoors': '运动与户外',
  'outdoor': '户外',
  'outdoors': '户外',
  'camping': '露营',
  'camping & hiking': '露营与徒步',
  'camping hiking': '露营与徒步',
  'outdoor recreation': '户外休闲',
  'camp furniture': '露营家具',
  'outdoor furniture': '户外家具',
  'outdoor chairs': '户外椅',
  'outdoor chair': '户外椅',
  'chair': '椅子',
  'chairs': '椅子',
  'tent': '帐篷',
  'tents': '帐篷',
  'toys & games': '玩具与游戏',
  'baby products': '母婴用品',
  'office products': '办公用品'
}

const CATEGORY_FRAGMENT_MAP: Record<string, string> = {
  'pet supplies': '宠物用品',
  'pet camera': '宠物摄像头',
  'pet cameras': '宠物摄像头',
  'pet cameras feeders': '宠物摄像头与喂食器',
  'smart home': '智能家居',
  'home kitchen': '家居厨房',
  'home decor': '家居装饰',
  'kitchen dining': '厨房餐饮',
  'table linens': '桌布',
  'household supplies': '家居日用',
  'storage organization': '收纳整理',
  'patio lawn garden': '庭院、草坪与花园',
  'health personal care': '健康与个护',
  'beauty personal care': '美妆个护',
  'sports outdoors': '运动与户外',
  'outdoor recreation': '户外休闲',
  'camping hiking': '露营与徒步',
  'camp furniture': '露营家具',
  'outdoor furniture': '户外家具',
  'outdoor chair': '户外椅',
  'outdoor chairs': '户外椅',
  'feeders waterers': '喂食器与饮水器',
  'toys games': '玩具与游戏',
  'baby products': '母婴用品',
  'office products': '办公用品'
}

const CATEGORY_TOKEN_MAP: Record<string, string> = {
  pet: '宠物',
  supplies: '用品',
  feeder: '喂食器',
  feeders: '喂食器',
  waterer: '饮水器',
  waterers: '饮水器',
  camera: '摄像头',
  cameras: '摄像头',
  electronics: '电子产品',
  electronic: '电子产品',
  smart: '智能',
  home: '家居',
  kitchen: '厨房',
  decor: '装饰',
  dining: '餐饮',
  table: '桌',
  linens: '布艺',
  household: '家居',
  storage: '收纳',
  organization: '整理',
  furniture: '家具',
  patio: '庭院',
  lawn: '草坪',
  garden: '花园',
  health: '健康',
  beauty: '美妆',
  personal: '个人',
  care: '护理',
  sports: '运动',
  sport: '运动',
  outdoor: '户外',
  outdoors: '户外',
  recreation: '休闲',
  camping: '露营',
  camp: '露营',
  hiking: '徒步',
  tent: '帐篷',
  tents: '帐篷',
  toys: '玩具',
  toy: '玩具',
  games: '游戏',
  game: '游戏',
  baby: '母婴',
  products: '用品',
  product: '用品',
  automotive: '汽车用品',
  office: '办公',
  fashion: '时尚',
  clothing: '服装',
  chair: '椅',
  chairs: '椅',
  furnitures: '家具'
}

const stripNoiseTokens = (value: string): string =>
  cleanCategorySegment(value)
    .split(' ')
    .filter((token) => {
      if (!token) return false
      if (/^[A-Za-z]$/.test(token)) return false
      if (/^[&/+-]+$/.test(token)) return false
      return true
    })
    .join(' ')
    .trim()

const isGenericCategorySegment = (value: string): boolean => {
  const key = normalizeCategoryKey(value)
  return GENERIC_CATEGORY_KEYS.has(key)
}

const translateCategoryByFragments = (raw: string): string => {
  const key = normalizeCategoryKey(raw)
  if (!key) return ''

  if (CATEGORY_FRAGMENT_MAP[key]) {
    return CATEGORY_FRAGMENT_MAP[key]
  }

  const tokens = key.split(' ').filter(Boolean)
  if (!tokens.length) {
    return ''
  }

  const translated: string[] = []

  for (let index = 0; index < tokens.length;) {
    let matched = ''
    let matchedLength = 0

    for (let size = Math.min(3, tokens.length - index); size >= 1; size -= 1) {
      const fragment = tokens.slice(index, index + size).join(' ')
      if (CATEGORY_FRAGMENT_MAP[fragment]) {
        matched = CATEGORY_FRAGMENT_MAP[fragment]
        matchedLength = size
        break
      }
    }

    if (matched) {
      translated.push(matched)
      index += matchedLength
      continue
    }

    const current = CATEGORY_TOKEN_MAP[tokens[index]]
    if (!current) {
      return ''
    }

    translated.push(current)
    index += 1
  }

  return translated.join('')
}

const localizeCategorySegment = (value: string): string => {
  const raw = stripNoiseTokens(value)
  if (!raw) return raw
  if (hasChinese(raw) && !hasEnglishLetters(raw)) return raw

  const key = normalizeCategoryKey(raw)
  const mapped = SEGMENT_CATEGORY_MAP[key] || EXACT_CATEGORY_MAP[key]
  if (mapped) {
    return mapped
  }

  if (isGenericCategorySegment(raw)) {
    return '其他相关类目'
  }

  const translated = translateCategoryByFragments(raw)
  if (translated && !hasEnglishLetters(translated)) {
    return translated
  }

  return '其他相关类目'
}

const normalizeEnglishSuggestedCategory = (value: string): string => {
  const raw = value.trim()
  if (!raw) return raw
  if (hasChinese(raw)) return raw

  if (raw.includes('>')) {
    return raw
      .split(/\s*>\s*/)
      .map((segment) => cleanCategorySegment(segment))
      .filter((segment) => segment.length > 0)
      .join(' > ')
  }

  return cleanCategorySegment(raw)
}

const localizeSuggestedCategory = (value: string, language?: string): string => {
  const locale = normalizeComplianceLocale(language)
  const raw = value.trim()
  if (!raw) return raw
  if (locale === 'en-US') return normalizeEnglishSuggestedCategory(raw)
  if (hasChinese(raw)) return raw

  const exact = EXACT_CATEGORY_MAP[normalizeCategoryKey(raw)]
  if (exact) {
    return exact
  }

  if (raw.includes('>')) {
    const segments = raw
      .split(/\s*>\s*/)
      .map((segment) => localizeCategorySegment(segment))
      .filter((segment) => segment.length > 0)

    const normalizedSegments: string[] = []
    for (const segment of segments) {
      const key = normalizeCategoryKey(segment)
      if (!key) {
        continue
      }
      if (isGenericCategorySegment(segment) && normalizedSegments.length > 0) {
        continue
      }
      const prev = normalizedSegments[normalizedSegments.length - 1]
      if (prev && normalizeCategoryKey(prev) === key) {
        continue
      }
      normalizedSegments.push(segment)
    }

    if (normalizedSegments.length > 0) {
      return normalizedSegments.join(' > ')
    }
    return '其他相关类目'
  }

  return localizeCategorySegment(raw)
}

export const localizeSuggestedCategories = (categories: string[], language?: string): string[] => {
  const result: string[] = []
  const seen = new Set<string>()

  for (const item of categories) {
    const localized = localizeSuggestedCategory(item, language)
    const key = localized.toLowerCase()
    if (!localized || seen.has(key)) {
      continue
    }
    seen.add(key)
    result.push(localized)
  }

  return result
}

export const normalizeComplianceResult = (value: unknown, language?: string): ComplianceResult | null => {
  if (!value) return null

  let raw: any = value
  if (typeof value === 'string') {
    try {
      raw = JSON.parse(value)
    } catch {
      return null
    }
  }

  if (!raw || typeof raw !== 'object') {
    return null
  }

  const defaults = getDefaultCompliance(language)
  const riskMetaMap = getRiskMetaMap(language)
  const parsedScore = Number(raw.score)
  const score = Number.isFinite(parsedScore) ? clampScore(Math.round(parsedScore)) : defaults.score
  const level = isRiskLevel(raw.level) ? raw.level : getLevelByScore(score)

  const summary = typeof raw.summary === 'string' && raw.summary.trim()
    ? raw.summary.trim()
    : defaults.summary

  const levelLabel = typeof raw.level_label === 'string' && raw.level_label.trim()
    ? raw.level_label.trim()
    : riskMetaMap[level].text

  return {
    score,
    level,
    level_label: levelLabel,
    summary,
    non_compliance_points: toStringList(raw.non_compliance_points),
    rectification_suggestions: toStringList(raw.rectification_suggestions),
    suggested_categories: localizeSuggestedCategories(toStringList(raw.suggested_categories), language)
  }
}

export const getComplianceRiskMeta = (compliance: ComplianceResult, language?: string): ComplianceRiskMeta => {
  const riskMetaMap = getRiskMetaMap(language)
  return riskMetaMap[compliance.level] || riskMetaMap.green
}

export const buildCreateDramaPayload = (form: CreateDramaRequest): CreateDramaRequest => {
  return {
    title: form.title.trim(),
    description: form.description.trim(),
    target_country: Array.isArray(form.target_country)
      ? form.target_country.map((item) => String(item).trim()).filter((item) => item.length > 0)
      : [],
    material_composition: form.material_composition?.trim() || '',
    marketing_selling_points: form.marketing_selling_points?.trim() || '',
    genre: form.genre,
    tags: form.tags
  }
}
