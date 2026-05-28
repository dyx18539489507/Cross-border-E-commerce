/**
 * 模块说明：数字丝路数字人生成接口封装。
 * 业务场景：Agent 或内容创作阶段需要把商品人物图、音频/文本口播生成数字人视频素材。
 * 核心职责：提交 multipart 表单给后端数字人接口，并约束返回的视频任务和视频地址结构。
 */
import request from '../utils/request'

export interface DigitalHumanResult {
  task_id: string
  video_url: string
  mask_urls?: string[]
  subject_detected?: boolean
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
  }
}
