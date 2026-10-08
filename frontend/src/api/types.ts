export interface ApiErrorDetail {
  code: string
  message: string
  details?: Record<string, any>
}

export interface ApiResponse<T = any> {
  ok: boolean
  data: T | null
  error: ApiErrorDetail | null
}

export class AppApiError extends Error {
  code: string
  details?: Record<string, any>
  statusCode?: number

  constructor(message: string, code = 'UNKNOWN_ERROR', details?: Record<string, any>, statusCode?: number) {
    super(message)
    this.name = 'AppApiError'
    this.code = code
    this.details = details
    this.statusCode = statusCode
  }
}

export interface HealthStatus {
  status: string
  services: {
    postgres: string
    redis: string
  }
}
