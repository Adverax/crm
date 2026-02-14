import { http } from './http'
import type {
  ValidationRule,
  CreateValidationRuleRequest,
  UpdateValidationRuleRequest,
} from '@/types/validationRules'
import type { ApiResponse } from '@/types/metadata'

const BASE = '/api/v1/admin/metadata'

export const validationRulesApi = {
  list(objectId: string): Promise<ApiResponse<ValidationRule[]>> {
    return http.get<ApiResponse<ValidationRule[]>>(`${BASE}/objects/${objectId}/rules`)
  },

  get(objectId: string, ruleId: string): Promise<ApiResponse<ValidationRule>> {
    return http.get<ApiResponse<ValidationRule>>(`${BASE}/objects/${objectId}/rules/${ruleId}`)
  },

  create(objectId: string, data: CreateValidationRuleRequest): Promise<ApiResponse<ValidationRule>> {
    return http.post<ApiResponse<ValidationRule>>(`${BASE}/objects/${objectId}/rules`, data)
  },

  update(objectId: string, ruleId: string, data: UpdateValidationRuleRequest): Promise<ApiResponse<ValidationRule>> {
    return http.put<ApiResponse<ValidationRule>>(`${BASE}/objects/${objectId}/rules/${ruleId}`, data)
  },

  delete(objectId: string, ruleId: string): Promise<void> {
    return http.delete(`${BASE}/objects/${objectId}/rules/${ruleId}`)
  },
}
