/**
 * CEL Expression Editor — Type Definitions
 *
 * Standalone types for the CEL expression editor.
 * In the CRM project these are derived from OpenAPI spec via CamelCaseKeys<T>.
 * This file provides self-contained definitions for standalone usage.
 */

// --- CEL Context ---

/**
 * Evaluation context determines which variables are available in the CEL environment.
 *
 * - validation_rule: record, old, user, now — for field validation rules
 * - when_expression: record, old, user, now — for conditional expressions
 * - default_expr: record, user, now — for default field values
 * - function_body: only function parameters — for custom function bodies
 * - visibility_expr: field visibility conditions
 * - portal_when: args — for portal-level arg validation
 * - gate_when: args, datasets, data, user — for gate-level conditions
 */
export type CelContext =
  | 'validation_rule'
  | 'when_expression'
  | 'default_expr'
  | 'function_body'
  | 'visibility_expr'
  | 'portal_when'
  | 'gate_when'

// --- Functions ---

export interface FunctionParam {
  name: string
  type?: string       // 'string' | 'number' | 'boolean' | 'list' | 'map' | 'any'
  description?: string
}

export interface Function {
  id: string
  name?: string
  description?: string
  body?: string
  params?: FunctionParam[]
  returnType?: 'string' | 'number' | 'boolean' | 'list' | 'map' | 'any'
}

// --- Validation API ---

export interface CelParamDef {
  name: string
  type?: string
}

export interface CelValidateRequest {
  expression: string
  context: CelContext
  objectApiName?: string
  params?: CelParamDef[]
}

export interface CelValidateResponse {
  valid: boolean
  returnType?: string
  errors?: CelValidateError[]
}

export interface CelValidateError {
  message: string
  line?: number
  column?: number
}

// --- Field Info (for autocomplete and pickers) ---

export interface FieldInfo {
  apiName: string
  label: string
  fieldType: string   // 'text' | 'number' | 'boolean' | 'datetime' | 'reference' | etc.
}

// --- CEL Evaluation Result (client-side) ---

export interface CelEvalContext {
  record?: Record<string, unknown>
  old?: Record<string, unknown>
  user?: { id: string; profile_id: string; role_id: string }
  now?: Date
  [key: string]: unknown
}

export interface CelEvalResult {
  success: boolean
  value?: unknown
  type?: string       // 'null' | 'bool' | 'number' | 'int' | 'string' | 'list' | 'map' | 'unknown'
  error?: string
  position?: number   // cursor offset for error location
}
