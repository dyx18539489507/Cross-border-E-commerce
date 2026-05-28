/**
 * 模块说明：数字丝路商品录入类型定义。
 * 业务场景：商品资料可能来自手动录入或丝路 Agent 预填，并继续流向合规分析、脚本生成和内容生产。
 * 核心职责：统一描述商品基础信息、目标市场、卖点、合规提示和本地化提示的业务字段。
 */
export type ProductEntrySource = 'manual' | 'agent'

export interface ProductAttachment {
  name: string
  url: string
  size: number
  mimeType: string
  type: string
  category?: string
}

// ProductContext 是跨步骤草稿的最小业务闭环，后续接口适配都从这里派生。
export interface ProductContext {
  source: ProductEntrySource
  productName: string
  category: string
  brand: string
  productImage: string
  description: string
  coreSellingPoints: string[]
  materialSpec: string
  usageScenario: string
  targetMarket: string
  targetPlatform: string
  targetAudience: string
  marketingGoal: string[]
  budgetPreference: string
  agentSummary?: string
  complianceHints?: string[]
  localizationHints?: string[]
  attachments: ProductAttachment[]
  createdAt: string
  updatedAt: string
}

export interface ProductEntryDraft extends ProductContext {}
