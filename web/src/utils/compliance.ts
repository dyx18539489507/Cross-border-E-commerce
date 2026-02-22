import type { ComplianceResult, ComplianceRiskLevel, CreateDramaRequest } from '@/types/drama'

export interface ComplianceRiskMeta {
  key: ComplianceRiskLevel
  badge: string
  text: string
  range: string
  color: string
}

const DEFAULT_COMPLIANCE: ComplianceResult = {
  score: 0,
  level: 'green',
  level_label: '低',
  summary: '未获取到合规评估结果',
  non_compliance_points: [],
  rectification_suggestions: [],
  suggested_categories: []
}

const RISK_META_MAP: Record<ComplianceRiskLevel, ComplianceRiskMeta> = {
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

const normalizeCategoryKey = (value: string): string =>
  value
    .toLowerCase()
    .replace(/[()]/g, '')
    .replace(/\s+/g, ' ')
    .trim()

const EXACT_CATEGORY_MAP: Record<string, string> = {
  'pet supplies > feeders & waterers': '宠物用品 > 喂食器与饮水器',
  'electronics > smart home > pet cameras & feeders': '电子产品 > 智能家居 > 宠物摄像头与喂食器',
  'electronics > smart home': '电子产品 > 智能家居',
  'pet supplies': '宠物用品',
  'home & kitchen': '家居厨房',
  'health & personal care': '健康与个护',
  'beauty & personal care': '美妆个护',
  'sports & outdoors': '运动与户外',
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
  'feeders & waterers': '喂食器与饮水器',
  'pet cameras & feeders': '宠物摄像头与喂食器',
  'electronics': '电子产品',
  'smart home': '智能家居',
  'home & kitchen': '家居厨房',
  'health & personal care': '健康与个护',
  'beauty & personal care': '美妆个护',
  'sports & outdoors': '运动与户外',
  'toys & games': '玩具与游戏',
  'baby products': '母婴用品',
  'office products': '办公用品'
}

const CATEGORY_TOKEN_REPLACEMENTS: Array<[RegExp, string]> = [
  [/\bpet supplies\b/gi, '宠物用品'],
  [/\bpet cameras?\b/gi, '宠物摄像头'],
  [/\bfeeders?\b/gi, '喂食器'],
  [/\bwaterers?\b/gi, '饮水器'],
  [/\belectronics?\b/gi, '电子产品'],
  [/\bsmart home\b/gi, '智能家居'],
  [/\bhome\b/gi, '家居'],
  [/\bkitchen\b/gi, '厨房'],
  [/\bhealth\b/gi, '健康'],
  [/\bpersonal care\b/gi, '个护'],
  [/\bbeauty\b/gi, '美妆'],
  [/\bsports?\b/gi, '运动'],
  [/\boutdoors?\b/gi, '户外'],
  [/\btoys?\b/gi, '玩具'],
  [/\bgames?\b/gi, '游戏'],
  [/\bbaby\b/gi, '母婴'],
  [/\boffice\b/gi, '办公'],
  [/\bfashion\b/gi, '时尚服饰'],
  [/\bclothing\b/gi, '服装'],
  [/\bautomotive\b/gi, '汽车用品'],
  [/\s*&\s*/g, '与']
]

const localizeCategorySegment = (value: string): string => {
  const raw = value.trim()
  if (!raw) return raw
  if (hasChinese(raw)) return raw

  const mapped = SEGMENT_CATEGORY_MAP[normalizeCategoryKey(raw)]
  if (mapped) {
    return mapped
  }

  let translated = raw
  for (const [pattern, replacement] of CATEGORY_TOKEN_REPLACEMENTS) {
    translated = translated.replace(pattern, replacement)
  }
  translated = translated.replace(/\s+/g, ' ').trim()

  if (!hasChinese(translated)) {
    return '其他相关类目'
  }
  return translated
}

const localizeSuggestedCategory = (value: string): string => {
  const raw = value.trim()
  if (!raw) return raw
  if (hasChinese(raw)) return raw

  const exact = EXACT_CATEGORY_MAP[normalizeCategoryKey(raw)]
  if (exact) {
    return exact
  }

  if (raw.includes('>')) {
    return raw
      .split(/\s*>\s*/)
      .map((segment) => localizeCategorySegment(segment))
      .join(' > ')
  }

  return localizeCategorySegment(raw)
}

export const localizeSuggestedCategories = (categories: string[]): string[] => {
  const result: string[] = []
  const seen = new Set<string>()

  for (const item of categories) {
    const localized = localizeSuggestedCategory(item)
    const key = localized.toLowerCase()
    if (!localized || seen.has(key)) {
      continue
    }
    seen.add(key)
    result.push(localized)
  }

  return result
}

export const normalizeComplianceResult = (value: unknown): ComplianceResult | null => {
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

  const parsedScore = Number(raw.score)
  const score = Number.isFinite(parsedScore) ? clampScore(Math.round(parsedScore)) : DEFAULT_COMPLIANCE.score
  const level = isRiskLevel(raw.level) ? raw.level : getLevelByScore(score)

  const summary = typeof raw.summary === 'string' && raw.summary.trim()
    ? raw.summary.trim()
    : DEFAULT_COMPLIANCE.summary

  const levelLabel = typeof raw.level_label === 'string' && raw.level_label.trim()
    ? raw.level_label.trim()
    : RISK_META_MAP[level].text

  return {
    score,
    level,
    level_label: levelLabel,
    summary,
    non_compliance_points: toStringList(raw.non_compliance_points),
    rectification_suggestions: toStringList(raw.rectification_suggestions),
    suggested_categories: localizeSuggestedCategories(toStringList(raw.suggested_categories))
  }
}

export const getComplianceRiskMeta = (compliance: ComplianceResult): ComplianceRiskMeta => {
  return RISK_META_MAP[compliance.level] || RISK_META_MAP.green
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
