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

  const olsLoading = ref(false)
  const olsError = ref<string | null>(null)
  const flsLoading = ref(false)
  const flsError = ref<string | null>(null)

  const getObjectPermission = computed(() => {
    return (objectId: string): ObjectPermission | undefined =>
      objectPermissions.value.find((op) => op.objectId === objectId)
  })

  const getFieldPermission = computed(() => {
    return (fieldId: string): FieldPermission | undefined =>
      fieldPermissions.value.find((fp) => fp.fieldId === fieldId)
  })

  let olsRequestId = 0
  let flsRequestId = 0

  async function loadForPermissionSet(psId: string) {
    const requestId = ++olsRequestId
    permissionSetId.value = psId
    olsLoading.value = true
    olsError.value = null
    try {
      const [olsResponse, objectsResponse] = await Promise.all([
        securityApi.listObjectPermissions(psId),
        metadataApi.listObjects({ perPage: 1000 }),
      ])
      if (requestId !== olsRequestId) return
      objectPermissions.value = olsResponse.data ?? []
      objectDefinitions.value = objectsResponse.data ?? []
    } catch (err) {
      if (requestId !== olsRequestId) return
      olsError.value = err instanceof Error ? err.message : 'Failed to load permissions'
      throw err
    } finally {
      if (requestId === olsRequestId) {
        olsLoading.value = false
      }
    }
  }

  async function selectObjectForFls(objectId: string) {
    if (!permissionSetId.value) return
    const requestId = ++flsRequestId
    selectedObjectId.value = objectId
    flsLoading.value = true
    flsError.value = null
    try {
      const [flsResponse, fieldsResponse] = await Promise.all([
        securityApi.listFieldPermissions(permissionSetId.value, objectId),
        metadataApi.listFields(objectId),
      ])
      if (requestId !== flsRequestId) return
      fieldPermissions.value = flsResponse.data ?? []
      fieldsForObject.value = fieldsResponse.data ?? []
    } catch (err) {
      if (requestId !== flsRequestId) return
      flsError.value = err instanceof Error ? err.message : 'Failed to load field permissions'
      throw err
    } finally {
      if (requestId === flsRequestId) {
        flsLoading.value = false
      }
    }
  }

  async function setObjectPermission(objectId: string, permissions: number) {
    if (!permissionSetId.value) return
    olsError.value = null
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
      olsError.value = err instanceof Error ? err.message : 'Failed to save permission'
      throw err
    }
  }

  async function setFieldPermission(fieldId: string, permissions: number) {
    if (!permissionSetId.value) return
    flsError.value = null
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
      flsError.value = err instanceof Error ? err.message : 'Failed to save field permission'
      throw err
    }
  }

  function reset() {
    olsRequestId++
    flsRequestId++
    permissionSetId.value = null
    objectPermissions.value = []
    objectDefinitions.value = []
    fieldPermissions.value = []
    selectedObjectId.value = null
    fieldsForObject.value = []
    olsLoading.value = false
    olsError.value = null
    flsLoading.value = false
    flsError.value = null
  }

  return {
    permissionSetId,
    objectPermissions,
    objectDefinitions,
    fieldPermissions,
    selectedObjectId,
    fieldsForObject,
    olsLoading,
    olsError,
    flsLoading,
    flsError,
    getObjectPermission,
    getFieldPermission,
    loadForPermissionSet,
    selectObjectForFls,
    setObjectPermission,
    setFieldPermission,
    reset,
  }
})
