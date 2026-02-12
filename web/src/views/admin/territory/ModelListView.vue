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
    toast.success('Модель удалена')
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
  return new Date(iso).toLocaleDateString('ru-RU')
}

const breadcrumbs = [
  { label: 'Админ', to: '/admin' },
  { label: 'Модели территорий' },
]
</script>

<template>
  <div>
    <PageHeader title="Модели территорий" :breadcrumbs="breadcrumbs">
      <template #actions>
        <Button @click="router.push({ name: 'admin-territory-model-create' })">
          Создать модель
        </Button>
      </template>
    </PageHeader>

    <div v-if="modelsLoading && models.length === 0" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <EmptyState
      v-else-if="!modelsLoading && models.length === 0"
      title="Нет моделей территорий"
      description="Создайте первую модель территорий для организации данных"
    >
      <template #action>
        <Button @click="router.push({ name: 'admin-territory-model-create' })">
          Создать модель
        </Button>
      </template>
    </EmptyState>

    <template v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>API Name</TableHead>
            <TableHead>Название</TableHead>
            <TableHead>Статус</TableHead>
            <TableHead>Создана</TableHead>
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
                    <span class="sr-only">Действия</span>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01" /></svg>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click.stop="goToDetail(model)">
                    Открыть
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive"
                    @click.stop="confirmDelete(model)"
                  >
                    Удалить
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
      title="Удалить модель?"
      :description="`Модель «${deleteTarget?.label}» (${deleteTarget?.apiName}) будет удалена без возможности восстановления.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
