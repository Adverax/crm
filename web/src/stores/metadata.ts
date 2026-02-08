import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { metadataApi } from '@/api/metadata'
import type {
  ObjectDefinition,
  FieldDefinition,
  CreateObjectRequest,
  UpdateObjectRequest,
  CreateFieldRequest,
  UpdateFieldRequest,
  PaginationMeta,
  ObjectFilter,
} from '@/types/metadata'

export const useMetadataStore = defineStore('metadata', () => {
  const objects = ref<ObjectDefinition[]>([])
  const currentObject = ref<ObjectDefinition | null>(null)
  const fields = ref<FieldDefinition[]>([])
  const pagination = ref<PaginationMeta | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const hasObjects = computed(() => objects.value.length > 0)

  async function fetchObjects(filter?: ObjectFilter) {
    isLoading.value = true
    error.value = null
    try {
      const response = await metadataApi.listObjects(filter)
      objects.value = response.data
      pagination.value = response.pagination
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки объектов'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchObject(objectId: string) {
    isLoading.value = true
    error.value = null
    try {
      const response = await metadataApi.getObject(objectId)
      currentObject.value = response.data
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки объекта'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function createObject(data: CreateObjectRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await metadataApi.createObject(data)
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка создания объекта'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function updateObject(objectId: string, data: UpdateObjectRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await metadataApi.updateObject(objectId, data)
      currentObject.value = response.data
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка обновления объекта'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function deleteObject(objectId: string) {
    isLoading.value = true
    error.value = null
    try {
      await metadataApi.deleteObject(objectId)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка удаления объекта'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchFields(objectId: string) {
    isLoading.value = true
    error.value = null
    try {
      const response = await metadataApi.listFields(objectId)
      fields.value = response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки полей'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function createField(objectId: string, data: CreateFieldRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await metadataApi.createField(objectId, data)
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка создания поля'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function updateField(objectId: string, fieldId: string, data: UpdateFieldRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await metadataApi.updateField(objectId, fieldId, data)
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка обновления поля'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function deleteField(objectId: string, fieldId: string) {
    isLoading.value = true
    error.value = null
    try {
      await metadataApi.deleteField(objectId, fieldId)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка удаления поля'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  return {
    objects,
    currentObject,
    fields,
    pagination,
    isLoading,
    error,
    hasObjects,
    fetchObjects,
    fetchObject,
    createObject,
    updateObject,
    deleteObject,
    fetchFields,
    createField,
    updateField,
    deleteField,
  }
})
