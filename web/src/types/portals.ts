import type { components } from './openapi'
import type { CamelCaseKeys } from './camelcase'

// --- Derived from OpenAPI spec (single source of truth) ---

export type Portal = CamelCaseKeys<components['schemas']['ObjectView']>
export type PortalConfig = CamelCaseKeys<components['schemas']['ObjectViewConfig']>
export type PortalGate = CamelCaseKeys<components['schemas']['PortalGate']>
export type GateBodyStep = CamelCaseKeys<components['schemas']['GateBodyStep']>
export type GateLayout = CamelCaseKeys<components['schemas']['GateLayout']>
export type GateOutcome = CamelCaseKeys<components['schemas']['GateOutcome']>
export type PortalViewField = CamelCaseKeys<components['schemas']['OVViewField']>
export type PortalArg = CamelCaseKeys<components['schemas']['PortalArg']>
export type PortalArgRule = CamelCaseKeys<components['schemas']['PortalArgRule']>
export type CreatePortalRequest = CamelCaseKeys<components['schemas']['CreateObjectViewRequest']>
export type UpdatePortalRequest = CamelCaseKeys<components['schemas']['UpdateObjectViewRequest']>
export type GatePostRequest = CamelCaseKeys<components['schemas']['GatePostRequest']>
export type GateResponse = CamelCaseKeys<components['schemas']['GateResponse']>

export type FormDescribe = CamelCaseKeys<components['schemas']['FormDescribe']>
export type FormArg = CamelCaseKeys<components['schemas']['FormArg']>
export type FormQuery = CamelCaseKeys<components['schemas']['FormQuery']>
export type FormSection = CamelCaseKeys<components['schemas']['FormSection']>
export type FormAction = CamelCaseKeys<components['schemas']['FormAction']>
export type FormValidationRule = CamelCaseKeys<components['schemas']['FormValidationRule']>
export type FormRelatedList = CamelCaseKeys<components['schemas']['FormRelatedList']>
