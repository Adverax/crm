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
const { permissionSets, permissionSetsPagination, isLoading } = storeToRefs(store)

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
    toast.success('Набор разрешений удалён')
    loadPermissionSets()
  } catch (err) {
    toast.errorFromApi(err)
  } finally {
    showDeleteDialog.value = false
    deleteTarget.value = null
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ru-RU')
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Наборы разрешений' },
]
</script>

<template>
  <div>
    <PageHeader title="Наборы разрешений" :breadcrumbs="breadcrumbs">
      <template #actions>
        <Button @click="router.push({ name: 'admin-permission-set-create' })">
          Создать набор
        </Button>
      </template>
    </PageHeader>

    <div class="mb-4">
      <Select v-model="filterType">
        <SelectTrigger class="w-48">
          <SelectValue placeholder="Все типы" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">Все типы</SelectItem>
          <SelectItem value="grant">Grant</SelectItem>
          <SelectItem value="deny">Deny</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div v-if="isLoading && permissionSets.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!isLoading && permissionSets.length === 0"
      title="Нет наборов разрешений"
      description="Создайте первый набор разрешений"
    >
      <template #action>
        <Button @click="router.push({ name: 'admin-permission-set-create' })">
          Создать набор
        </Button>
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>API Name</TableHead>
            <TableHead>Название</TableHead>
            <TableHead>Тип</TableHead>
            <TableHead>Создан</TableHead>
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
                    <span class="sr-only">Действия</span>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01" /></svg>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(ps)">
                    Открыть
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(ps)"
                  >
                    Удалить
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
      title="Удалить набор разрешений?"
      :description="`Набор «${deleteTarget?.label}» (${deleteTarget?.apiName}) будет удалён без возможности восстановления.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
