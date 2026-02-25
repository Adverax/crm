export type EventType =
  | 'before_insert'
  | 'after_insert'
  | 'before_update'
  | 'after_update'
  | 'before_delete'
  | 'after_delete'

export type ExecutionMode = 'per_record' | 'per_batch'

export interface AutomationRule {
  id: string
  objectId: string
  name: string
  description: string
  eventType: EventType
  condition: string | null
  procedureCode: string
  executionMode: ExecutionMode
  sortOrder: number
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateAutomationRuleRequest {
  name: string
  description?: string
  event_type: EventType
  condition?: string | null
  procedure_code: string
  execution_mode?: ExecutionMode
  sort_order?: number
  is_active?: boolean
}

export interface UpdateAutomationRuleRequest {
  name: string
  description?: string
  event_type: EventType
  condition?: string | null
  procedure_code: string
  execution_mode?: ExecutionMode
  sort_order?: number
  is_active?: boolean
}
