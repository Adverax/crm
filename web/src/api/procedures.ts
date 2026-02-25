import { http } from './http'
import type {
  Procedure,
  ProcedureWithVersions,
  ProcedureVersion,
  CreateProcedureRequest,
  UpdateProcedureMetadataRequest,
  SaveDraftRequest,
  ExecuteRequest,
  ExecutionResult,
} from '@/types/procedures'
import type { ApiResponse } from '@/types/metadata'

const BASE = '/api/v1/admin/procedures'

export const proceduresApi = {
  list(): Promise<ApiResponse<Procedure[]>> {
    return http.get<ApiResponse<Procedure[]>>(BASE)
  },

  get(id: string): Promise<ApiResponse<ProcedureWithVersions>> {
    return http.get<ApiResponse<ProcedureWithVersions>>(`${BASE}/${id}`)
  },

  create(data: CreateProcedureRequest): Promise<ApiResponse<ProcedureWithVersions>> {
    return http.post<ApiResponse<ProcedureWithVersions>>(BASE, data)
  },

  updateMetadata(id: string, data: UpdateProcedureMetadataRequest): Promise<ApiResponse<Procedure>> {
    return http.put<ApiResponse<Procedure>>(`${BASE}/${id}`, data)
  },

  delete(id: string): Promise<void> {
    return http.delete(`${BASE}/${id}`)
  },

  saveDraft(id: string, data: SaveDraftRequest): Promise<ApiResponse<ProcedureVersion>> {
    return http.put<ApiResponse<ProcedureVersion>>(`${BASE}/${id}/draft`, data)
  },

  discardDraft(id: string): Promise<void> {
    return http.delete(`${BASE}/${id}/draft`)
  },

  createDraftFromPublished(id: string): Promise<ApiResponse<ProcedureVersion>> {
    return http.post<ApiResponse<ProcedureVersion>>(`${BASE}/${id}/draft/from-published`, {})
  },

  publish(id: string): Promise<ApiResponse<ProcedureVersion>> {
    return http.post<ApiResponse<ProcedureVersion>>(`${BASE}/${id}/publish`, {})
  },

  rollback(id: string): Promise<ApiResponse<ProcedureVersion>> {
    return http.post<ApiResponse<ProcedureVersion>>(`${BASE}/${id}/rollback`, {})
  },

  listVersions(id: string): Promise<ApiResponse<ProcedureVersion[]>> {
    return http.get<ApiResponse<ProcedureVersion[]>>(`${BASE}/${id}/versions`)
  },

  execute(id: string, data?: ExecuteRequest): Promise<ApiResponse<ExecutionResult>> {
    return http.post<ApiResponse<ExecutionResult>>(`${BASE}/${id}/execute`, data ?? {})
  },

  dryRun(id: string, data?: ExecuteRequest): Promise<ApiResponse<ExecutionResult>> {
    return http.post<ApiResponse<ExecutionResult>>(`${BASE}/${id}/dry-run`, data ?? {})
  },
}
