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
import type { UserRole } from '@/types/security'
import { storeToRefs } from 'pinia'

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { roles, rolesPagination, rolesLoading } = storeToRefs(store)

const deleteTarget = ref<UserRole | null>(null)
const showDeleteDialog = ref(false)

const { isFirstPage, isLastPage, pageInfo, nextPage, prevPage } = usePagination(
  rolesPagination,
  (page) => loadRoles(page),
)

function loadRoles(page = 1) {
  store.fetchRoles({ page, perPage: 20 }).catch((err) => toast.errorFromApi(err))
}

onMounted(() => loadRoles())

function goToDetail(role: UserRole) {
  router.push({ name: 'admin-role-detail', params: { roleId: role.id } })
}

function confirmDelete(role: UserRole) {
  deleteTarget.value = role
  showDeleteDialog.value = true
}

async function onDeleteConfirmed() {
  if (!deleteTarget.value) return
  try {
    await store.deleteRole(deleteTarget.value.id)
    toast.success('Role deleted')
    loadRoles()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
    deleteTarget.value = null
  }
}

function getParentLabel(parentId: string | null): string {
  if (!parentId) return '—'
  const parent = roles.value.find((r) => r.id === parentId)
  return parent?.label ?? parentId
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US')
}

const breadcrumbs = [
  { label: 'Admin', to: '/admin' },
  { label: 'Roles' },
]
</script>

<template>
  <div>
    <PageHeader title="Roles" :breadcrumbs="breadcrumbs">
      <template #actions>
        <IconButton
          :icon="Plus"
          tooltip="Create Role"
          variant="default"
          @click="router.push({ name: 'admin-role-create' })"
        />
      </template>
    </PageHeader>

    <div v-if="rolesLoading && roles.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!rolesLoading && roles.length === 0"
      title="No Roles"
      description="Create your first role in the hierarchy"
    >
      <template #action>
        <IconButton
          :icon="Plus"
          tooltip="Create Role"
          variant="default"
          @click="router.push({ name: 'admin-role-create' })"
        />
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>API Name</TableHead>
            <TableHead>Label</TableHead>
            <TableHead>Parent Role</TableHead>
            <TableHead>Created</TableHead>
            <TableHead class="w-16" />
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="role in roles"
            :key="role.id"
            class="cursor-pointer"
            @click="goToDetail(role)"
          >
            <TableCell class="font-medium">
              <RouterLink
                :to="{ name: 'admin-role-detail', params: { roleId: role.id } }"
                class="text-primary hover:underline"
                @click.stop
              >
                {{ role.apiName }}
              </RouterLink>
            </TableCell>
            <TableCell>{{ role.label }}</TableCell>
            <TableCell class="text-muted-foreground">{{ getParentLabel(role.parentId) }}</TableCell>
            <TableCell class="text-muted-foreground">{{ formatDate(role.createdAt) }}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="sm" class="h-8 w-8 p-0" @click.stop>
                    <span class="sr-only">Actions</span>
                    <MoreVertical />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(role)">
                    Open
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(role)"
                  >
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <div v-if="rolesPagination && rolesPagination.totalPages > 1" class="flex items-center justify-between mt-4">
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
      title="Delete Role?"
      :description="`Role '${deleteTarget?.label}' (${deleteTarget?.apiName}) will be permanently deleted.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
