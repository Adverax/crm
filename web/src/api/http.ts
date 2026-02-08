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

class HttpClient {
  private baseUrl: string
  private token: string | null = null

  constructor(baseUrl = '') {
    this.baseUrl = baseUrl
  }

  setToken(token: string | null) {
    this.token = token
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
  }): Promise<T> {
    const url = this.buildUrl(path, options?.params)
    const init: RequestInit = {
      method,
      headers: this.buildHeaders(),
    }

    if (options?.body) {
      init.body = JSON.stringify(camelToSnake(options.body))
    }

    const response = await fetch(url, init)

    if (response.status === 204) {
      return undefined as T
    }

    const json = await response.json()

    if (!response.ok) {
      const apiError = snakeToCamel<{ error: ApiError }>(json)
      throw new HttpError(response.status, apiError.error)
    }

    return snakeToCamel<T>(json)
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
