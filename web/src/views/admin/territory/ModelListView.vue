<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import type { TerritoryModel } from '@/types/territory'
import { MODEL_STATUS_LABELS } from '@/types/territory'
import { storeToRefs } from 'pinia'

const router = useRouter()
const store = useTerritoryAdminStore()
const toast = useToast()
const { models, modelsPagination, modelsLoading } = storeToRefs(store)

const deleteTarget = ref<TerritoryModel | null>(null)
const showDeleteDialog = ref(false)

const { isFirstPage, isLastPage, pageInfo, nextPage, prevPage } = usePagination(
  modelsPagination,
  (page) => loadModels(page),
)

function loadModels(page = 1) {
  store.fetchModels({ page, perPage: 20 }).catch((err) => toast.errorFromApi(err))
}

onMounted(() => loadModels())

function goToDetail(model: TerritoryModel) {
  router.push({ name: 'admin-territory-model-detail', params: { modelId: model.id } })
}

function confirmDelete(model: TerritoryModel) {
  deleteTarget.value = model
  showDeleteDialog.value = true
}

async function onDeleteConfirmed() {
  if (!deleteTarget.value) return
  try {
    await store.deleteModel(deleteTarget.value.id)
    toast.success('Model deleted')
    loadModels()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
    deleteTarget.value = null
  }
}

function statusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'active': return 'default'
    case 'archived': return 'secondary'
    default: return 'outline'
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US')
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Territory Models' },
]
</script>

<template>
  <div>
    <PageHeader title="Territory Models" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create model"
          variant="default"
          @click="router.push({ name: 'admin-territory-model-create' })"
        />
      </template>
    </PageHeader>

    <div v-if="modelsLoading && models.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!modelsLoading && models.length === 0"
      title="No territory models"
      description="Create the first territory model to organize your data"
    >
      <template #action>
        <IconButton
          :icon="Plus"
          tooltip="Create model"
          variant="default"
          @click="router.push({ name: 'admin-territory-model-create' })"
        />
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>API Name</TableHead>
            <TableHead>Label</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created</TableHead>
            <TableHead class="w-16" />
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="model in models"
            :key="model.id"
            class="cursor-pointer"
            @click="goToDetail(model)"
          >
            <TableCell class="font-medium">
              <RouterLink
                :to="{ name: 'admin-territory-model-detail', params: { modelId: model.id } }"
                class="text-primary hover:underline"
                @click.stop
              >
                {{ model.apiName }}
              </RouterLink>
            </TableCell>
            <TableCell>{{ model.label }}</TableCell>
            <TableCell>
              <Badge :variant="statusVariant(model.status)">
                {{ MODEL_STATUS_LABELS[model.status] }}
              </Badge>
            </TableCell>
            <TableCell class="text-muted-foreground">{{ formatDate(model.createdAt) }}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="sm" class="h-8 w-8 p-0" @click.stop>
                    <span class="sr-only">Actions</span>
                    <MoreVertical />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(model)">
                    Open
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(model)"
                  >
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <div v-if="modelsPagination && modelsPagination.totalPages > 1" class="flex items-center justify-between mt-4">
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
      title="Delete model?"
      :description="`Model '${deleteTarget?.label}' (${deleteTarget?.apiName}) will be permanently deleted.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
