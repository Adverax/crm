<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useTerritoryAdminStore } from '@/stores/territoryAdmin'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Plus, ChevronLeft, ChevronRight, MoreVertical } from 'lucide-vue-next'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import type { Territory } from '@/types/territory'
import { storeToRefs } from 'pinia'

const router = useRouter()
const route = useRoute()
const store = useTerritoryAdminStore()
const toast = useToast()
const { territories, territoriesPagination, territoriesLoading, models } = storeToRefs(store)

const selectedModelId = ref<string>('')
const deleteTarget = ref<Territory | null>(null)
const showDeleteDialog = ref(false)

const { isFirstPage, isLastPage, pageInfo, nextPage, prevPage } = usePagination(
  territoriesPagination,
  (page) => loadTerritories(page),
)

function loadTerritories(page = 1) {
  if (!selectedModelId.value) return
  store.fetchTerritories({ modelId: selectedModelId.value, page, perPage: 20 })
    .catch((err) => toast.errorFromApi(err))
}

onMounted(async () => {
  await store.fetchModels({ perPage: 1000 }).catch((err) => toast.errorFromApi(err))
  const queryModelId = route.query.modelId as string | undefined
  if (queryModelId) {
    selectedModelId.value = queryModelId
  } else if (models.value.length > 0 && models.value[0]) {
    selectedModelId.value = models.value[0].id
  }
})

watch(selectedModelId, () => {
  if (selectedModelId.value) {
    loadTerritories()
  }
})

function goToDetail(territory: Territory) {
  router.push({ name: 'admin-territory-detail', params: { territoryId: territory.id } })
}

function confirmDelete(territory: Territory) {
  deleteTarget.value = territory
  showDeleteDialog.value = true
}

async function onDeleteConfirmed() {
  if (!deleteTarget.value) return
  try {
    await store.deleteTerritory(deleteTarget.value.id)
    toast.success('Territory deleted')
    loadTerritories()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
    deleteTarget.value = null
  }
}

function getParentLabel(parentId: string | null): string {
  if (!parentId) return '—'
  const parent = territories.value.find((t) => t.id === parentId)
  return parent?.label ?? parentId
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US')
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onModelChange(value: any) {
  selectedModelId.value = String(value)
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Territories' },
]
</script>

<template>
  <div>
    <PageHeader title="Territories" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create territory"
          variant="default"
          :disabled="!selectedModelId"
          @click="router.push({ name: 'admin-territory-create', query: { modelId: selectedModelId } })"
        />
      </template>
    </PageHeader>

    <div class="mb-4 max-w-xs">
      <Select :model-value="selectedModelId" @update:model-value="onModelChange">
        <SelectTrigger>
          <SelectValue placeholder="Select model" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="model in models" :key="model.id" :value="model.id">
            {{ model.label }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div v-if="territoriesLoading && territories.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!territoriesLoading && territories.length === 0 && selectedModelId"
      title="No territories"
      description="Create the first territory in the selected model"
    >
      <template #action>
        <IconButton
          :icon="Plus"
          tooltip="Create territory"
          variant="default"
          @click="router.push({ name: 'admin-territory-create', query: { modelId: selectedModelId } })"
        />
      </template>
    </EmptyState>

    <EmptyState
      v-else-if="!selectedModelId"
      title="Select a model"
      description="Select a model from the list above to view territories"
    />

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>API Name</TableHead>
            <TableHead>Label</TableHead>
            <TableHead>Parent</TableHead>
            <TableHead>Created</TableHead>
            <TableHead class="w-16" />
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="territory in territories"
            :key="territory.id"
            class="cursor-pointer"
            @click="goToDetail(territory)"
          >
            <TableCell class="font-medium">
              <RouterLink
                :to="{ name: 'admin-territory-detail', params: { territoryId: territory.id } }"
                class="text-primary hover:underline"
                @click.stop
              >
                {{ territory.apiName }}
              </RouterLink>
            </TableCell>
            <TableCell>{{ territory.label }}</TableCell>
            <TableCell class="text-muted-foreground">{{ getParentLabel(territory.parentId) }}</TableCell>
            <TableCell class="text-muted-foreground">{{ formatDate(territory.createdAt) }}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="sm" class="h-8 w-8 p-0" @click.stop>
                    <span class="sr-only">Actions</span>
                    <MoreVertical />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(territory)">
                    Open
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(territory)"
                  >
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <div v-if="territoriesPagination && territoriesPagination.totalPages > 1" class="flex items-center justify-between mt-4">
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

    <ConfirmDialog
      :open="showDeleteDialog"
      title="Delete territory?"
      :description="`Territory '${deleteTarget?.label}' (${deleteTarget?.apiName}) will be permanently deleted.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
