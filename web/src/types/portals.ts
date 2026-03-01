import type { components } from './openapi'
import type { CamelCaseKeys } from './camelcase'

// --- Derived from OpenAPI spec (single source of truth) ---

export type Portal = CamelCaseKeys<components['schemas']['Portal']>
export type PortalConfig = CamelCaseKeys<components['schemas']['PortalConfig']>
export type PortalReadConfig = CamelCaseKeys<components['schemas']['PortalReadConfig']>
export type PortalAction = CamelCaseKeys<components['schemas']['PortalAction']>
export type PortalQuery = CamelCaseKeys<components['schemas']['PortalQuery']>
export type PortalViewField = CamelCaseKeys<components['schemas']['PortalViewField']>
export type PortalActionApply = CamelCaseKeys<components['schemas']['PortalActionApply']>
export type PortalScenarioRef = CamelCaseKeys<components['schemas']['PortalScenarioRef']>
export type PortalArg = CamelCaseKeys<components['schemas']['PortalArg']>
export type CreatePortalRequest = CamelCaseKeys<components['schemas']['CreatePortalRequest']>
export type UpdatePortalRequest = CamelCaseKeys<components['schemas']['UpdatePortalRequest']>

export type FormDescribe = CamelCaseKeys<components['schemas']['FormDescribe']>
export type FormArg = CamelCaseKeys<components['schemas']['FormArg']>
export type FormQuery = CamelCaseKeys<components['schemas']['FormQuery']>
export type FormSection = CamelCaseKeys<components['schemas']['FormSection']>
export type FormAction = CamelCaseKeys<components['schemas']['FormAction']>
export type FormValidationRule = CamelCaseKeys<components['schemas']['FormValidationRule']>
export type FormRelatedList = CamelCaseKeys<components['schemas']['FormRelatedList']>
