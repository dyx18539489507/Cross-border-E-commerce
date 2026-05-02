import type { AgentInput, AgentResult } from '@/types/agent'
import request from '@/utils/request'

export const agentAPI = {
  generate(data: AgentInput) {
    return request.post<AgentResult>('/agent/generate', data)
  }
}
