<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
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
import type { Group } from '@/types/security'
import { storeToRefs } from 'pinia'

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { groups, groupsPagination, groupsLoading } = storeToRefs(store)

const deleteTarget = ref<Group | null>(null)
const showDeleteDialog = ref(false)

const { isFirstPage, isLastPage, pageInfo, nextPage, prevPage } = usePagination(
  groupsPagination,
  (page) => loadGroups(page),
)

function loadGroups(page = 1) {
  store.fetchGroups({ page, perPage: 20 }).catch((err) => toast.errorFromApi(err))
}

onMounted(() => loadGroups())

function goToDetail(group: Group) {
  router.push({ name: 'admin-group-detail', params: { groupId: group.id } })
}

function confirmDelete(group: Group) {
  deleteTarget.value = group
  showDeleteDialog.value = true
}

async function onDeleteConfirmed() {
  if (!deleteTarget.value) return
  try {
    await store.deleteGroup(deleteTarget.value.id)
    toast.success('Group deleted')
    loadGroups()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
    deleteTarget.value = null
  }
}

function groupTypeLabel(type: string): string {
  const labels: Record<string, string> = {
    personal: 'Personal',
    role: 'Role',
    role_and_subordinates: 'Role & Subordinates',
    public: 'Public',
    territory: 'Territory',
  }
  return labels[type] ?? type
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US')
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Groups' },
]
</script>

<template>
  <div>
    <PageHeader title="Groups" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create Group"
          variant="default"
          @click="router.push({ name: 'admin-group-create' })"
        />
      </template>
    </PageHeader>

    <div v-if="groupsLoading && groups.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!groupsLoading && groups.length === 0"
      title="No Groups"
      description="Create your first group"
    >
      <template #action>
        <IconButton
          :icon="Plus"
          tooltip="Create Group"
          variant="default"
          @click="router.push({ name: 'admin-group-create' })"
        />
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Label</TableHead>
            <TableHead>API Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Created</TableHead>
            <TableHead class="w-16" />
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="group in groups"
            :key="group.id"
            class="cursor-pointer"
            @click="goToDetail(group)"
          >
            <TableCell class="font-medium">
              <RouterLink
                :to="{ name: 'admin-group-detail', params: { groupId: group.id } }"
                class="text-primary hover:underline"
                @click.stop
              >
                {{ group.label }}
              </RouterLink>
            </TableCell>
            <TableCell class="text-muted-foreground">{{ group.apiName }}</TableCell>
            <TableCell>{{ groupTypeLabel(group.groupType) }}</TableCell>
            <TableCell class="text-muted-foreground">{{ formatDate(group.createdAt) }}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="sm" class="h-8 w-8 p-0" @click.stop>
                    <span class="sr-only">Actions</span>
                    <MoreVertical />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(group)">
                    Open
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(group)"
                  >
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <div v-if="groupsPagination && groupsPagination.totalPages > 1" class="flex items-center justify-between mt-4">
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
      title="Delete Group?"
      :description="`Group '${deleteTarget?.label}' (${deleteTarget?.apiName}) will be permanently deleted.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
