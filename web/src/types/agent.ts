/**
 * 模块说明：丝路 Agent 前端数据结构定义。
 * 业务场景：Agent 输入、流式过渡页、结果页、追问和商品录入预填都需要共享同一套字段语义。
 * 核心职责：约束浏览器会话缓存和接口返回结构，保证合规、本地化、脚本、数字人、投放模块可稳定渲染。
 */
export const SILKROAD_AGENT_INPUT_KEY = 'silkroad_agent_input'
export const SILKROAD_AGENT_RESULT_KEY = 'silkroad_agent_result'
export const SILKROAD_AGENT_SCHEMA_VERSION = 2

// schemaVersion 用来隔离历史缓存，避免旧结构的 Agent 结果被新版结果页误读。
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

// AgentResult 是结果页的业务合同：摘要、信息缺口和追问都会围绕这些模块做增量更新。
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
