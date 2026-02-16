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
import { IconButton } from '@/components/ui/icon-button'
import { Plus, MoreVertical } from 'lucide-vue-next'
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
      <h2 class="text-lg font-semibold">Fields</h2>
      <IconButton
        :icon="Plus"
        tooltip="Add field"
        variant="outline"
        @click="emit('create')"
      />
    </div>

    <EmptyState
      v-if="!loading && props.fields.length === 0"
      title="No fields"
      description="Add the first field for this object"
    >
      <template #action>
        <IconButton
          :icon="Plus"
          tooltip="Add field"
          variant="outline"
          @click="emit('create')"
        />
      </template>
    </EmptyState>

    <Table v-if="props.fields.length > 0">
      <TableHeader>
        <TableRow>
          <TableHead>API Name</TableHead>
          <TableHead>Label</TableHead>
          <TableHead>Type</TableHead>
          <TableHead class="text-center">Required</TableHead>
          <TableHead class="text-center">Unique</TableHead>
          <TableHead class="text-center">System</TableHead>
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
            <Badge v-if="field.isRequired" variant="destructive" class="text-xs">Yes</Badge>
          </TableCell>
          <TableCell class="text-center">
            <Badge v-if="field.isUnique" variant="outline" class="text-xs">Yes</Badge>
          </TableCell>
          <TableCell class="text-center">
            <Badge v-if="field.isSystemField" variant="secondary" class="text-xs">System</Badge>
          </TableCell>
          <TableCell>
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="sm" class="h-8 w-8 p-0">
                  <span class="sr-only">Actions</span>
                  <MoreVertical />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem @click="emit('edit', field)">
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                  v-if="!field.isSystemField && !field.isPlatformManaged"
                  class="text-destructive"
                  @click="confirmDelete(field)"
                >
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <ConfirmDialog
      :open="showDeleteDialog"
      title="Delete field?"
      :description="`Field '${deleteTarget?.label}' (${deleteTarget?.apiName}) will be permanently deleted.`"
      @update:open="showDeleteDialog = $event"
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>
