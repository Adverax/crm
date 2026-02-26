import { http } from './http'
import type {
  ObjectNavItem,
  ObjectDescribe,
  RecordData,
  RecordPagination,
} from '@/types/records'

interface DescribeListResponse {
  data: ObjectNavItem[]
}

interface DescribeObjectResponse {
  data: ObjectDescribe
}

interface RecordListResponse {
  data: RecordData[]
  pagination: RecordPagination
}

interface RecordSingleResponse {
  data: RecordData
}

interface RecordCreateResponse {
  data: { id: string }
}

export interface DescribeOptions {
  formFactor?: string
  formMode?: string
}

export const recordsApi = {
  listObjects(): Promise<DescribeListResponse> {
    return http.get<DescribeListResponse>('/api/v1/describe')
  },

  describeObject(objectName: string, options?: DescribeOptions): Promise<DescribeObjectResponse> {
    const headers: Record<string, string> = {}
    if (options?.formFactor) {
      headers['X-Form-Factor'] = options.formFactor
    }
    if (options?.formMode) {
      headers['X-Form-Mode'] = options.formMode
    }
    return http.request<DescribeObjectResponse>('GET', `/api/v1/describe/${objectName}`, { headers })
  },

  listRecords(objectName: string, page = 1, perPage = 20): Promise<RecordListResponse> {
    return http.request<RecordListResponse>('GET', `/api/v1/records/${objectName}`, {
      params: { page, per_page: perPage },
      skipCaseConversion: true,
    })
  },

  getRecord(objectName: string, recordId: string): Promise<RecordSingleResponse> {
    return http.request<RecordSingleResponse>('GET', `/api/v1/records/${objectName}/${recordId}`, {
      skipCaseConversion: true,
    })
  },

  createRecord(objectName: string, data: RecordData): Promise<RecordCreateResponse> {
    return http.request<RecordCreateResponse>('POST', `/api/v1/records/${objectName}`, {
      body: data,
      skipCaseConversion: true,
    })
  },

  updateRecord(objectName: string, recordId: string, data: RecordData): Promise<void> {
    return http.request<void>('PUT', `/api/v1/records/${objectName}/${recordId}`, {
      body: data,
      skipCaseConversion: true,
    })
  },

  deleteRecord(objectName: string, recordId: string): Promise<void> {
    return http.request<void>('DELETE', `/api/v1/records/${objectName}/${recordId}`, {
      skipCaseConversion: true,
    })
  },
}
