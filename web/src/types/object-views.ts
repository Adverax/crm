import type { components } from './openapi'
import type { CamelCaseKeys } from './camelcase'

// --- Derived from OpenAPI spec (single source of truth) ---

export type ObjectView = CamelCaseKeys<components['schemas']['ObjectView']>
export type ObjectViewConfig = CamelCaseKeys<components['schemas']['ObjectViewConfig']>
export type OVReadConfig = CamelCaseKeys<components['schemas']['OVReadConfig']>
export type OVWriteConfig = CamelCaseKeys<components['schemas']['OVWriteConfig']>
export type OVAction = CamelCaseKeys<components['schemas']['OVAction']>
export type OVQuery = CamelCaseKeys<components['schemas']['OVQuery']>
export type OVMutation = CamelCaseKeys<components['schemas']['OVMutation']>
export type OVMutSync = CamelCaseKeys<components['schemas']['OVMutSync']>
export type OVReadComputed = CamelCaseKeys<components['schemas']['OVReadComputed']>
export type OVValidation = CamelCaseKeys<components['schemas']['OVValidation']>
export type OVDefault = CamelCaseKeys<components['schemas']['OVDefault']>
export type OVComputed = CamelCaseKeys<components['schemas']['OVComputed']>
export type CreateObjectViewRequest = CamelCaseKeys<components['schemas']['CreateObjectViewRequest']>
export type UpdateObjectViewRequest = CamelCaseKeys<components['schemas']['UpdateObjectViewRequest']>

export type FormDescribe = CamelCaseKeys<components['schemas']['FormDescribe']>
export type FormSection = CamelCaseKeys<components['schemas']['FormSection']>
export type FormAction = CamelCaseKeys<components['schemas']['FormAction']>
export type FormRelatedList = CamelCaseKeys<components['schemas']['FormRelatedList']>
