import type { PaginationMeta } from '@/types/metadata'
import { computed, type Ref } from 'vue'

export function usePagination(pagination: Ref<PaginationMeta | null>, load: (page: number) => void) {
  const isFirstPage = computed(() => !pagination.value || pagination.value.page <= 1)
  const isLastPage = computed(() => !pagination.value || pagination.value.page >= pagination.value.totalPages)
  const pageInfo = computed(() => {
    if (!pagination.value) return ''
    return `Page ${pagination.value.page} of ${pagination.value.totalPages}`
  })

  function goToPage(page: number) {
    load(page)
  }

  function nextPage() {
    if (!isLastPage.value && pagination.value) {
      goToPage(pagination.value.page + 1)
    }
  }

  function prevPage() {
    if (!isFirstPage.value && pagination.value) {
      goToPage(pagination.value.page - 1)
    }
  }

  return { isFirstPage, isLastPage, pageInfo, goToPage, nextPage, prevPage }
}
