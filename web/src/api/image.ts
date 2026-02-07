import type {
    GenerateImageRequest,
    ImageGeneration,
    ImageGenerationListParams
} from '../types/image'
import request from '../utils/request'

export const imageAPI = {
  generateImage(data: GenerateImageRequest) {
    return request.post<ImageGeneration>('/images', data)
  },

  createImageRecord(data: {
    drama_id: string
    storyboard_id?: number
    scene_id?: number
    frame_type?: string
    image_type?: string
    image_url: string
    prompt?: string
  }) {
    return request.post<ImageGeneration>('/images/manual', data)
  },

  generateForScene(sceneId: number) {
    return request.post<ImageGeneration[]>(`/images/scene/${sceneId}`)
  },

  batchGenerateForEpisode(episodeId: number) {
    return request.post<ImageGeneration[]>(`/images/episode/${episodeId}/batch`)
  },

  getImage(id: number) {
    return request.get<ImageGeneration>(`/images/${id}`)
  },

  listImages(params: ImageGenerationListParams) {
    return request.get<{
      items: ImageGeneration[]
      pagination: {
        page: number
        page_size: number
        total: number
        total_pages: number
      }
    }>('/images', { params })
  },

  deleteImage(id: number) {
    return request.delete(`/images/${id}`)
  }
}
