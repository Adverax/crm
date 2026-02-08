import { reactive, computed } from 'vue'
import type { PermissionSet, CreatePermissionSetRequest, UpdatePermissionSetRequest, PsType } from '@/types/security'

const API_NAME_REGEX = /^[A-Za-z][A-Za-z0-9_]*$/

export interface PermissionSetFormState {
  apiName: string
  label: string
  description: string
  psType: PsType
}

export interface PermissionSetFormErrors {
  apiName?: string
  label?: string
  psType?: string
}

function defaultState(): PermissionSetFormState {
  return {
    apiName: '',
    label: '',
    description: '',
    psType: 'grant',
  }
}

export function usePermissionSetForm(existing?: PermissionSet) {
  const state = reactive<PermissionSetFormState>(existing ? {
    apiName: existing.apiName,
    label: existing.label,
    description: existing.description,
    psType: existing.psType,
  } : defaultState())

  const errors = reactive<PermissionSetFormErrors>({})

  function validate(): boolean {
    errors.apiName = undefined
    errors.label = undefined
    errors.psType = undefined

    let valid = true

    if (!state.apiName.trim()) {
      errors.apiName = 'API имя обязательно'
      valid = false
    } else if (state.apiName.length < 2) {
      errors.apiName = 'Минимум 2 символа'
      valid = false
    } else if (state.apiName.length > 100) {
      errors.apiName = 'Максимум 100 символов'
      valid = false
    } else if (!API_NAME_REGEX.test(state.apiName)) {
      errors.apiName = 'Только латинские буквы, цифры и подчёркивания. Начинается с буквы.'
      valid = false
    }

    if (!state.label.trim()) {
      errors.label = 'Название обязательно'
      valid = false
    }

    if (!state.psType) {
      errors.psType = 'Тип обязателен'
      valid = false
    }

    return valid
  }

  const isValid = computed(() => {
    return state.apiName.trim().length >= 2
      && API_NAME_REGEX.test(state.apiName)
      && state.label.trim().length > 0
      && !!state.psType
  })

  function toCreateRequest(): CreatePermissionSetRequest {
    return {
      apiName: state.apiName,
      label: state.label,
      description: state.description || undefined,
      psType: state.psType,
    }
  }

  function toUpdateRequest(): UpdatePermissionSetRequest {
    return {
      label: state.label,
      description: state.description || undefined,
    }
  }

  function reset() {
    Object.assign(state, defaultState())
    errors.apiName = undefined
    errors.label = undefined
    errors.psType = undefined
  }

  function initFrom(ps: PermissionSet) {
    state.apiName = ps.apiName
    state.label = ps.label
    state.description = ps.description
    state.psType = ps.psType
  }

  return { state, errors, validate, isValid, toCreateRequest, toUpdateRequest, reset, initFrom }
}
