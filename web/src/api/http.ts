import { snakeToCamel, camelToSnake } from '@/lib/case'
import type { ApiError } from '@/types/metadata'

export class HttpError extends Error {
  constructor(
    public status: number,
    public apiError: ApiError,
  ) {
    super(apiError.message)
    this.name = 'HttpError'
  }
}

type RefreshHandler = () => Promise<boolean>

class HttpClient {
  private baseUrl: string
  private token: string | null = null
  private refreshHandler: RefreshHandler | null = null
  private isRefreshing = false

  constructor(baseUrl = '') {
    this.baseUrl = baseUrl
  }

  setToken(token: string | null) {
    this.token = token
  }

  setRefreshHandler(handler: RefreshHandler) {
    this.refreshHandler = handler
  }

  private buildHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }
    return headers
  }

  private buildUrl(path: string, params?: Record<string, string | number | undefined>): string {
    const url = new URL(`${this.baseUrl}${path}`, window.location.origin)
    if (params) {
      for (const [key, value] of Object.entries(params)) {
        if (value !== undefined) {
          url.searchParams.set(key, String(value))
        }
      }
    }
    return url.toString()
  }

  async request<T>(method: string, path: string, options?: {
    body?: unknown
    params?: Record<string, string | number | undefined>
    skipCaseConversion?: boolean
  }): Promise<T> {
    const url = this.buildUrl(path, options?.params)
    const skip = options?.skipCaseConversion ?? false
    const init: RequestInit = {
      method,
      headers: this.buildHeaders(),
    }

    if (options?.body) {
      init.body = JSON.stringify(skip ? options.body : camelToSnake(options.body))
    }

    const response = await fetch(url, init)

    if (response.status === 204) {
      return undefined as T
    }

    const text = await response.text()
    if (!text) {
      if (!response.ok) {
        throw new HttpError(response.status, {
          code: 'INTERNAL',
          message: `${method} ${path} → ${response.status}: empty response from server`,
        })
      }
      return undefined as T
    }

    let json: unknown
    try {
      json = JSON.parse(text)
    } catch {
      throw new HttpError(response.status, {
        code: 'INTERNAL',
        message: `${method} ${path} → ${response.status}: invalid JSON`,
      })
    }

    if (!response.ok) {
      if (response.status === 401 && this.refreshHandler && !this.isRefreshing) {
        this.isRefreshing = true
        try {
          const refreshed = await this.refreshHandler()
          if (refreshed) {
            // Retry the original request with new token
            const retryInit: RequestInit = {
              method,
              headers: this.buildHeaders(),
            }
            if (options?.body) {
              retryInit.body = JSON.stringify(skip ? options.body : camelToSnake(options.body))
            }
            const retryResponse = await fetch(url, retryInit)
            if (retryResponse.status === 204) return undefined as T
            const retryText = await retryResponse.text()
            if (!retryText) {
              if (!retryResponse.ok) {
                throw new HttpError(retryResponse.status, { code: 'INTERNAL', message: 'empty response from server' })
              }
              return undefined as T
            }
            const retryJson = JSON.parse(retryText)
            if (!retryResponse.ok) {
              const retryError = snakeToCamel<{ error: ApiError }>(retryJson)
              throw new HttpError(retryResponse.status, retryError.error)
            }
            return skip ? (retryJson as T) : snakeToCamel<T>(retryJson)
          }
        } finally {
          this.isRefreshing = false
        }
      }
      const apiError = snakeToCamel<{ error: ApiError }>(json)
      throw new HttpError(response.status, apiError.error)
    }

    return skip ? (json as T) : snakeToCamel<T>(json)
  }

  get<T>(path: string, params?: Record<string, string | number | undefined>): Promise<T> {
    return this.request<T>('GET', path, { params })
  }

  post<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>('POST', path, { body })
  }

  put<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>('PUT', path, { body })
  }

  delete<T>(path: string): Promise<T> {
    return this.request<T>('DELETE', path)
  }
}

export const http = new HttpClient()
