export type ProductEntrySource = 'manual' | 'agent'

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
  createdAt: string
  updatedAt: string
}

export interface ProductEntryDraft extends ProductContext {}
