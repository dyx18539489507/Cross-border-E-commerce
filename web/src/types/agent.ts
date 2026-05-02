export const SILKROAD_AGENT_INPUT_KEY = 'silkroad_agent_input'
export const SILKROAD_AGENT_RESULT_KEY = 'silkroad_agent_result'
export const SILKROAD_AGENT_SCHEMA_VERSION = 2

export interface AgentInput {
  schemaVersion?: number
  requestId?: string
  productName?: string
  category?: string
  targetMarket?: string
  targetPlatform?: string
  targetAudience?: string
  coreSellingPoints?: string[]
  materialSpec?: string
  usageScenario?: string
  rawPrompt?: string
  imageDataUrl?: string
}

export interface AgentResult {
  schemaVersion?: number
  requestId?: string
  recognizedInfo: {
    productName: string
    category: string
    targetMarket: string
    targetPlatform: string
    targetAudience: string
    coreSellingPoints: string[]
    imageUnderstanding: string
  }
  overview: {
    complianceRiskLevel: string
    marketStrategy: string
    recommendedVideoStyle: string
    recommendedDigitalHuman: string
  }
  compliance: {
    title: string
    summary: string
    riskTags: string[]
    missingInfo: string[]
    suggestions: string[]
    forbiddenExpressions: string[]
    saferExpressions: string[]
  }
  localization: {
    direction: string
    reason: string
    keywords: string[]
    tone: string
    sceneSuggestions: string[]
  }
  script: {
    title: string
    duration: string
    opening: AgentScriptSegment
    middle: AgentScriptSegment
    ending: AgentScriptSegment
    storyboard: AgentStoryboardShot[]
  }
  digitalHuman: {
    persona: string
    tone: string
    videoRatio: string
    subtitleAdvice: string
    visualStyle: string
    shootingStyle: string
  }
  promotion: {
    platforms: string[]
    contentTags: string[]
    focusMetrics: string[]
    optimizationAdvice: string
  }
  agentMessage: {
    summary: string
    missingInfoNotice: string
    quickActions: string[]
  }
  errorMessage?: string
  isMock?: boolean
  model?: string
}

export interface AgentScriptSegment {
  time: string
  content: string
}

export interface AgentStoryboardShot {
  shot: string
  visual: string
  voiceover: string
  subtitle: string
}
