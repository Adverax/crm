import { reactive, computed } from 'vue'
import type {
  FieldDefinition,
  FieldType,
  FieldSubtype,
  FieldConfig,
  CreateFieldRequest,
  UpdateFieldRequest,
} from '@/types/metadata'
import { FIELD_TYPE_SUBTYPES, CONFIG_FIELDS_BY_TYPE, type ConfigFieldDef } from '@/types/field-types'

const API_NAME_REGEX = /^[A-Za-z][A-Za-z0-9_]*$/

export interface FieldFormState {
  apiName: string
  label: string
  fieldType: FieldType
  fieldSubtype: FieldSubtype | ''
  description: string
  helpText: string
  isRequired: boolean
  isUnique: boolean
  sortOrder: number
  config: FieldConfig
}

export interface FieldFormErrors {
  apiName?: string
  label?: string
  fieldType?: string
}

function defaultState(): FieldFormState {
  return {
    apiName: '',
    label: '',
    fieldType: 'text',
    fieldSubtype: 'plain',
    description: '',
    helpText: '',
    isRequired: false,
    isUnique: false,
    sortOrder: 0,
    config: {},
  }
}

export function useFieldForm(existing?: FieldDefinition) {
  const state = reactive<FieldFormState>(existing ? {
    apiName: existing.apiName,
    label: existing.label,
    fieldType: existing.fieldType,
    fieldSubtype: existing.fieldSubtype ?? '',
    description: existing.description,
    helpText: existing.helpText,
    isRequired: existing.isRequired,
    isUnique: existing.isUnique,
    sortOrder: existing.sortOrder,
    config: { ...existing.config },
  } : defaultState())

  const errors = reactive<FieldFormErrors>({})

  const availableSubtypes = computed(() => {
    return FIELD_TYPE_SUBTYPES[state.fieldType] ?? []
  })

  const configFields = computed<ConfigFieldDef[]>(() => {
    const key = state.fieldSubtype
      ? `${state.fieldType}/${state.fieldSubtype}`
      : state.fieldType
    return CONFIG_FIELDS_BY_TYPE[key] ?? []
  })

  function validate(): boolean {
    errors.apiName = undefined
    errors.label = undefined
    errors.fieldType = undefined

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
      errors.apiName = 'Только латинские буквы, цифры и подчёркивания'
      valid = false
    }

    if (!state.label.trim()) {
      errors.label = 'Название обязательно'
      valid = false
    }

    if (!state.fieldType) {
      errors.fieldType = 'Тип поля обязателен'
      valid = false
    }

    return valid
  }

  function onFieldTypeChange(newType: FieldType) {
    state.fieldType = newType
    const subtypes = FIELD_TYPE_SUBTYPES[newType]
    state.fieldSubtype = subtypes.length > 0 ? subtypes[0]! : ''
    state.config = {}
  }

  function toCreateRequest(): CreateFieldRequest {
    const req: CreateFieldRequest = {
      apiName: state.apiName,
      label: state.label,
      fieldType: state.fieldType,
      description: state.description,
      helpText: state.helpText,
      isRequired: state.isRequired,
      isUnique: state.isUnique,
      sortOrder: state.sortOrder,
      isCustom: true,
    }
    if (state.fieldSubtype) {
      req.fieldSubtype = state.fieldSubtype as FieldSubtype
    }
    if (Object.keys(state.config).length > 0) {
      req.config = { ...state.config }
    }
    return req
  }

  function toUpdateRequest(): UpdateFieldRequest {
    const req: UpdateFieldRequest = {
      label: state.label,
      description: state.description,
      helpText: state.helpText,
      isRequired: state.isRequired,
      isUnique: state.isUnique,
      sortOrder: state.sortOrder,
    }
    if (Object.keys(state.config).length > 0) {
      req.config = { ...state.config }
    }
    return req
  }

  function reset() {
    Object.assign(state, defaultState())
    errors.apiName = undefined
    errors.label = undefined
    errors.fieldType = undefined
  }

  function initFrom(field: FieldDefinition) {
    state.apiName = field.apiName
    state.label = field.label
    state.fieldType = field.fieldType
    state.fieldSubtype = field.fieldSubtype ?? ''
    state.description = field.description
    state.helpText = field.helpText
    state.isRequired = field.isRequired
    state.isUnique = field.isUnique
    state.sortOrder = field.sortOrder
    state.config = { ...field.config }
  }

  return {
    state,
    errors,
    availableSubtypes,
    configFields,
    validate,
    onFieldTypeChange,
    toCreateRequest,
    toUpdateRequest,
    reset,
    initFrom,
  }
}
