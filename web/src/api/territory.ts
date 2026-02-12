// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

import { http } from './http'
import type {
  TerritoryModel,
  CreateModelRequest,
  UpdateModelRequest,
  ModelFilter,
  Territory,
  CreateTerritoryRequest,
  UpdateTerritoryRequest,
  TerritoryFilter,
  TerritoryObjectDefault,
  SetObjectDefaultRequest,
  UserTerritoryAssignment,
  AssignUserRequest,
  AssignmentRule,
  CreateAssignmentRuleRequest,
  UpdateAssignmentRuleRequest,
  AssignmentRuleFilter,
} from '@/types/territory'
import type { ApiResponse, ApiListResponse } from '@/types/metadata'

const BASE = '/api/v1/admin/territory'

export const territoryApi = {
  // Models
  listModels(filter?: ModelFilter): Promise<ApiListResponse<TerritoryModel>> {
    const params: Record<string, string | number | undefined> = {}
    if (filter?.page) params['page'] = filter.page
    if (filter?.perPage) params['per_page'] = filter.perPage
    return http.get<ApiListResponse<TerritoryModel>>(`${BASE}/models`, params)
  },

  getModel(modelId: string): Promise<ApiResponse<TerritoryModel>> {
    return http.get<ApiResponse<TerritoryModel>>(`${BASE}/models/${modelId}`)
  },

  createModel(data: CreateModelRequest): Promise<ApiResponse<TerritoryModel>> {
    return http.post<ApiResponse<TerritoryModel>>(`${BASE}/models`, data)
  },

  updateModel(modelId: string, data: UpdateModelRequest): Promise<ApiResponse<TerritoryModel>> {
    return http.put<ApiResponse<TerritoryModel>>(`${BASE}/models/${modelId}`, data)
  },

  deleteModel(modelId: string): Promise<void> {
    return http.delete(`${BASE}/models/${modelId}`)
  },

  activateModel(modelId: string): Promise<void> {
    return http.post(`${BASE}/models/${modelId}/activate`, {})
  },

  archiveModel(modelId: string): Promise<void> {
    return http.post(`${BASE}/models/${modelId}/archive`, {})
  },

  // Territories
  listTerritories(filter: TerritoryFilter): Promise<ApiListResponse<Territory>> {
    const params: Record<string, string | number | undefined> = {
      model_id: filter.modelId,
    }
    if (filter.page) params['page'] = filter.page
    if (filter.perPage) params['per_page'] = filter.perPage
    return http.get<ApiListResponse<Territory>>(`${BASE}/territories`, params)
  },

  getTerritory(territoryId: string): Promise<ApiResponse<Territory>> {
    return http.get<ApiResponse<Territory>>(`${BASE}/territories/${territoryId}`)
  },

  createTerritory(data: CreateTerritoryRequest): Promise<ApiResponse<Territory>> {
    return http.post<ApiResponse<Territory>>(`${BASE}/territories`, data)
  },

  updateTerritory(territoryId: string, data: UpdateTerritoryRequest): Promise<ApiResponse<Territory>> {
    return http.put<ApiResponse<Territory>>(`${BASE}/territories/${territoryId}`, data)
  },

  deleteTerritory(territoryId: string): Promise<void> {
    return http.delete(`${BASE}/territories/${territoryId}`)
  },

  // Object Defaults
  listObjectDefaults(territoryId: string): Promise<ApiResponse<TerritoryObjectDefault[]>> {
    return http.get<ApiResponse<TerritoryObjectDefault[]>>(
      `${BASE}/territories/${territoryId}/object-defaults`,
    )
  },

  setObjectDefault(territoryId: string, data: SetObjectDefaultRequest): Promise<ApiResponse<TerritoryObjectDefault>> {
    return http.post<ApiResponse<TerritoryObjectDefault>>(
      `${BASE}/territories/${territoryId}/object-defaults`,
      data,
    )
  },

  removeObjectDefault(territoryId: string, objectId: string): Promise<void> {
    return http.delete(`${BASE}/territories/${territoryId}/object-defaults/${objectId}`)
  },

  // User Assignments
  listTerritoryUsers(territoryId: string): Promise<ApiResponse<UserTerritoryAssignment[]>> {
    return http.get<ApiResponse<UserTerritoryAssignment[]>>(
      `${BASE}/territories/${territoryId}/users`,
    )
  },

  assignUser(territoryId: string, data: AssignUserRequest): Promise<ApiResponse<UserTerritoryAssignment>> {
    return http.post<ApiResponse<UserTerritoryAssignment>>(
      `${BASE}/territories/${territoryId}/users`,
      data,
    )
  },

  unassignUser(territoryId: string, userId: string): Promise<void> {
    return http.delete(`${BASE}/territories/${territoryId}/users/${userId}`)
  },

  // Assignment Rules
  listAssignmentRules(filter?: AssignmentRuleFilter): Promise<ApiListResponse<AssignmentRule>> {
    const params: Record<string, string | number | undefined> = {}
    if (filter?.territoryId) params['territory_id'] = filter.territoryId
    if (filter?.page) params['page'] = filter.page
    if (filter?.perPage) params['per_page'] = filter.perPage
    return http.get<ApiListResponse<AssignmentRule>>(`${BASE}/assignment-rules`, params)
  },

  getAssignmentRule(ruleId: string): Promise<ApiResponse<AssignmentRule>> {
    return http.get<ApiResponse<AssignmentRule>>(`${BASE}/assignment-rules/${ruleId}`)
  },

  createAssignmentRule(data: CreateAssignmentRuleRequest): Promise<ApiResponse<AssignmentRule>> {
    return http.post<ApiResponse<AssignmentRule>>(`${BASE}/assignment-rules`, data)
  },

  updateAssignmentRule(ruleId: string, data: UpdateAssignmentRuleRequest): Promise<ApiResponse<AssignmentRule>> {
    return http.put<ApiResponse<AssignmentRule>>(`${BASE}/assignment-rules/${ruleId}`, data)
  },

  deleteAssignmentRule(ruleId: string): Promise<void> {
    return http.delete(`${BASE}/assignment-rules/${ruleId}`)
  },
}
