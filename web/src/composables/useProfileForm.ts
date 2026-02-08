import { reactive, computed } from 'vue'
import type { Profile, CreateProfileRequest, UpdateProfileRequest } from '@/types/security'

const API_NAME_REGEX = /^[A-Za-z][A-Za-z0-9_]*$/

export interface ProfileFormState {
  apiName: string
  label: string
  description: string
}

export interface ProfileFormErrors {
  apiName?: string
  label?: string
}

function defaultState(): ProfileFormState {
  return {
    apiName: '',
    label: '',
    description: '',
  }
}

export function useProfileForm(existing?: Profile) {
  const state = reactive<ProfileFormState>(existing ? {
    apiName: existing.apiName,
    label: existing.label,
    description: existing.description,
  } : defaultState())

  const errors = reactive<ProfileFormErrors>({})

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

  function toCreateRequest(): CreateProfileRequest {
    return {
      apiName: state.apiName,
      label: state.label,
      description: state.description || undefined,
    }
  }

  function toUpdateRequest(): UpdateProfileRequest {
    return {
      label: state.label,
      description: state.description || undefined,
    }
  }

  function reset() {
    Object.assign(state, defaultState())
    errors.apiName = undefined
    errors.label = undefined
  }

  function initFrom(profile: Profile) {
    state.apiName = profile.apiName
    state.label = profile.label
    state.description = profile.description
  }

  return { state, errors, validate, isValid, toCreateRequest, toUpdateRequest, reset, initFrom }
}
