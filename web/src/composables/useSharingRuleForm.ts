import { reactive, computed } from 'vue'
import type {
  SharingRule,
  CreateSharingRuleRequest,
  UpdateSharingRuleRequest,
  RuleType,
  AccessLevel,
} from '@/types/security'

export interface SharingRuleFormState {
  objectId: string
  ruleType: RuleType
  sourceGroupId: string
  targetGroupId: string
  accessLevel: AccessLevel
  criteriaField: string
  criteriaOp: string
  criteriaValue: string
}

export interface SharingRuleFormErrors {
  objectId?: string
  sourceGroupId?: string
  targetGroupId?: string
  criteriaField?: string
  criteriaOp?: string
  criteriaValue?: string
}

function defaultState(): SharingRuleFormState {
  return {
    objectId: '',
    ruleType: 'owner_based',
    sourceGroupId: '',
    targetGroupId: '',
    accessLevel: 'read',
    criteriaField: '',
    criteriaOp: '',
    criteriaValue: '',
  }
}

export function useSharingRuleForm(existing?: SharingRule) {
  const state = reactive<SharingRuleFormState>(existing ? {
    objectId: existing.objectId,
    ruleType: existing.ruleType,
    sourceGroupId: existing.sourceGroupId,
    targetGroupId: existing.targetGroupId,
    accessLevel: existing.accessLevel,
    criteriaField: existing.criteriaField ?? '',
    criteriaOp: existing.criteriaOp ?? '',
    criteriaValue: existing.criteriaValue ?? '',
  } : defaultState())

  const errors = reactive<SharingRuleFormErrors>({})

  function validate(): boolean {
    errors.objectId = undefined
    errors.sourceGroupId = undefined
    errors.targetGroupId = undefined
    errors.criteriaField = undefined
    errors.criteriaOp = undefined
    errors.criteriaValue = undefined

    let valid = true

    if (!state.objectId) {
      errors.objectId = 'Объект обязателен'
      valid = false
    }

    if (!state.sourceGroupId) {
      errors.sourceGroupId = 'Группа-источник обязательна'
      valid = false
    }

    if (!state.targetGroupId) {
      errors.targetGroupId = 'Группа-получатель обязательна'
      valid = false
    }

    if (state.ruleType === 'criteria_based') {
      if (!state.criteriaField.trim()) {
        errors.criteriaField = 'Поле критерия обязательно'
        valid = false
      }
      if (!state.criteriaOp.trim()) {
        errors.criteriaOp = 'Оператор обязателен'
        valid = false
      }
      if (!state.criteriaValue.trim()) {
        errors.criteriaValue = 'Значение обязательно'
        valid = false
      }
    }

    return valid
  }

  const isValid = computed(() => {
    const base = !!state.objectId && !!state.sourceGroupId && !!state.targetGroupId
    if (state.ruleType === 'criteria_based') {
      return base && !!state.criteriaField.trim() && !!state.criteriaOp.trim() && !!state.criteriaValue.trim()
    }
    return base
  })

  function toCreateRequest(): CreateSharingRuleRequest {
    return {
      objectId: state.objectId,
      ruleType: state.ruleType,
      sourceGroupId: state.sourceGroupId,
      targetGroupId: state.targetGroupId,
      accessLevel: state.accessLevel,
      criteriaField: state.ruleType === 'criteria_based' ? state.criteriaField || null : null,
      criteriaOp: state.ruleType === 'criteria_based' ? state.criteriaOp || null : null,
      criteriaValue: state.ruleType === 'criteria_based' ? state.criteriaValue || null : null,
    }
  }

  function toUpdateRequest(): UpdateSharingRuleRequest {
    return {
      targetGroupId: state.targetGroupId,
      accessLevel: state.accessLevel,
      criteriaField: state.ruleType === 'criteria_based' ? state.criteriaField || null : null,
      criteriaOp: state.ruleType === 'criteria_based' ? state.criteriaOp || null : null,
      criteriaValue: state.ruleType === 'criteria_based' ? state.criteriaValue || null : null,
    }
  }

  function reset() {
    Object.assign(state, defaultState())
    errors.objectId = undefined
    errors.sourceGroupId = undefined
    errors.targetGroupId = undefined
    errors.criteriaField = undefined
    errors.criteriaOp = undefined
    errors.criteriaValue = undefined
  }

  function initFrom(rule: SharingRule) {
    state.objectId = rule.objectId
    state.ruleType = rule.ruleType
    state.sourceGroupId = rule.sourceGroupId
    state.targetGroupId = rule.targetGroupId
    state.accessLevel = rule.accessLevel
    state.criteriaField = rule.criteriaField ?? ''
    state.criteriaOp = rule.criteriaOp ?? ''
    state.criteriaValue = rule.criteriaValue ?? ''
  }

  return { state, errors, validate, isValid, toCreateRequest, toUpdateRequest, reset, initFrom }
}
