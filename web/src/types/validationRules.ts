export type Severity = 'error' | 'warning'

export interface ValidationRule {
  id: string
  objectId: string
  apiName: string
  label: string
  description: string
  expression: string
  errorMessage: string
  errorCode: string
  severity: Severity
  whenExpression: string | null
  appliesTo: string
  sortOrder: number
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateValidationRuleRequest {
  apiName: string
  label: string
  expression: string
  errorMessage: string
  errorCode?: string
  severity?: Severity
  whenExpression?: string
  appliesTo?: string
  sortOrder?: number
  isActive?: boolean
  description?: string
}

export interface UpdateValidationRuleRequest {
  label: string
  expression: string
  errorMessage: string
  errorCode?: string
  severity?: Severity
  whenExpression?: string
  appliesTo?: string
  sortOrder?: number
  isActive?: boolean
  description?: string
}
