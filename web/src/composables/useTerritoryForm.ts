// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

import { reactive } from 'vue'
import type { Territory, CreateTerritoryRequest, UpdateTerritoryRequest } from '@/types/territory'

const API_NAME_REGEX = /^[A-Za-z][A-Za-z0-9_]*$/

export interface TerritoryFormState {
  modelId: string
  parentId: string | null
  apiName: string
  label: string
  description: string
}

export interface TerritoryFormErrors {
  apiName?: string
  label?: string
  modelId?: string
}

function defaultState(): TerritoryFormState {
  return {
    modelId: '',
    parentId: null,
    apiName: '',
    label: '',
    description: '',
  }
}

export function useTerritoryForm() {
  const state = reactive<TerritoryFormState>(defaultState())
  const errors = reactive<TerritoryFormErrors>({})

  function validate(): boolean {
    errors.apiName = undefined
    errors.label = undefined
    errors.modelId = undefined

    let valid = true

    if (!state.modelId) {
      errors.modelId = 'Модель обязательна'
      valid = false
    }

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

  function toCreateRequest(): CreateTerritoryRequest {
    return {
      modelId: state.modelId,
      parentId: state.parentId,
      apiName: state.apiName,
      label: state.label,
      description: state.description || undefined,
    }
  }

  function toUpdateRequest(): UpdateTerritoryRequest {
    return {
      parentId: state.parentId,
      label: state.label,
      description: state.description || undefined,
    }
  }

  function initFrom(territory: Territory) {
    state.modelId = territory.modelId
    state.parentId = territory.parentId
    state.apiName = territory.apiName
    state.label = territory.label
    state.description = territory.description
  }

  function reset() {
    Object.assign(state, defaultState())
    errors.apiName = undefined
    errors.label = undefined
    errors.modelId = undefined
  }

  return { state, errors, validate, toCreateRequest, toUpdateRequest, initFrom, reset }
}
