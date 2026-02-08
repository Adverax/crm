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
const { objects, pagination, isLoading } = storeToRefs(store)

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
    toast.success('Объект удалён')
    loadObjects()
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
  { label: 'Объекты' },
]
</script>

<template>
  <div>
    <PageHeader title="Объекты" :breadcrumbs="breadcrumbs">
      <template #actions>
        <Button @click="router.push({ name: 'admin-object-create' })">
          Создать объект
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
          <SelectItem value="standard">Standard</SelectItem>
          <SelectItem value="custom">Custom</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div v-if="isLoading && objects.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!isLoading && objects.length === 0"
      title="Нет объектов"
      description="Создайте первый объект метаданных"
    >
      <template #action>
        <Button @click="router.push({ name: 'admin-object-create' })">
          Создать объект
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
                    <span class="sr-only">Действия</span>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01" /></svg>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(obj)">
                    Открыть
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    v-if="obj.isDeleteableObject && !obj.isPlatformManaged"
                    class="text-destructive"
                    @click.stop="confirmDelete(obj)"
                  >
                    Удалить
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
      title="Удалить объект?"
      :description="`Объект «${deleteTarget?.label}» (${deleteTarget?.apiName}) и все его поля будут удалены без возможности восстановления.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
