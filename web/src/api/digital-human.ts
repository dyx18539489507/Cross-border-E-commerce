import request from '../utils/request'

export interface DigitalHumanResult {
  task_id: string
  video_url: string
  mask_urls?: string[]
  subject_detected?: boolean
}

export const digitalHumanAPI = {
  generate(formData: FormData) {
    return request.post<DigitalHumanResult>('/digital-humans', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}
