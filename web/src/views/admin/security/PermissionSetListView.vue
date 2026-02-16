<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import PsTypeBadge from '@/components/admin/security/PsTypeBadge.vue'
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
import type { PermissionSet, PsType } from '@/types/security'
import { storeToRefs } from 'pinia'

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { permissionSets, permissionSetsPagination, permissionSetsLoading } = storeToRefs(store)

const filterType = ref<string>('all')
const deleteTarget = ref<PermissionSet | null>(null)
const showDeleteDialog = ref(false)

const { isFirstPage, isLastPage, pageInfo, nextPage, prevPage } = usePagination(
  permissionSetsPagination,
  (page) => loadPermissionSets(page),
)

function loadPermissionSets(page = 1) {
  const filter: { page: number; perPage: number; psType?: PsType } = { page, perPage: 20 }
  if (filterType.value !== 'all') {
    filter.psType = filterType.value as PsType
  }
  store.fetchPermissionSets(filter).catch((err) => toast.errorFromApi(err))
}

watch(filterType, () => loadPermissionSets(1))
onMounted(() => loadPermissionSets())

function goToDetail(ps: PermissionSet) {
  router.push({ name: 'admin-permission-set-detail', params: { permissionSetId: ps.id } })
}

function confirmDelete(ps: PermissionSet) {
  deleteTarget.value = ps
  showDeleteDialog.value = true
}

async function onDeleteConfirmed() {
  if (!deleteTarget.value) return
  try {
    await store.deletePermissionSet(deleteTarget.value.id)
    toast.success('Permission set deleted')
    loadPermissionSets()
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
  { label: 'Permission Sets' },
]
</script>

<template>
  <div>
    <PageHeader title="Permission Sets" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create Permission Set"
          variant="default"
          @click="router.push({ name: 'admin-permission-set-create' })"
        />
      </template>
    </PageHeader>

    <div class="mb-4">
      <Select v-model="filterType">
        <SelectTrigger class="w-48">
          <SelectValue placeholder="All Types" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Types</SelectItem>
          <SelectItem value="grant">Grant</SelectItem>
          <SelectItem value="deny">Deny</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div v-if="permissionSetsLoading && permissionSets.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!permissionSetsLoading && permissionSets.length === 0"
      title="No Permission Sets"
      description="Create your first permission set"
    >
      <template #action>
        <IconButton
          :icon="Plus"
          tooltip="Create Permission Set"
          variant="default"
          @click="router.push({ name: 'admin-permission-set-create' })"
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
            v-for="ps in permissionSets"
            :key="ps.id"
            class="cursor-pointer"
            @click="goToDetail(ps)"
          >
            <TableCell class="font-medium">
              <RouterLink
                :to="{ name: 'admin-permission-set-detail', params: { permissionSetId: ps.id } }"
                class="text-primary hover:underline"
                @click.stop
              >
                {{ ps.apiName }}
              </RouterLink>
            </TableCell>
            <TableCell>{{ ps.label }}</TableCell>
            <TableCell>
              <PsTypeBadge :type="ps.psType" />
            </TableCell>
            <TableCell class="text-muted-foreground">{{ formatDate(ps.createdAt) }}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="sm" class="h-8 w-8 p-0" @click.stop>
                    <span class="sr-only">Actions</span>
                    <MoreVertical />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(ps)">
                    Open
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(ps)"
                  >
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <div v-if="permissionSetsPagination && permissionSetsPagination.totalPages > 1" class="flex items-center justify-between mt-4">
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
      title="Delete Permission Set?"
      :description="`Permission set '${deleteTarget?.label}' (${deleteTarget?.apiName}) will be permanently deleted.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
