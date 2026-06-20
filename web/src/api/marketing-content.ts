import request from '@/utils/request'
import { projectAPI } from './project'

export const marketingContentAPI = {
  getScript: projectAPI.getScript,
  saveScript: projectAPI.saveScript,
  generateShotPlan(contentVersionId: string | number, model?: string) {
    return request.post(`/episodes/${contentVersionId}/storyboards`, { model })
  },
  updateShot(shotId: string | number, data: Record<string, unknown>) {
    return request.put(`/storyboards/${shotId}`, data)
  },
  deleteShot(shotId: string | number) {
    return request.delete(`/storyboards/${shotId}`)
  }
}
