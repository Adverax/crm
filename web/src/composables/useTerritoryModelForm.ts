// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

import { reactive } from 'vue'
import type { TerritoryModel, CreateModelRequest, UpdateModelRequest } from '@/types/territory'

const API_NAME_REGEX = /^[A-Za-z][A-Za-z0-9_]*$/

export interface ModelFormState {
  apiName: string
  label: string
  description: string
}

export interface ModelFormErrors {
  apiName?: string
  label?: string
}

function defaultState(): ModelFormState {
  return {
    apiName: '',
    label: '',
    description: '',
  }
}

export function useTerritoryModelForm() {
  const state = reactive<ModelFormState>(defaultState())
  const errors = reactive<ModelFormErrors>({})

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

  function toCreateRequest(): CreateModelRequest {
    return {
      apiName: state.apiName,
      label: state.label,
      description: state.description || undefined,
    }
  }

  function toUpdateRequest(): UpdateModelRequest {
    return {
      label: state.label,
      description: state.description || undefined,
    }
  }

  function initFrom(model: TerritoryModel) {
    state.apiName = model.apiName
    state.label = model.label
    state.description = model.description
  }

  function reset() {
    Object.assign(state, defaultState())
    errors.apiName = undefined
    errors.label = undefined
  }

  return { state, errors, validate, toCreateRequest, toUpdateRequest, initFrom, reset }
}
