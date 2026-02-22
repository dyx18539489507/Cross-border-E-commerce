import type { AxiosError, AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import axios from 'axios'

interface CustomAxiosInstance extends Omit<AxiosInstance, 'get' | 'post' | 'put' | 'patch' | 'delete'> {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
}

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 900000, // 15分钟超时，匹配较慢的AI/媒体接口
  headers: {
    'Content-Type': 'application/json'
  }
}) as CustomAxiosInstance

const DEVICE_ID_STORAGE_KEY = 'drama_device_id'
const DEVICE_ID_HEADER = 'X-Device-ID'

const generateDeviceID = (): string => {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return `dev_${crypto.randomUUID().replace(/-/g, '')}`
  }

  const timestamp = Date.now().toString(16)
  const random = `${Math.random().toString(16).slice(2)}${Math.random().toString(16).slice(2)}`
  return `dev_${timestamp}${random}`.slice(0, 48)
}

const getOrCreateDeviceID = (): string => {
  const existing = localStorage.getItem(DEVICE_ID_STORAGE_KEY)
  if (existing && /^[a-zA-Z0-9_-]{16,128}$/.test(existing)) {
    return existing
  }

  const created = generateDeviceID()
  localStorage.setItem(DEVICE_ID_STORAGE_KEY, created)
  return created
}

// 开源版本 - 无需认证token
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const deviceID = getOrCreateDeviceID()
    if (config.headers && typeof (config.headers as any).set === 'function') {
      ;(config.headers as any).set(DEVICE_ID_HEADER, deviceID)
    } else {
      ;(config.headers as Record<string, string>)[DEVICE_ID_HEADER] = deviceID
    }
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res && typeof res === 'object' && 'success' in res) {
      if (res.success) {
        return res.data
      }
      // 某些接口会错误地返回 success=false 但仍带有 data
      if (res.data !== undefined && res.data !== null) {
        return res.data
      }
      const wrappedError = new Error(res.error?.message || '请求失败') as Error & {
        code?: string
        details?: any
      }
      wrappedError.code = res.error?.code
      wrappedError.details = res.error?.details
      return Promise.reject(wrappedError)
    }
    return res
  },
  (error: AxiosError<any>) => {
    // 不在拦截器中自动显示错误提示，让业务代码根据具体情况处理
    const serverError = error.response?.data?.error
    const serverMessage = serverError?.message || error.response?.data?.message
    if (serverMessage) {
      const wrappedError = new Error(serverMessage) as Error & {
        code?: string
        details?: any
        status?: number
      }
      wrappedError.code = serverError?.code
      wrappedError.details = serverError?.details
      wrappedError.status = error.response?.status
      return Promise.reject(wrappedError)
    }
    return Promise.reject(error)
  }
)

export default request
