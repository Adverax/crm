import { reactive, computed } from 'vue'
import type { Group, CreateGroupRequest, GroupType } from '@/types/security'

const API_NAME_REGEX = /^[A-Za-z][A-Za-z0-9_]*$/

export interface GroupFormState {
  apiName: string
  label: string
  groupType: GroupType
}

export interface GroupFormErrors {
  apiName?: string
  label?: string
}

function defaultState(): GroupFormState {
  return {
    apiName: '',
    label: '',
    groupType: 'public',
  }
}

export function useGroupForm(existing?: Group) {
  const state = reactive<GroupFormState>(existing ? {
    apiName: existing.apiName,
    label: existing.label,
    groupType: existing.groupType,
  } : defaultState())

  const errors = reactive<GroupFormErrors>({})

  function validate(): boolean {
    errors.apiName = undefined
    errors.label = undefined

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

    return valid
  }

  const isValid = computed(() => {
    return state.apiName.trim().length >= 2
      && API_NAME_REGEX.test(state.apiName)
      && state.label.trim().length > 0
  })

  function toCreateRequest(): CreateGroupRequest {
    return {
      apiName: state.apiName,
      label: state.label,
      groupType: state.groupType,
    }
  }

  function reset() {
    Object.assign(state, defaultState())
    errors.apiName = undefined
    errors.label = undefined
  }

  function initFrom(group: Group) {
    state.apiName = group.apiName
    state.label = group.label
    state.groupType = group.groupType
  }

  return { state, errors, validate, isValid, toCreateRequest, reset, initFrom }
}
