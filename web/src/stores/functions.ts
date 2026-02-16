import { ref } from 'vue'
import { defineStore } from 'pinia'
import { functionsApi } from '@/api/functions'
import type { Function } from '@/types/functions'

export const useFunctionsStore = defineStore('functions', () => {
  const functions = ref<Function[]>([])
  const loaded = ref(false)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function ensureLoaded() {
    if (loaded.value || loading.value) return
    await load()
  }

  async function load() {
    loading.value = true
    error.value = null
    try {
      const response = await functionsApi.list()
      functions.value = response.data ?? []
      loaded.value = true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load functions'
    } finally {
      loading.value = false
    }
  }

  async function invalidate() {
    loaded.value = false
    await load()
  }

  return {
    functions,
    loaded,
    loading,
    error,
    ensureLoaded,
    invalidate,
  }
})
