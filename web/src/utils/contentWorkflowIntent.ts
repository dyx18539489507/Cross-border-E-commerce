/**
 * 模块说明：数字丝路内容生产意图缓存。
 * 业务场景：Agent 结果页、商品录入页和内容创作页需要共享“下一步要生成什么内容”的业务意图。
 * 核心职责：把 Agent 输出的商品、市场、脚本、数字人和投放建议整理为内容创作页可直接消费的上下文。
 */
import type { AgentInput, AgentResult, AgentStoryboardShot } from '@/types/agent'

export type ContentWorkflowTarget = 'combo' | 'image' | 'video' | 'avatar' | 'promotion'
export type ContentWorkflowFormat = 'image' | 'video' | 'avatar'
export type ContentWorkflowSource = 'agent' | 'product-entry' | 'manual'

export interface ContentWorkflowIntent {
  source: ContentWorkflowSource
  target: ContentWorkflowTarget
  selectedMode: ContentWorkflowTarget
  selectedFormats: ContentWorkflowFormat[]
  productName: string
  category: string
  targetMarket: string
  targetPlatform: string
  targetAudience: string
  productImage: string
  sellingPoints: string[]
  scriptSummary: string
  storyboard: AgentStoryboardShot[]
  digitalHumanPersona: string
  digitalHumanTone: string
  digitalHumanStyle: string
  promotionPlatforms: string[]
  promotionTags: string[]
  promotionAdvice: string
  complianceRiskLevel: string
  requestId?: string
  createdAt: string
  updatedAt: string
}

export const CONTENT_WORKFLOW_INTENT_KEY = 'silkroad:content-workflow:intent'

const nowIso = () => new Date().toISOString()

const toText = (value: unknown) => (typeof value === 'string' ? value.trim() : '')

const toStringList = (value: unknown): string[] => {
  if (Array.isArray(value)) {
    return value.map((item) => String(item).trim()).filter(Boolean)
  }
  if (typeof value === 'string') {
    return value
      .split(/[，,、;；\n]/)
      .map((item) => item.trim())
      .filter(Boolean)
  }
  return []
}

const uniqueList = (values: unknown[]) => Array.from(new Set(values.flatMap((value) => toStringList(value))))

export const formatsForContentTarget = (target: ContentWorkflowTarget): ContentWorkflowFormat[] => {
  if (target === 'image') return ['image']
  if (target === 'video') return ['video']
  if (target === 'avatar') return ['avatar']
  if (target === 'promotion') return ['image', 'video']
  return ['image', 'video', 'avatar']
}

export const selectedModeForContentTarget = (target: ContentWorkflowTarget): ContentWorkflowTarget => {
  if (target === 'promotion') return 'combo'
  return target
}

/**
 * 功能：从 Agent 结构化结果创建内容生产意图。
 * 参数：result 为当前 Agent 方案；input 为用户启动 Agent 时的原始商品图片和补充字段；target 为用户想继续生产的内容类型。
 * 返回：ContentWorkflowIntent；包含内容创作页预选模式和专业编辑器可复用的脚本、数字人、投放素材线索。
 */
export const buildContentWorkflowIntentFromAgent = (
  result: AgentResult,
  input: AgentInput | null | undefined,
  target: ContentWorkflowTarget
): ContentWorkflowIntent => {
  const timestamp = nowIso()
  const recognized = (result.recognizedInfo || {}) as Partial<AgentResult['recognizedInfo']>
  const script = (result.script || {}) as Partial<AgentResult['script']>
  const digitalHuman = (result.digitalHuman || {}) as Partial<AgentResult['digitalHuman']>
  const promotion = (result.promotion || {}) as Partial<AgentResult['promotion']>
  const overview = (result.overview || {}) as Partial<AgentResult['overview']>

  const scriptSummary = [
    script.opening?.content,
    script.middle?.content,
    script.ending?.content
  ].map(toText).filter(Boolean).join('\n')

  return {
    source: 'agent',
    target,
    selectedMode: selectedModeForContentTarget(target),
    selectedFormats: formatsForContentTarget(target),
    productName: toText(recognized.productName) || toText(input?.productName),
    category: toText(recognized.category) || toText(input?.category),
    targetMarket: toText(recognized.targetMarket) || toText(input?.targetMarket),
    targetPlatform: toText(recognized.targetPlatform) || toText(input?.targetPlatform),
    targetAudience: toText(recognized.targetAudience) || toText(input?.targetAudience),
    productImage: toText(input?.imageDataUrl),
    sellingPoints: uniqueList([recognized.coreSellingPoints, input?.coreSellingPoints]),
    scriptSummary,
    storyboard: Array.isArray(script.storyboard) ? script.storyboard : [],
    digitalHumanPersona: toText(digitalHuman.persona) || toText(overview.recommendedDigitalHuman),
    digitalHumanTone: toText(digitalHuman.tone),
    digitalHumanStyle: toText(digitalHuman.visualStyle || digitalHuman.shootingStyle),
    promotionPlatforms: toStringList(promotion.platforms),
    promotionTags: toStringList(promotion.contentTags),
    promotionAdvice: toText(promotion.optimizationAdvice),
    complianceRiskLevel: toText(overview.complianceRiskLevel),
    requestId: result.requestId || input?.requestId,
    createdAt: timestamp,
    updatedAt: timestamp
  }
}

export const saveContentWorkflowIntent = (intent: ContentWorkflowIntent) => {
  if (typeof window === 'undefined') return
  window.sessionStorage.setItem(CONTENT_WORKFLOW_INTENT_KEY, JSON.stringify({
    ...intent,
    updatedAt: nowIso()
  }))
}

export const readContentWorkflowIntent = (): ContentWorkflowIntent | null => {
  if (typeof window === 'undefined') return null

  const raw = window.sessionStorage.getItem(CONTENT_WORKFLOW_INTENT_KEY)
  if (!raw) return null

  try {
    const parsed = JSON.parse(raw) as Partial<ContentWorkflowIntent>
    const target = parsed.target && ['combo', 'image', 'video', 'avatar', 'promotion'].includes(parsed.target)
      ? parsed.target
      : 'combo'

    return {
      source: parsed.source === 'agent' || parsed.source === 'product-entry' ? parsed.source : 'manual',
      target,
      selectedMode: selectedModeForContentTarget(target),
      selectedFormats: Array.isArray(parsed.selectedFormats) && parsed.selectedFormats.length
        ? parsed.selectedFormats.filter((item): item is ContentWorkflowFormat => ['image', 'video', 'avatar'].includes(item))
        : formatsForContentTarget(target),
      productName: toText(parsed.productName),
      category: toText(parsed.category),
      targetMarket: toText(parsed.targetMarket),
      targetPlatform: toText(parsed.targetPlatform),
      targetAudience: toText(parsed.targetAudience),
      productImage: toText(parsed.productImage),
      sellingPoints: toStringList(parsed.sellingPoints),
      scriptSummary: toText(parsed.scriptSummary),
      storyboard: Array.isArray(parsed.storyboard) ? parsed.storyboard as AgentStoryboardShot[] : [],
      digitalHumanPersona: toText(parsed.digitalHumanPersona),
      digitalHumanTone: toText(parsed.digitalHumanTone),
      digitalHumanStyle: toText(parsed.digitalHumanStyle),
      promotionPlatforms: toStringList(parsed.promotionPlatforms),
      promotionTags: toStringList(parsed.promotionTags),
      promotionAdvice: toText(parsed.promotionAdvice),
      complianceRiskLevel: toText(parsed.complianceRiskLevel),
      requestId: toText(parsed.requestId) || undefined,
      createdAt: toText(parsed.createdAt) || nowIso(),
      updatedAt: toText(parsed.updatedAt) || nowIso()
    }
  } catch {
    window.sessionStorage.removeItem(CONTENT_WORKFLOW_INTENT_KEY)
    return null
  }
}

export const clearContentWorkflowIntent = () => {
  if (typeof window === 'undefined') return
  window.sessionStorage.removeItem(CONTENT_WORKFLOW_INTENT_KEY)
}
