/**
 * 模块说明：数字丝路素材上传接口。
 * 业务场景：商品主图、检测报告、营销素材需要先上传成可复用 URL，再进入合规分析和内容生产。
 * 核心职责：封装 multipart 上传，返回后端生成的访问地址、文件元信息和可映射的素材类型。
 */
import request from '@/utils/request'
import type { AssetType } from '@/types/asset'

export interface UploadFileResponse {
  url: string
  filename: string
  size: number
  content_type: string
  category: string
  asset_type: AssetType
}

export interface UploadFileOptions {
  category?: string
}

const uploadMultipart = <T>(url: string, file: File, options: UploadFileOptions = {}) => {
  const formData = new FormData()
  formData.append('file', file)
  if (options.category) {
    formData.append('category', options.category)
  }

  return request.post<T>(url, formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const uploadAPI = {
  /**
   * 功能：上传商品或营销素材文件。
   * 参数：file 为用户选择的文件；options.category 用来区分 product-images、product-attachments、timeline-materials 等业务目录。
   * 返回：UploadFileResponse，包含可被商品草稿和素材库继续引用的 URL。
   */
  uploadFile(file: File, options: UploadFileOptions = {}) {
    return uploadMultipart<UploadFileResponse>('/upload/file', file, options)
  },

  uploadImage(file: File, options: UploadFileOptions = {}) {
    return uploadMultipart<UploadFileResponse>('/upload/image', file, options)
  },

  uploadVideo(file: File, options: UploadFileOptions = {}) {
    return uploadMultipart<UploadFileResponse>('/upload/file', file, { category: options.category || 'videos' })
  },

  uploadAudio(file: File, options: UploadFileOptions = {}) {
    return uploadMultipart<UploadFileResponse>('/upload/file', file, { category: options.category || 'audios' })
  },

  uploadMaterial(file: File, options: UploadFileOptions = {}) {
    return uploadMultipart<UploadFileResponse>('/upload/file', file, { category: options.category || 'materials' })
  }
}
