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
    toast.success('Роль удалена')
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
  return new Date(iso).toLocaleDateString('ru-RU')
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Роли' },
]
</script>

<template>
  <div>
    <PageHeader title="Роли" :breadcrumbs="breadcrumbs">
      <template #actions>
        <Button @click="router.push({ name: 'admin-role-create' })">
          Создать роль
        </Button>
      </template>
    </PageHeader>

    <div v-if="rolesLoading && roles.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!rolesLoading && roles.length === 0"
      title="Нет ролей"
      description="Создайте первую роль в иерархии"
    >
      <template #action>
        <Button @click="router.push({ name: 'admin-role-create' })">
          Создать роль
        </Button>
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>API Name</TableHead>
            <TableHead>Название</TableHead>
            <TableHead>Родительская роль</TableHead>
            <TableHead>Создан</TableHead>
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
                    <span class="sr-only">Действия</span>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01" /></svg>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(role)">
                    Открыть
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(role)"
                  >
                    Удалить
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
          <Button variant="outline" size="sm" :disabled="isFirstPage" @click="prevPage">
            Назад
          </Button>
          <Button variant="outline" size="sm" :disabled="isLastPage" @click="nextPage">
            Вперёд
          </Button>
        </div>
      </div>
    </template>

    <ConfirmDialog
      :open="showDeleteDialog"
      title="Удалить роль?"
      :description="`Роль «${deleteTarget?.label}» (${deleteTarget?.apiName}) будет удалена без возможности восстановления.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
