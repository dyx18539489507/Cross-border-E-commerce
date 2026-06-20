/**
 * 模块说明：数字丝路数字人生成接口封装。
 * 业务场景：Agent 或内容创作阶段需要把商品人物图、音频/文本口播生成数字人视频素材。
 * 核心职责：提交 multipart 表单给后端数字人接口，并约束返回的视频任务和视频地址结构。
 */
import request from '../utils/request'

export interface DigitalHumanResult {
  task_id: string
  local_task_id?: string
  upstream_task_id?: string
  video_url: string
  image_url?: string
  audio_url?: string
  speech_text?: string
  motion_text?: string
  mask_urls?: string[]
  subject_detected?: boolean
  marketing_use_case?: string
}

export type DigitalHumanTaskStatus = 'pending' | 'processing' | 'completed' | 'failed'

export interface DigitalHumanTask {
  id: string
  type: string
  status: DigitalHumanTaskStatus | string
  progress: number
  message?: string
  error?: string
  resource_id?: string
  result?: DigitalHumanResult
  video_url?: string
  task_id?: string
  upstream_task_id?: string
  created_at: string
  updated_at: string
  completed_at?: string
}

export interface DigitalHumanTaskListResponse {
  items: DigitalHumanTask[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export const digitalHumanAPI = {
  /**
   * 功能：发起数字人视频生成。
   * 参数：formData 包含角色图片、可选音频、口播文本、音色和动作提示。
   * 返回：DigitalHumanResult，包含任务 ID、视频地址和主体识别状态。
   */
  generate(formData: FormData) {
    return request.post<DigitalHumanResult>('/digital-humans', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  createDigitalHumanTask(formData: FormData) {
    return request.post<DigitalHumanResult>('/digital-humans', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  getDigitalHumanTasks(params?: { page?: number; page_size?: number; status?: string }) {
    return request.get<DigitalHumanTaskListResponse>('/digital-humans', { params })
  },

  getDigitalHumanHistory(params?: { page?: number; page_size?: number; status?: string }) {
    return request.get<DigitalHumanTaskListResponse>('/digital-humans/history', { params })
  },

  getDigitalHumanTaskDetail(id: number | string) {
    return request.get<DigitalHumanTask>(`/digital-humans/${id}`)
  },

  getDigitalHumanTaskStatus(id: number | string) {
    return request.get<Pick<DigitalHumanTask, 'id' | 'status' | 'progress' | 'message' | 'error' | 'video_url' | 'task_id' | 'upstream_task_id' | 'updated_at'>>(`/digital-humans/${id}/status`)
  },

  getDigitalHumanResult(id: number | string) {
    return request.get<DigitalHumanResult | { id: string; status: string; progress: number; error?: string }>(`/digital-humans/${id}/result`)
  },

  deleteDigitalHumanTask(id: number | string) {
    return request.delete(`/digital-humans/${id}`)
  }
}
