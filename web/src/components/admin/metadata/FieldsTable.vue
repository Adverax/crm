<script setup lang="ts">
import { ref } from 'vue'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Badge } from '@/components/ui/badge'
import FieldTypeBadge from './FieldTypeBadge.vue'
import ConfirmDialog from '@/components/admin/ConfirmDialog.vue'
import EmptyState from '@/components/admin/EmptyState.vue'
import type { FieldDefinition } from '@/types/metadata'

const props = defineProps<{
  fields: FieldDefinition[]
  loading?: boolean
}>()

const emit = defineEmits<{
  create: []
  edit: [field: FieldDefinition]
  delete: [field: FieldDefinition]
}>()

const deleteTarget = ref<FieldDefinition | null>(null)
const showDeleteDialog = ref(false)

function confirmDelete(field: FieldDefinition) {
  deleteTarget.value = field
  showDeleteDialog.value = true
}

function onDeleteConfirmed() {
  if (deleteTarget.value) {
    emit('delete', deleteTarget.value)
  }
  showDeleteDialog.value = false
  deleteTarget.value = null
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold">Поля</h2>
      <Button size="sm" @click="emit('create')">
        Добавить поле
      </Button>
    </div>

    <EmptyState
      v-if="!loading && props.fields.length === 0"
      title="Нет полей"
      description="Добавьте первое поле для этого объекта"
    >
      <template #action>
        <Button size="sm" @click="emit('create')">Добавить поле</Button>
      </template>
    </EmptyState>

    <Table v-if="props.fields.length > 0">
      <TableHeader>
        <TableRow>
          <TableHead>API Name</TableHead>
          <TableHead>Название</TableHead>
          <TableHead>Тип</TableHead>
          <TableHead class="text-center">Обязательное</TableHead>
          <TableHead class="text-center">Уникальное</TableHead>
          <TableHead class="text-center">Системное</TableHead>
          <TableHead class="w-16" />
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="field in props.fields" :key="field.id">
          <TableCell class="font-mono text-sm">{{ field.apiName }}</TableCell>
          <TableCell>{{ field.label }}</TableCell>
          <TableCell>
            <FieldTypeBadge :field-type="field.fieldType" :field-subtype="field.fieldSubtype" />
          </TableCell>
          <TableCell class="text-center">
            <Badge v-if="field.isRequired" variant="destructive" class="text-xs">Да</Badge>
          </TableCell>
          <TableCell class="text-center">
            <Badge v-if="field.isUnique" variant="outline" class="text-xs">Да</Badge>
          </TableCell>
          <TableCell class="text-center">
            <Badge v-if="field.isSystemField" variant="secondary" class="text-xs">Системное</Badge>
          </TableCell>
          <TableCell>
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="sm" class="h-8 w-8 p-0">
                  <span class="sr-only">Действия</span>
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01" /></svg>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem @click="emit('edit', field)">
                  Редактировать
                </DropdownMenuItem>
                <DropdownMenuItem
                  v-if="!field.isSystemField && !field.isPlatformManaged"
                  class="text-destructive"
                  @click="confirmDelete(field)"
                >
                  Удалить
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <ConfirmDialog
      :open="showDeleteDialog"
      title="Удалить поле?"
      :description="`Поле «${deleteTarget?.label}» (${deleteTarget?.apiName}) будет удалено без возможности восстановления.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
