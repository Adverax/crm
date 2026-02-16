<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMetadataStore } from '@/stores/metadata'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import ObjectTypeBadge from '@/components/admin/metadata/ObjectTypeBadge.vue'
import { Button } from '@/components/ui/button'
import { IconButton } from '@/components/ui/icon-button'
import { Plus, ChevronLeft, ChevronRight, MoreVertical } from 'lucide-vue-next'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
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
import type { ObjectDefinition, ObjectType } from '@/types/metadata'
import { storeToRefs } from 'pinia'

const router = useRouter()
const store = useMetadataStore()
const toast = useToast()
const { objects, pagination, objectsLoading } = storeToRefs(store)

const filterType = ref<string>('all')
const deleteTarget = ref<ObjectDefinition | null>(null)
const showDeleteDialog = ref(false)

const { isFirstPage, isLastPage, pageInfo, nextPage, prevPage } = usePagination(
  pagination,
  (page) => loadObjects(page),
)

function loadObjects(page = 1) {
  const filter: { page: number; perPage: number; objectType?: ObjectType } = { page, perPage: 20 }
  if (filterType.value !== 'all') {
    filter.objectType = filterType.value as ObjectType
  }
  store.fetchObjects(filter).catch((err) => toast.errorFromApi(err))
}

watch(filterType, () => loadObjects(1))
onMounted(() => loadObjects())

function goToDetail(obj: ObjectDefinition) {
  router.push({ name: 'admin-object-detail', params: { objectId: obj.id } })
}

function confirmDelete(obj: ObjectDefinition) {
  deleteTarget.value = obj
  showDeleteDialog.value = true
}

async function onDeleteConfirmed() {
  if (!deleteTarget.value) return
  try {
    await store.deleteObject(deleteTarget.value.id)
    toast.success('Object deleted')
    loadObjects()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
    deleteTarget.value = null
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US')
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Objects' },
]
</script>

<template>
  <div>
    <PageHeader title="Objects" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create object"
          variant="default"
          @click="router.push({ name: 'admin-object-create' })"
        />
      </template>
    </PageHeader>

    <div class="mb-4">
      <Select v-model="filterType">
        <SelectTrigger class="w-48">
          <SelectValue placeholder="All types" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All types</SelectItem>
          <SelectItem value="standard">Standard</SelectItem>
          <SelectItem value="custom">Custom</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div v-if="objectsLoading && objects.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!objectsLoading && objects.length === 0"
      title="No objects"
      description="Create your first metadata object"
    >
      <template #action>
        <IconButton
          :icon="Plus"
          tooltip="Create object"
          variant="default"
          @click="router.push({ name: 'admin-object-create' })"
        />
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>API Name</TableHead>
            <TableHead>Label</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Created</TableHead>
            <TableHead class="w-16" />
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="obj in objects"
            :key="obj.id"
            class="cursor-pointer"
            @click="goToDetail(obj)"
          >
            <TableCell class="font-medium">
              <RouterLink
                :to="{ name: 'admin-object-detail', params: { objectId: obj.id } }"
                class="text-primary hover:underline"
                @click.stop
              >
                {{ obj.apiName }}
              </RouterLink>
            </TableCell>
            <TableCell>{{ obj.label }}</TableCell>
            <TableCell>
              <ObjectTypeBadge :type="obj.objectType" />
            </TableCell>
            <TableCell class="text-muted-foreground">{{ formatDate(obj.createdAt) }}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="sm" class="h-8 w-8 p-0" @click.stop>
                    <span class="sr-only">Actions</span>
                    <MoreVertical />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(obj)">
                    Open
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    v-if="obj.isDeleteableObject && !obj.isPlatformManaged"
                    class="text-destructive"
                    @click.stop="confirmDelete(obj)"
                  >
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <div v-if="pagination && pagination.totalPages > 1" class="flex items-center justify-between mt-4">
        <span class="text-sm text-muted-foreground">{{ pageInfo }}</span>
        <div class="flex gap-2">
          <IconButton
            :icon="ChevronLeft"
            tooltip="Previous"
            variant="outline"
            :disabled="isFirstPage"
            @click="prevPage"
          />
          <IconButton
            :icon="ChevronRight"
            tooltip="Next"
            variant="outline"
            :disabled="isLastPage"
            @click="nextPage"
          />
        </div>
      </div>
    </template>

    <ConfirmDialog
      :open="showDeleteDialog"
      title="Delete object?"
      :description="`Object '${deleteTarget?.label}' (${deleteTarget?.apiName}) and all its fields will be permanently deleted.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
