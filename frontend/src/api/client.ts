import axios, { type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios';
import { AppApiError, type ApiResponse } from './types';


export const apiClient: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 10_000,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (!config.headers['X-Request-ID']) {
      config.headers['X-Request-ID'] = crypto.randomUUID()
    }
    return config
  },

  (error) => {
    return Promise.reject(error)
  }
)

apiClient.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const payload = response.data

    if (payload && payload.ok === false) {
      const err = payload.error
      throw new AppApiError(
        err?.message || "Произошла ошибка при обработке запроса.",
        err?.code || "SERVER_ERROR",
        err?.details,
        response.status
      )
    }

    return payload?.data ?? payload
  },
  (error) => {
    if (error.response) {
      const errPayload = error.response.data as ApiResponse | undefined
      const errDetail = errPayload?.error

      throw new AppApiError(
        errDetail?.message || error.message || "Ошибка сервера",
        errDetail?.code || `HTTP_${error.response.status}`,
        errDetail?.details,
        error.response.status
      )
    } else if (error.request) {
      throw new AppApiError("Сервер недоступен. Проверьте соединение с бэкендом", "NETWORK_ERROR")
    }

    throw new AppApiError(error.message || "Неизвестная ошибка клиента", "UNKNOWN_ERROR")
  }
)

export const api = {
  get: <T>(url: string) => apiClient.get<T>(url) as Promise<T>,
  post: <T>(url: string, body?: any) => apiClient.post<T>(url, body) as Promise<T>,
}
