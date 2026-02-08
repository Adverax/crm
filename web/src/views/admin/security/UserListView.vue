<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSecurityAdminStore } from '@/stores/securityAdmin'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/admin/PageHeader.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import ActiveStatusBadge from '@/components/admin/security/ActiveStatusBadge.vue'
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
import type { User } from '@/types/security'
import { storeToRefs } from 'pinia'

const router = useRouter()
const store = useSecurityAdminStore()
const toast = useToast()
const { users, usersPagination, profiles, roles, isLoading } = storeToRefs(store)

const deleteTarget = ref<User | null>(null)
const showDeleteDialog = ref(false)

const { isFirstPage, isLastPage, pageInfo, nextPage, prevPage } = usePagination(
  usersPagination,
  (page) => loadUsers(page),
)

function loadUsers(page = 1) {
  store.fetchUsers({ page, perPage: 20 }).catch((err) => toast.errorFromApi(err))
}

onMounted(async () => {
  loadUsers()
  try {
    await Promise.all([
      store.fetchProfiles({ perPage: 1000 }),
      store.fetchRoles({ perPage: 1000 }),
    ])
  } catch (err) {
    toast.errorFromApi(err)
  }
})

function goToDetail(user: User) {
  router.push({ name: 'admin-user-detail', params: { userId: user.id } })
}

function confirmDelete(user: User) {
  deleteTarget.value = user
  showDeleteDialog.value = true
}

async function onDeleteConfirmed() {
  if (!deleteTarget.value) return
  try {
    await store.deleteUser(deleteTarget.value.id)
    toast.success('Пользователь удалён')
    loadUsers()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
    deleteTarget.value = null
  }
}

function getProfileLabel(profileId: string): string {
  const profile = profiles.value.find((p) => p.id === profileId)
  return profile?.label ?? '—'
}

function getRoleLabel(roleId: string | null): string {
  if (!roleId) return '—'
  const role = roles.value.find((r) => r.id === roleId)
  return role?.label ?? '—'
}

function getUserDisplayName(user: User): string {
  const parts = [user.firstName, user.lastName].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : '—'
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Пользователи' },
]
</script>

<template>
  <div>
    <PageHeader title="Пользователи" :breadcrumbs="breadcrumbs">
      <template #actions>
        <Button @click="router.push({ name: 'admin-user-create' })">
          Создать пользователя
        </Button>
      </template>
    </PageHeader>

    <div v-if="isLoading && users.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!isLoading && users.length === 0"
      title="Нет пользователей"
      description="Создайте первого пользователя"
    >
      <template #action>
        <Button @click="router.push({ name: 'admin-user-create' })">
          Создать пользователя
        </Button>
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Username</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Имя</TableHead>
            <TableHead>Профиль</TableHead>
            <TableHead>Роль</TableHead>
            <TableHead>Статус</TableHead>
            <TableHead class="w-16" />
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="user in users"
            :key="user.id"
            class="cursor-pointer"
            @click="goToDetail(user)"
          >
            <TableCell class="font-medium">
              <RouterLink
                :to="{ name: 'admin-user-detail', params: { userId: user.id } }"
                class="text-primary hover:underline"
                @click.stop
              >
                {{ user.username }}
              </RouterLink>
            </TableCell>
            <TableCell>{{ user.email }}</TableCell>
            <TableCell>{{ getUserDisplayName(user) }}</TableCell>
            <TableCell class="text-muted-foreground">{{ getProfileLabel(user.profileId) }}</TableCell>
            <TableCell class="text-muted-foreground">{{ getRoleLabel(user.roleId) }}</TableCell>
            <TableCell>
              <ActiveStatusBadge :active="user.isActive" />
            </TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="sm" class="h-8 w-8 p-0" @click.stop>
                    <span class="sr-only">Действия</span>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01" /></svg>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(user)">
                    Открыть
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(user)"
                  >
                    Удалить
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <div v-if="usersPagination && usersPagination.totalPages > 1" class="flex items-center justify-between mt-4">
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
      title="Удалить пользователя?"
      :description="`Пользователь «${deleteTarget?.username}» будет удалён без возможности восстановления.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
