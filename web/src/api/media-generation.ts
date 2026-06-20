import { imageAPI } from './image'
import { musicAPI } from './music'
import { sfxAPI } from './sfx'
import { videoAPI } from './video'

export const mediaGenerationAPI = {
  images: {
    listByProject(projectId?: string, pageSize = 50) {
      return imageAPI.listImages({ drama_id: projectId, page_size: pageSize })
    },
    generateForProject(input: { project_id: string; prompt: string; material_type?: string }) {
      return imageAPI.generateImage({
        drama_id: input.project_id,
        prompt: input.prompt,
        image_type: input.material_type
      })
    }
  },
  videos: {
    listByProject(projectId?: string, pageSize = 50) {
      return videoAPI.listVideos({ drama_id: projectId, page_size: pageSize })
    },
    generateForProject(input: { project_id: string; prompt: string; image_url?: string; aspect_ratio?: string; duration?: number }) {
      return videoAPI.generateVideo({
        drama_id: input.project_id,
        prompt: input.prompt,
        image_url: input.image_url,
        aspect_ratio: input.aspect_ratio,
        duration: input.duration
      })
    }
  },
  music: musicAPI,
  soundEffects: sfxAPI
}
