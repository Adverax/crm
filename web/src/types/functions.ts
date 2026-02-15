import type { components } from './openapi'
import type { CamelCaseKeys } from './camelcase'

// --- Derived from OpenAPI spec (single source of truth) ---

export type Function = CamelCaseKeys<components['schemas']['Function']>
export type FunctionParam = CamelCaseKeys<components['schemas']['FunctionParam']>
export type CreateFunctionRequest = CamelCaseKeys<components['schemas']['CreateFunctionRequest']>
export type UpdateFunctionRequest = CamelCaseKeys<components['schemas']['UpdateFunctionRequest']>

export type CelValidateRequest = CamelCaseKeys<components['schemas']['CelValidateRequest']>
export type CelParamDef = CamelCaseKeys<components['schemas']['CelParamDef']>
export type CelValidateResponse = CamelCaseKeys<components['schemas']['CelValidateResponse']>
export type CelValidateError = CamelCaseKeys<components['schemas']['CelValidateError']>

export type CelContext = components['schemas']['CelValidateRequest']['context']
export type FunctionReturnType = NonNullable<components['schemas']['Function']['return_type']>
export type FunctionParamType = components['schemas']['FunctionParam']['type']
