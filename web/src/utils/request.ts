/**
 * 模块说明：数字丝路前端请求封装。
 * 业务场景：Agent、商品录入、合规预检、数字人和分发接口都需要统一携带匿名设备标识并解析后端响应。
 * 核心职责：配置 API 基地址、超时时间、设备 ID 请求头、响应数据解包和业务错误标准化。
 */
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

/**
 * 功能：生成匿名设备 ID。
 * 参数：无。
 * 返回：用于后端隔离商品草稿、合规 token 和分发账号的浏览器设备标识。
 */
const generateDeviceID = (): string => {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return `dev_${crypto.randomUUID().replace(/-/g, '')}`
  }

  const timestamp = Date.now().toString(16)
  const random = `${Math.random().toString(16).slice(2)}${Math.random().toString(16).slice(2)}`
  return `dev_${timestamp}${random}`.slice(0, 48)
}

/**
 * 功能：读取或创建当前浏览器的匿名设备 ID。
 * 参数：无。
 * 返回：稳定写入 localStorage 的设备标识，供每个数字丝路 API 请求复用。
 */
const getOrCreateDeviceID = (): string => {
  const existing = localStorage.getItem(DEVICE_ID_STORAGE_KEY)
  if (existing && /^[a-zA-Z0-9_-]{16,128}$/.test(existing)) {
    return existing
  }

  const created = generateDeviceID()
  localStorage.setItem(DEVICE_ID_STORAGE_KEY, created)
  return created
}

export const getClientDeviceID = getOrCreateDeviceID

// 数字丝路演示版不要求登录，因此用设备 ID 在后端隔离合规预检、项目和分发目标。
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
        // 后端统一响应壳中的 data 才是业务数据，页面层不需要重复关心 success 字段。
        return res.data
      }
      // 某些接口会错误地返回 success=false 但仍带有 data
      if (res.data !== undefined && res.data !== null) {
        // 兼容旧接口的非标准响应，避免历史接口影响数字丝路页面主流程。
        return res.data
      }
      // 不在这里显示错误提示，让业务代码自行处理
      return Promise.reject(new Error(res.error?.message || '请求失败'))
    }
    // 兼容直接返回数据的接口
    return res
  },
  (error: AxiosError<any>) => {
    // 不在拦截器中自动显示错误提示，让业务代码根据具体情况处理
    // 将后端错误信息转为标准 Error，便于业务层展示
    const serverError = error.response?.data?.error
    const serverMessage = serverError?.message || error.response?.data?.message
    if (serverMessage) {
      // code/details/status 会被合规页、商品创建页用来区分“需重新预检”和“红色风险拦截”。
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
