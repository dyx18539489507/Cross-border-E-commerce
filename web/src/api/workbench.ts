import request from '@/utils/request'
import type { WorkbenchSummary } from '@/types/workbench'

export const workbenchAPI = {
  summary() {
    return request.get<WorkbenchSummary>('/workbench/summary')
  }
}
