import { reactive, computed } from 'vue'
import type { ObjectDefinition, CreateObjectRequest, UpdateObjectRequest, ObjectType } from '@/types/metadata'

const API_NAME_REGEX = /^[A-Za-z][A-Za-z0-9_]*$/

export interface ObjectFormState {
  apiName: string
  label: string
  pluralLabel: string
  description: string
  objectType: ObjectType
  isVisibleInSetup: boolean
  isCustomFieldsAllowed: boolean
  isDeleteableObject: boolean
  isCreateable: boolean
  isUpdateable: boolean
  isDeleteable: boolean
  isQueryable: boolean
  isSearchable: boolean
  hasActivities: boolean
  hasNotes: boolean
  hasHistoryTracking: boolean
  hasSharingRules: boolean
}

export interface ObjectFormErrors {
  apiName?: string
  label?: string
  pluralLabel?: string
}

function defaultState(): ObjectFormState {
  return {
    apiName: '',
    label: '',
    pluralLabel: '',
    description: '',
    objectType: 'custom',
    isVisibleInSetup: true,
    isCustomFieldsAllowed: true,
    isDeleteableObject: true,
    isCreateable: true,
    isUpdateable: true,
    isDeleteable: true,
    isQueryable: true,
    isSearchable: false,
    hasActivities: false,
    hasNotes: false,
    hasHistoryTracking: false,
    hasSharingRules: false,
  }
}

export function useObjectForm(existing?: ObjectDefinition) {
  const state = reactive<ObjectFormState>(existing ? {
    apiName: existing.apiName,
    label: existing.label,
    pluralLabel: existing.pluralLabel,
    description: existing.description,
    objectType: existing.objectType,
    isVisibleInSetup: existing.isVisibleInSetup,
    isCustomFieldsAllowed: existing.isCustomFieldsAllowed,
    isDeleteableObject: existing.isDeleteableObject,
    isCreateable: existing.isCreateable,
    isUpdateable: existing.isUpdateable,
    isDeleteable: existing.isDeleteable,
    isQueryable: existing.isQueryable,
    isSearchable: existing.isSearchable,
    hasActivities: existing.hasActivities,
    hasNotes: existing.hasNotes,
    hasHistoryTracking: existing.hasHistoryTracking,
    hasSharingRules: existing.hasSharingRules,
  } : defaultState())

  const errors = reactive<ObjectFormErrors>({})

  function validate(): boolean {
    errors.apiName = undefined
    errors.label = undefined
    errors.pluralLabel = undefined

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

    if (!state.pluralLabel.trim()) {
      errors.pluralLabel = 'Название во мн. числе обязательно'
      valid = false
    }

    return valid
  }

  const isValid = computed(() => {
    return state.apiName.trim().length >= 2
      && API_NAME_REGEX.test(state.apiName)
      && state.label.trim().length > 0
      && state.pluralLabel.trim().length > 0
  })

  function toCreateRequest(): CreateObjectRequest {
    return { ...state }
  }

  function toUpdateRequest(): UpdateObjectRequest {
    return {
      label: state.label,
      pluralLabel: state.pluralLabel,
      description: state.description,
      isVisibleInSetup: state.isVisibleInSetup,
      isCustomFieldsAllowed: state.isCustomFieldsAllowed,
      isDeleteableObject: state.isDeleteableObject,
      isCreateable: state.isCreateable,
      isUpdateable: state.isUpdateable,
      isDeleteable: state.isDeleteable,
      isQueryable: state.isQueryable,
      isSearchable: state.isSearchable,
      hasActivities: state.hasActivities,
      hasNotes: state.hasNotes,
      hasHistoryTracking: state.hasHistoryTracking,
      hasSharingRules: state.hasSharingRules,
    }
  }

  function reset() {
    Object.assign(state, defaultState())
    errors.apiName = undefined
    errors.label = undefined
    errors.pluralLabel = undefined
  }

  function initFrom(obj: ObjectDefinition) {
    state.apiName = obj.apiName
    state.label = obj.label
    state.pluralLabel = obj.pluralLabel
    state.description = obj.description
    state.objectType = obj.objectType
    state.isVisibleInSetup = obj.isVisibleInSetup
    state.isCustomFieldsAllowed = obj.isCustomFieldsAllowed
    state.isDeleteableObject = obj.isDeleteableObject
    state.isCreateable = obj.isCreateable
    state.isUpdateable = obj.isUpdateable
    state.isDeleteable = obj.isDeleteable
    state.isQueryable = obj.isQueryable
    state.isSearchable = obj.isSearchable
    state.hasActivities = obj.hasActivities
    state.hasNotes = obj.hasNotes
    state.hasHistoryTracking = obj.hasHistoryTracking
    state.hasSharingRules = obj.hasSharingRules
  }

  return { state, errors, validate, isValid, toCreateRequest, toUpdateRequest, reset, initFrom }
}
