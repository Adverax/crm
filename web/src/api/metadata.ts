import { http } from './http'
import type {
  ObjectDefinition,
  CreateObjectRequest,
  UpdateObjectRequest,
  FieldDefinition,
  CreateFieldRequest,
  UpdateFieldRequest,
  ApiResponse,
  ApiListResponse,
  ObjectFilter,
} from '@/types/metadata'

const BASE = '/api/v1/admin/metadata'

export const metadataApi = {
  listObjects(filter?: ObjectFilter): Promise<ApiListResponse<ObjectDefinition>> {
    const params: Record<string, string | number | undefined> = {}
    if (filter?.page) params['page'] = filter.page
    if (filter?.perPage) params['per_page'] = filter.perPage
    if (filter?.objectType) params['object_type'] = filter.objectType
    return http.get<ApiListResponse<ObjectDefinition>>(`${BASE}/objects`, params)
  },

  getObject(objectId: string): Promise<ApiResponse<ObjectDefinition>> {
    return http.get<ApiResponse<ObjectDefinition>>(`${BASE}/objects/${objectId}`)
  },

  createObject(data: CreateObjectRequest): Promise<ApiResponse<ObjectDefinition>> {
    return http.post<ApiResponse<ObjectDefinition>>(`${BASE}/objects`, data)
  },

  updateObject(objectId: string, data: UpdateObjectRequest): Promise<ApiResponse<ObjectDefinition>> {
    return http.put<ApiResponse<ObjectDefinition>>(`${BASE}/objects/${objectId}`, data)
  },

  deleteObject(objectId: string): Promise<void> {
    return http.delete(`${BASE}/objects/${objectId}`)
  },

  listFields(objectId: string): Promise<ApiListResponse<FieldDefinition>> {
    return http.get<ApiListResponse<FieldDefinition>>(`${BASE}/objects/${objectId}/fields`)
  },

  getField(objectId: string, fieldId: string): Promise<ApiResponse<FieldDefinition>> {
    return http.get<ApiResponse<FieldDefinition>>(`${BASE}/objects/${objectId}/fields/${fieldId}`)
  },

  createField(objectId: string, data: CreateFieldRequest): Promise<ApiResponse<FieldDefinition>> {
    return http.post<ApiResponse<FieldDefinition>>(`${BASE}/objects/${objectId}/fields`, data)
  },

  updateField(objectId: string, fieldId: string, data: UpdateFieldRequest): Promise<ApiResponse<FieldDefinition>> {
    return http.put<ApiResponse<FieldDefinition>>(`${BASE}/objects/${objectId}/fields/${fieldId}`, data)
  },

  deleteField(objectId: string, fieldId: string): Promise<void> {
    return http.delete(`${BASE}/objects/${objectId}/fields/${fieldId}`)
  },
}
