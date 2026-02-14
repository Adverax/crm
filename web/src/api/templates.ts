import { http } from './http'

const BASE = '/api/v1/admin/templates'

export interface TemplateInfo {
  id: string
  label: string
  description: string
  status: 'available' | 'applied' | 'blocked'
  objects: number
  fields: number
}

export interface TemplateListResponse {
  data: TemplateInfo[]
}

export interface TemplateApplyResponse {
  data: {
    templateId: string
    message: string
  }
}

export const templatesApi = {
  list(): Promise<TemplateListResponse> {
    return http.get<TemplateListResponse>(BASE)
  },

  apply(templateId: string): Promise<TemplateApplyResponse> {
    return http.post<TemplateApplyResponse>(`${BASE}/${templateId}/apply`, {})
  },
}
