import type { components } from './openapi'
import type { CamelCaseKeys } from './camelcase'

// --- Derived from OpenAPI spec (single source of truth) ---

export type Severity = components['schemas']['CreateValidationRuleRequest']['severity']
export type ValidationRule = CamelCaseKeys<components['schemas']['ValidationRule']>
export type CreateValidationRuleRequest = CamelCaseKeys<components['schemas']['CreateValidationRuleRequest']>
export type UpdateValidationRuleRequest = CamelCaseKeys<components['schemas']['UpdateValidationRuleRequest']>
