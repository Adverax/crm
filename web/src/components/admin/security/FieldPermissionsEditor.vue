<script setup lang="ts">
import { usePermissionEditorStore } from '@/stores/permissionEditor'
import { useToast } from '@/composables/useToast'
import BitmaskCheckboxGroup from '@/components/admin/security/BitmaskCheckboxGroup.vue'
import { FLS_FLAGS } from '@/types/security'
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
import { Skeleton } from '@/components/ui/skeleton'
import { Label } from '@/components/ui/label'
import { storeToRefs } from 'pinia'

const store = usePermissionEditorStore()
const toast = useToast()
const { objectDefinitions, selectedObjectId, fieldsForObject, isLoading } = storeToRefs(store)

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onObjectSelect(value: any) {
  store.selectObjectForFls(String(value)).catch((err) => toast.errorFromApi(err))
}

function getPermissionValue(fieldId: string): number {
  return store.getFieldPermission(fieldId)?.permissions ?? 0
}

async function onPermissionChange(fieldId: string, value: number) {
  try {
    await store.setFieldPermission(fieldId, value)
  } catch (err) {
    toast.errorFromApi(err)
  }
}
</script>

<template>
  <div class="space-y-4">
    <div class="max-w-sm space-y-2">
      <Label>Объект</Label>
      <Select :model-value="selectedObjectId ?? undefined" @update:model-value="onObjectSelect">
        <SelectTrigger>
          <SelectValue placeholder="Выберите объект" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="obj in objectDefinitions" :key="obj.id" :value="obj.id">
            {{ obj.label }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div v-if="isLoading && selectedObjectId" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <div v-else-if="!selectedObjectId" class="text-sm text-muted-foreground py-8 text-center">
      Выберите объект для настройки разрешений на поля
    </div>

    <Table v-else-if="fieldsForObject.length > 0">
      <TableHeader>
        <TableRow>
          <TableHead>Поле</TableHead>
          <TableHead>Разрешения</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="field in fieldsForObject" :key="field.id">
          <TableCell class="font-medium">
            {{ field.label }}
            <span class="text-xs text-muted-foreground ml-1">({{ field.apiName }})</span>
          </TableCell>
          <TableCell>
            <BitmaskCheckboxGroup
              :flags="FLS_FLAGS"
              :model-value="getPermissionValue(field.id)"
              @update:model-value="onPermissionChange(field.id, $event)"
            />
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <div v-else-if="selectedObjectId && fieldsForObject.length === 0" class="text-sm text-muted-foreground py-8 text-center">
      У выбранного объекта нет полей
    </div>
  </div>
</template>
