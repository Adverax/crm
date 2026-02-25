export type VersionStatus = 'draft' | 'published' | 'superseded'

export type CommandType =
  | 'record.create' | 'record.update' | 'record.delete' | 'record.get' | 'record.query'
  | 'compute.transform' | 'compute.validate' | 'compute.fail'
  | 'flow.if' | 'flow.match' | 'flow.call' | 'flow.try'
  | 'integration.http'
  | 'notification.email' | 'notification.sms'
  | 'wait.delay' | 'wait.until'

export interface RetryConfig {
  maxAttempts: number
  delayMs: number
  backoffMult?: number
}

export interface CommandDef {
  type: CommandType | string
  as?: string
  optional?: boolean
  when?: string
  rollback?: CommandDef[]
  // record fields
  object?: string
  id?: string
  data?: Record<string, string>
  query?: string
  // compute fields
  value?: Record<string, string>
  condition?: string
  code?: string
  message?: string
  // flow fields
  then?: CommandDef[]
  else?: CommandDef[]
  cases?: Record<string, CommandDef[]>
  default?: CommandDef[]
  procedure?: string
  input?: Record<string, string>
  expression?: string
  // flow.try fields
  try?: CommandDef[]
  catch?: CommandDef[]
  // integration fields
  credential?: string
  method?: string
  path?: string
  headers?: Record<string, string>
  body?: string
  // retry
  retry?: RetryConfig
}

export interface ProcedureDefinition {
  commands: CommandDef[]
  result?: Record<string, string>
}

export interface ProcedureVersion {
  id: string
  procedureId: string
  version: number
  definition: ProcedureDefinition
  status: VersionStatus
  changeSummary: string
  createdBy: string | null
  createdAt: string
  publishedAt: string | null
}

export interface Procedure {
  id: string
  code: string
  name: string
  description: string
  draftVersionId: string | null
  publishedVersionId: string | null
  createdAt: string
  updatedAt: string
}

export interface ProcedureWithVersions {
  procedure: Procedure
  draftVersion: ProcedureVersion | null
  publishedVersion: ProcedureVersion | null
}

export interface CreateProcedureRequest {
  code: string
  name: string
  description?: string
}

export interface UpdateProcedureMetadataRequest {
  name: string
  description?: string
}

export interface SaveDraftRequest {
  definition: ProcedureDefinition
  changeSummary?: string
}

export interface ExecuteRequest {
  input?: Record<string, unknown>
}

export interface ExecutionResult {
  success: boolean
  result?: Record<string, unknown>
  warnings?: Array<{ command: string; message: string }>
  trace?: Array<{
    step: string
    type: string
    status: string
    durationMs: number
    error?: string
  }>
}
