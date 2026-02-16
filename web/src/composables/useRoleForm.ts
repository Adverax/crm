import { reactive, computed } from 'vue'
import type { UserRole, CreateRoleRequest, UpdateRoleRequest } from '@/types/security'

const API_NAME_REGEX = /^[A-Za-z][A-Za-z0-9_]*$/

export interface RoleFormState {
  apiName: string
  label: string
  description: string
  parentId: string | null
}

export interface RoleFormErrors {
  apiName?: string
  label?: string
}

function defaultState(): RoleFormState {
  return {
    apiName: '',
    label: '',
    description: '',
    parentId: null,
  }
}

export function useRoleForm(existing?: UserRole) {
  const state = reactive<RoleFormState>(existing ? {
    apiName: existing.apiName,
    label: existing.label,
    description: existing.description,
    parentId: existing.parentId,
  } : defaultState())

  const errors = reactive<RoleFormErrors>({})

  function validate(): boolean {
    errors.apiName = undefined
    errors.label = undefined

    let valid = true

    if (!state.apiName.trim()) {
      errors.apiName = 'API name is required'
      valid = false
    } else if (state.apiName.length < 2) {
      errors.apiName = 'Minimum 2 characters'
      valid = false
    } else if (state.apiName.length > 100) {
      errors.apiName = 'Maximum 100 characters'
      valid = false
    } else if (!API_NAME_REGEX.test(state.apiName)) {
      errors.apiName = 'Only letters, digits, and underscores. Must start with a letter.'
      valid = false
    }

    if (!state.label.trim()) {
      errors.label = 'Label is required'
      valid = false
    }

    return valid
  }

  const isValid = computed(() => {
    return state.apiName.trim().length >= 2
      && API_NAME_REGEX.test(state.apiName)
      && state.label.trim().length > 0
  })

  function toCreateRequest(): CreateRoleRequest {
    return {
      apiName: state.apiName,
      label: state.label,
      description: state.description || undefined,
      parentId: state.parentId,
    }
  }

  function toUpdateRequest(): UpdateRoleRequest {
    return {
      label: state.label,
      description: state.description || undefined,
      parentId: state.parentId,
    }
  }

  function reset() {
    Object.assign(state, defaultState())
    errors.apiName = undefined
    errors.label = undefined
  }

  function initFrom(role: UserRole) {
    state.apiName = role.apiName
    state.label = role.label
    state.description = role.description
    state.parentId = role.parentId
  }

  return { state, errors, validate, isValid, toCreateRequest, toUpdateRequest, reset, initFrom }
}
