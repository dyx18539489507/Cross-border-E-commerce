import type { AxiosError, AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import axios from 'axios'
import { ElMessage } from 'element-plus'

interface CustomAxiosInstance extends Omit<AxiosInstance, 'get' | 'post' | 'put' | 'patch' | 'delete'> {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
}

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 600000, // 10分钟超时，匹配后端AI生成接口
  headers: {
    'Content-Type': 'application/json'
  }
}) as CustomAxiosInstance

// 开源版本 - 无需认证token
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
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
      // 不在这里显示错误提示，让业务代码自行处理
      return Promise.reject(new Error(res.error?.message || '请求失败'))
    }
    // 兼容直接返回数据的接口
    return res
  },
  (error: AxiosError<any>) => {
    // 不在拦截器中自动显示错误提示，让业务代码根据具体情况处理
    // 将后端错误信息转为标准 Error，便于业务层展示
    const serverMessage = error.response?.data?.error?.message || error.response?.data?.message
    if (serverMessage) {
      return Promise.reject(new Error(serverMessage))
    }
    return Promise.reject(error)
  }
)

export default request
