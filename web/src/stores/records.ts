import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { recordsApi } from '@/api/records'
import type {
  ObjectNavItem,
  ObjectDescribe,
  FieldDescribe,
  RecordData,
  RecordPagination,
} from '@/types/records'

export const useRecordsStore = defineStore('records', () => {
  const navObjects = ref<ObjectNavItem[]>([])
  const currentDescribe = ref<ObjectDescribe | null>(null)
  const records = ref<RecordData[]>([])
  const currentRecord = ref<RecordData | null>(null)
  const recordsPagination = ref<RecordPagination | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const editableFields = computed<FieldDescribe[]>(() => {
    if (!currentDescribe.value) return []
    return currentDescribe.value.fields.filter(
      (f) => !f.isSystemField && !f.isReadOnly,
    )
  })

  const tableFields = computed<FieldDescribe[]>(() => {
    if (!currentDescribe.value) return []
    return currentDescribe.value.fields
      .filter((f) => !f.isSystemField || f.apiName === 'Id')
      .sort((a, b) => a.sortOrder - b.sortOrder)
      .slice(0, 8)
  })

  async function fetchNavObjects() {
    const res = await recordsApi.listObjects()
    navObjects.value = res.data
  }

  async function fetchDescribe(objectName: string) {
    loading.value = true
    error.value = null
    try {
      const res = await recordsApi.describeObject(objectName)
      currentDescribe.value = res.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'unknown error'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchRecords(objectName: string, page = 1, perPage = 20) {
    loading.value = true
    error.value = null
    try {
      const res = await recordsApi.listRecords(objectName, page, perPage)
      records.value = res.data ?? []
      recordsPagination.value = res.pagination
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'unknown error'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchRecord(objectName: string, recordId: string) {
    loading.value = true
    error.value = null
    try {
      const res = await recordsApi.getRecord(objectName, recordId)
      currentRecord.value = res.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'unknown error'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createRecord(objectName: string, data: RecordData): Promise<string> {
    loading.value = true
    error.value = null
    try {
      const res = await recordsApi.createRecord(objectName, data)
      return res.data.id
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'unknown error'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateRecord(objectName: string, recordId: string, data: RecordData) {
    loading.value = true
    error.value = null
    try {
      await recordsApi.updateRecord(objectName, recordId, data)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'unknown error'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteRecord(objectName: string, recordId: string) {
    loading.value = true
    error.value = null
    try {
      await recordsApi.deleteRecord(objectName, recordId)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'unknown error'
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    navObjects,
    currentDescribe,
    records,
    currentRecord,
    recordsPagination,
    loading,
    error,
    editableFields,
    tableFields,
    fetchNavObjects,
    fetchDescribe,
    fetchRecords,
    fetchRecord,
    createRecord,
    updateRecord,
    deleteRecord,
  }
})
