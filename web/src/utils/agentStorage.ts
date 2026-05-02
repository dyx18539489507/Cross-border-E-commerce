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

export const saveAgentInput = (input: AgentInput) => {
  writeJSON(SILKROAD_AGENT_INPUT_KEY, {
    ...input,
    schemaVersion: SILKROAD_AGENT_SCHEMA_VERSION,
    requestId: input.requestId || createAgentRequestId()
  })
}

export const readAgentResult = () => {
  const result = readJSON<AgentResult>(SILKROAD_AGENT_RESULT_KEY)
  if (!result) return null
  if (result.schemaVersion !== SILKROAD_AGENT_SCHEMA_VERSION) {
    sessionStorage.removeItem(SILKROAD_AGENT_RESULT_KEY)
    return null
  }
  return result
}

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
