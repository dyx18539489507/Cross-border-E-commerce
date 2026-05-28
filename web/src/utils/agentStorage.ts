/**
 * 模块说明：丝路 Agent 浏览器会话缓存。
 * 业务场景：用户从首页启动 Agent 后，过渡页、结果页和商品录入页需要共享同一次分析上下文。
 * 核心职责：保存 Agent 输入、用户原话和生成结果，并用 schemaVersion 避免旧缓存污染新版页面。
 */
import type { AgentInput, AgentResult } from '@/types/agent'
import { SILKROAD_AGENT_INPUT_KEY, SILKROAD_AGENT_RESULT_KEY, SILKROAD_AGENT_SCHEMA_VERSION } from '@/types/agent'

export const SILKROAD_AGENT_USER_INPUT_KEY = 'agent_user_input'

const readJSON = <T>(key: string): T | null => {
  try {
    const raw = sessionStorage.getItem(key)
    if (!raw) return null
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}

const writeJSON = (key: string, value: unknown) => {
  sessionStorage.setItem(key, JSON.stringify(value))
}

export const readAgentInput = () => readJSON<AgentInput>(SILKROAD_AGENT_INPUT_KEY)

/**
 * 功能：保存本次 Agent 分析输入。
 * 参数：input 为首页或商品录入页整理出的商品、市场、平台、图片等上下文。
 * 返回：无返回值；写入 sessionStorage，供过渡页流式分析和结果页生成复用。
 */
export const saveAgentInput = (input: AgentInput) => {
  writeJSON(SILKROAD_AGENT_INPUT_KEY, {
    ...input,
    schemaVersion: SILKROAD_AGENT_SCHEMA_VERSION,
    requestId: input.requestId || createAgentRequestId()
  })
}

/**
 * 功能：读取已生成的 Agent 结果。
 * 参数：无。
 * 返回：结构版本匹配时返回 AgentResult；旧版本缓存会被清理并返回 null。
 */
export const readAgentResult = () => {
  const result = readJSON<AgentResult>(SILKROAD_AGENT_RESULT_KEY)
  if (!result) return null
  if (result.schemaVersion !== SILKROAD_AGENT_SCHEMA_VERSION) {
    sessionStorage.removeItem(SILKROAD_AGENT_RESULT_KEY)
    return null
  }
  return result
}

/**
 * 功能：保存 Agent 完整方案。
 * 参数：result 为后端或本地兜底生成的结构化方案。
 * 返回：无返回值；结果会和当前 requestId 绑定，方便结果页与商品草稿继续衔接。
 */
export const saveAgentResult = (result: AgentResult) => {
  const input = readAgentInput()
  writeJSON(SILKROAD_AGENT_RESULT_KEY, {
    ...result,
    schemaVersion: SILKROAD_AGENT_SCHEMA_VERSION,
    requestId: input?.requestId || createAgentRequestId()
  })
}

export const clearAgentResult = () => {
  sessionStorage.removeItem(SILKROAD_AGENT_RESULT_KEY)
}

/**
 * 功能：保存用户原始输入文本。
 * 参数：value 为首页文本框里的自然语言需求。
 * 返回：无返回值；用于过渡页在结构化字段不足时仍能展示用户真实诉求。
 */
export const saveAgentUserInput = (value: string) => {
  sessionStorage.setItem(SILKROAD_AGENT_USER_INPUT_KEY, value)
}

export const readAgentUserInput = () => {
  return sessionStorage.getItem(SILKROAD_AGENT_USER_INPUT_KEY) || ''
}

const createAgentRequestId = () => {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}
