import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { securityApi } from '@/api/security'
import { metadataApi } from '@/api/metadata'
import type { ObjectPermission, FieldPermission } from '@/types/security'
import type { ObjectDefinition, FieldDefinition } from '@/types/metadata'

export const usePermissionEditorStore = defineStore('permissionEditor', () => {
  const permissionSetId = ref<string | null>(null)

  // OLS
  const objectPermissions = ref<ObjectPermission[]>([])
  const objectDefinitions = ref<ObjectDefinition[]>([])

  // FLS
  const fieldPermissions = ref<FieldPermission[]>([])
  const selectedObjectId = ref<string | null>(null)
  const fieldsForObject = ref<FieldDefinition[]>([])

  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const getObjectPermission = computed(() => {
    return (objectId: string): ObjectPermission | undefined =>
      objectPermissions.value.find((op) => op.objectId === objectId)
  })

  const getFieldPermission = computed(() => {
    return (fieldId: string): FieldPermission | undefined =>
      fieldPermissions.value.find((fp) => fp.fieldId === fieldId)
  })

  async function loadForPermissionSet(psId: string) {
    permissionSetId.value = psId
    isLoading.value = true
    error.value = null
    try {
      const [olsResponse, objectsResponse] = await Promise.all([
        securityApi.listObjectPermissions(psId),
        metadataApi.listObjects({ perPage: 1000 }),
      ])
      objectPermissions.value = olsResponse.data
      objectDefinitions.value = objectsResponse.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки разрешений'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function selectObjectForFls(objectId: string) {
    if (!permissionSetId.value) return
    selectedObjectId.value = objectId
    isLoading.value = true
    error.value = null
    try {
      const [flsResponse, fieldsResponse] = await Promise.all([
        securityApi.listFieldPermissions(permissionSetId.value, objectId),
        metadataApi.listFields(objectId),
      ])
      fieldPermissions.value = flsResponse.data
      fieldsForObject.value = fieldsResponse.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки разрешений на поля'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function setObjectPermission(objectId: string, permissions: number) {
    if (!permissionSetId.value) return
    error.value = null
    try {
      const response = await securityApi.setObjectPermission(permissionSetId.value, {
        objectId,
        permissions,
      })
      const idx = objectPermissions.value.findIndex((op) => op.objectId === objectId)
      if (idx >= 0) {
        objectPermissions.value[idx] = response.data
      } else {
        objectPermissions.value.push(response.data)
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка сохранения разрешения'
      throw err
    }
  }

  async function setFieldPermission(fieldId: string, permissions: number) {
    if (!permissionSetId.value) return
    error.value = null
    try {
      const response = await securityApi.setFieldPermission(permissionSetId.value, {
        fieldId,
        permissions,
      })
      const idx = fieldPermissions.value.findIndex((fp) => fp.fieldId === fieldId)
      if (idx >= 0) {
        fieldPermissions.value[idx] = response.data
      } else {
        fieldPermissions.value.push(response.data)
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка сохранения разрешения на поле'
      throw err
    }
  }

  function reset() {
    permissionSetId.value = null
    objectPermissions.value = []
    objectDefinitions.value = []
    fieldPermissions.value = []
    selectedObjectId.value = null
    fieldsForObject.value = []
    error.value = null
  }

  return {
    permissionSetId,
    objectPermissions,
    objectDefinitions,
    fieldPermissions,
    selectedObjectId,
    fieldsForObject,
    isLoading,
    error,
    getObjectPermission,
    getFieldPermission,
    loadForPermissionSet,
    selectObjectForFls,
    setObjectPermission,
    setFieldPermission,
    reset,
  }
})
