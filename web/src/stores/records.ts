import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { recordsApi, type DescribeOptions } from '@/api/records'
import type {
  ObjectNavItem,
  ObjectDescribe,
  FieldDescribe,
  RecordData,
  RecordPagination,
} from '@/types/records'
import type { FormDescribe, FormQuery } from '@/types/object-views'

export const useRecordsStore = defineStore('records', () => {
  const navObjects = ref<ObjectNavItem[]>([])
  const currentDescribe = ref<ObjectDescribe | null>(null)
  const records = ref<RecordData[]>([])
  const currentRecord = ref<RecordData | null>(null)
  const recordsPagination = ref<RecordPagination | null>(null)
  const queryResults = ref<Map<string, RecordData[]>>(new Map())
  const loading = ref(false)
  const error = ref<string | null>(null)

  const currentForm = computed<FormDescribe | undefined>(() => {
    return currentDescribe.value?.form
  })

  const fieldMap = computed<Map<string, FieldDescribe>>(() => {
    const map = new Map<string, FieldDescribe>()
    if (!currentDescribe.value) return map
    for (const f of currentDescribe.value.fields) {
      map.set(f.apiName, f)
    }
    return map
  })

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

  const formListFields = computed<FieldDescribe[]>(() => {
    const form = currentForm.value
    if (!form?.listFields?.length) return tableFields.value
    return resolveFields(form.listFields)
  })

  const formHighlightFields = computed<FieldDescribe[]>(() => {
    const form = currentForm.value
    if (!form?.highlightFields?.length) return []
    return resolveFields(form.highlightFields)
  })

  function resolveFields(apiNames: string[]): FieldDescribe[] {
    const result: FieldDescribe[] = []
    for (const name of apiNames) {
      const f = fieldMap.value.get(name)
      if (f) result.push(f)
    }
    return result
  }

  async function fetchNavObjects() {
    const res = await recordsApi.listObjects()
    navObjects.value = res.data
  }

  async function fetchDescribe(objectName: string, options?: DescribeOptions) {
    loading.value = true
    error.value = null
    try {
      const res = await recordsApi.describeObject(objectName, options)
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

  const formQueries = computed<FormQuery[]>(() => {
    return currentForm.value?.queries ?? []
  })

  async function fetchQuery(ovApiName: string, queryName: string, params?: Record<string, string>) {
    loading.value = true
    error.value = null
    try {
      const res = await recordsApi.executeQuery(ovApiName, queryName, params)
      queryResults.value.set(queryName, res.data.records ?? [])
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
    queryResults,
    loading,
    error,
    currentForm,
    formQueries,
    fieldMap,
    editableFields,
    tableFields,
    formListFields,
    formHighlightFields,
    resolveFields,
    fetchNavObjects,
    fetchDescribe,
    fetchRecords,
    fetchRecord,
    fetchQuery,
    createRecord,
    updateRecord,
    deleteRecord,
  }
})
