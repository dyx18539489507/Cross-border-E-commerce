import type { AgentInput, AgentResult } from '@/types/agent'
import request from '@/utils/request'

export const agentAPI = {
  extract(data: AgentInput) {
    return request.post<AgentInput>('/agent/extract', data)
  },
  generate(data: AgentInput) {
    return request.post<AgentResult>('/agent/generate', data)
  }
}
