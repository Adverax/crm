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
    toast.success('Группа удалена')
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
    personal: 'Персональная',
    role: 'Роль',
    role_and_subordinates: 'Роль и подчинённые',
    public: 'Публичная',
  }
  return labels[type] ?? type
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ru-RU')
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Группы' },
]
</script>

<template>
  <div>
    <PageHeader title="Группы" :breadcrumbs="breadcrumbs">
      <template #actions>
        <Button @click="router.push({ name: 'admin-group-create' })">
          Создать группу
        </Button>
      </template>
    </PageHeader>

    <div v-if="groupsLoading && groups.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!groupsLoading && groups.length === 0"
      title="Нет групп"
      description="Создайте первую группу"
    >
      <template #action>
        <Button @click="router.push({ name: 'admin-group-create' })">
          Создать группу
        </Button>
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Название</TableHead>
            <TableHead>API Name</TableHead>
            <TableHead>Тип</TableHead>
            <TableHead>Создан</TableHead>
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
                    <span class="sr-only">Действия</span>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01" /></svg>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(group)">
                    Открыть
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(group)"
                  >
                    Удалить
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
      title="Удалить группу?"
      :description="`Группа «${deleteTarget?.label}» (${deleteTarget?.apiName}) будет удалена без возможности восстановления.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
