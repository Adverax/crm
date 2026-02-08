import { reactive, computed } from 'vue'
import type { User, CreateUserRequest, UpdateUserRequest } from '@/types/security'

const USERNAME_REGEX = /^[A-Za-z][A-Za-z0-9_.-]*$/
const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export interface UserFormState {
  username: string
  email: string
  firstName: string
  lastName: string
  profileId: string
  roleId: string | null
  isActive: boolean
}

export interface UserFormErrors {
  username?: string
  email?: string
  profileId?: string
}

function defaultState(): UserFormState {
  return {
    username: '',
    email: '',
    firstName: '',
    lastName: '',
    profileId: '',
    roleId: null,
    isActive: true,
  }
}

export function useUserForm(existing?: User) {
  const state = reactive<UserFormState>(existing ? {
    username: existing.username,
    email: existing.email,
    firstName: existing.firstName,
    lastName: existing.lastName,
    profileId: existing.profileId,
    roleId: existing.roleId,
    isActive: existing.isActive,
  } : defaultState())

  const errors = reactive<UserFormErrors>({})

  function validate(): boolean {
    errors.username = undefined
    errors.email = undefined
    errors.profileId = undefined

    let valid = true

    if (!state.username.trim()) {
      errors.username = 'Имя пользователя обязательно'
      valid = false
    } else if (state.username.length < 2) {
      errors.username = 'Минимум 2 символа'
      valid = false
    } else if (state.username.length > 100) {
      errors.username = 'Максимум 100 символов'
      valid = false
    } else if (!USERNAME_REGEX.test(state.username)) {
      errors.username = 'Только латинские буквы, цифры, точки, дефисы и подчёркивания'
      valid = false
    }

    if (!state.email.trim()) {
      errors.email = 'Email обязателен'
      valid = false
    } else if (!EMAIL_REGEX.test(state.email)) {
      errors.email = 'Некорректный формат email'
      valid = false
    }

    if (!state.profileId) {
      errors.profileId = 'Профиль обязателен'
      valid = false
    }

    return valid
  }

  const isValid = computed(() => {
    return state.username.trim().length >= 2
      && USERNAME_REGEX.test(state.username)
      && EMAIL_REGEX.test(state.email)
      && !!state.profileId
  })

  function toCreateRequest(): CreateUserRequest {
    return {
      username: state.username,
      email: state.email,
      firstName: state.firstName || undefined,
      lastName: state.lastName || undefined,
      profileId: state.profileId,
      roleId: state.roleId,
    }
  }

  function toUpdateRequest(): UpdateUserRequest {
    return {
      email: state.email,
      firstName: state.firstName || undefined,
      lastName: state.lastName || undefined,
      profileId: state.profileId,
      roleId: state.roleId,
      isActive: state.isActive,
    }
  }

  function reset() {
    Object.assign(state, defaultState())
    errors.username = undefined
    errors.email = undefined
    errors.profileId = undefined
  }

  function initFrom(user: User) {
    state.username = user.username
    state.email = user.email
    state.firstName = user.firstName
    state.lastName = user.lastName
    state.profileId = user.profileId
    state.roleId = user.roleId
    state.isActive = user.isActive
  }

  return { state, errors, validate, isValid, toCreateRequest, toUpdateRequest, reset, initFrom }
}
