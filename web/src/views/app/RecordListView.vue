<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useRecordsStore } from '@/stores/records'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import FieldDisplay from '@/components/records/FieldDisplay.vue'
import { IconButton } from '@/components/ui/icon-button'
import { Plus, ChevronLeft, ChevronRight } from 'lucide-vue-next'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'
import type { PaginationMeta } from '@/types/metadata'
import { computed, type Ref } from 'vue'

const props = defineProps<{ objectName: string }>()

const router = useRouter()
const route = useRoute()
const store = useRecordsStore()
const toast = useToast()
const { currentDescribe, records, recordsPagination, loading } = storeToRefs(store)

const paginationRef = computed(() => {
  if (!recordsPagination.value) return null
  return {
    page: recordsPagination.value.page,
    perPage: recordsPagination.value.perPage,
    total: recordsPagination.value.total,
    totalPages: recordsPagination.value.totalPages,
  } as PaginationMeta
}) as Ref<PaginationMeta | null>

const { isFirstPage, isLastPage, pageInfo, nextPage, prevPage } = usePagination(
  paginationRef,
  (page) => loadRecords(page),
)

function loadRecords(page = 1) {
  store.fetchRecords(props.objectName, page).catch((err) => toast.errorFromApi(err))
}

function loadAll() {
  store.fetchDescribe(props.objectName).catch((err) => toast.errorFromApi(err))
  loadRecords()
}

onMounted(loadAll)

watch(() => route.params.objectName, () => {
  loadAll()
})

function goToDetail(record: Record<string, unknown>) {
  const id = String(record['Id'] ?? record['id'] ?? '')
  router.push({ name: 'record-detail', params: { objectName: props.objectName, recordId: id } })
}

const breadcrumbs = computed(() => [
  { label: 'CRM', to: '/app' },
  { label: currentDescribe.value?.pluralLabel ?? props.objectName },
])
</script>

<template>
  <div>
    <PageHeader :title="currentDescribe?.pluralLabel ?? objectName" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          v-if="currentDescribe?.isCreateable"
          :icon="Plus"
          tooltip="Create"
          variant="default"
          @click="router.push({ name: 'record-create', params: { objectName } })"
        />
      </template>
    </PageHeader>

    <div v-if="loading && records.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!loading && records.length === 0"
      title="No records"
      description="Create your first record"
    >
      <template v-if="currentDescribe?.isCreateable" #action>
        <IconButton
          :icon="Plus"
          tooltip="Create"
          variant="default"
          @click="router.push({ name: 'record-create', params: { objectName } })"
        />
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead v-for="field in store.tableFields" :key="field.apiName">
              {{ field.label }}
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="(record, idx) in records"
            :key="String(record['Id'] ?? idx)"
            class="cursor-pointer"
            @click="goToDetail(record)"
          >
            <TableCell v-for="field in store.tableFields" :key="field.apiName">
              <FieldDisplay :field="field" :value="record[field.apiName]" />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <div
        v-if="recordsPagination && recordsPagination.totalPages > 1"
        class="flex items-center justify-between mt-4"
      >
        <span class="text-sm text-muted-foreground">{{ pageInfo }}</span>
        <div class="flex gap-2">
          <IconButton
            :icon="ChevronLeft"
            tooltip="Back"
            variant="outline"
            :disabled="isFirstPage"
            @click="prevPage"
          />
          <IconButton
            :icon="ChevronRight"
            tooltip="Forward"
            variant="outline"
            :disabled="isLastPage"
            @click="nextPage"
          />
        </div>
      </div>
    </template>
  </div>
</template>
