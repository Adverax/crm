import { http } from './http'
import type {
  AutomationRule,
  CreateAutomationRuleRequest,
  UpdateAutomationRuleRequest,
} from '@/types/automation-rules'
import type { ApiResponse } from '@/types/metadata'

const BASE = '/api/v1/admin/metadata'

export const automationRulesApi = {
  list(objectId: string): Promise<ApiResponse<AutomationRule[]>> {
    return http.get<ApiResponse<AutomationRule[]>>(
      `${BASE}/objects/${objectId}/automation-rules`,
    )
  },

  get(id: string): Promise<ApiResponse<AutomationRule>> {
    return http.get<ApiResponse<AutomationRule>>(`${BASE}/automation-rules/${id}`)
  },

  create(
    objectId: string,
    data: CreateAutomationRuleRequest,
  ): Promise<ApiResponse<AutomationRule>> {
    return http.post<ApiResponse<AutomationRule>>(
      `${BASE}/objects/${objectId}/automation-rules`,
      data,
    )
  },

  update(id: string, data: UpdateAutomationRuleRequest): Promise<ApiResponse<AutomationRule>> {
    return http.put<ApiResponse<AutomationRule>>(`${BASE}/automation-rules/${id}`, data)
  },

  delete(id: string): Promise<void> {
    return http.delete(`${BASE}/automation-rules/${id}`)
  },
}
