<script setup lang="ts">
import { usePermissionEditorStore } from '@/stores/permissionEditor'
import { useToast } from '@/composables/useToast'
import BitmaskCheckboxGroup from '@/components/admin/security/BitmaskCheckboxGroup.vue'
import ErrorAlert from '@/components/admin/ErrorAlert.vue'
import { OLS_FLAGS } from '@/types/security'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'
import { storeToRefs } from 'pinia'

const store = usePermissionEditorStore()
const toast = useToast()
const { objectDefinitions, olsLoading, olsError } = storeToRefs(store)

function getPermissionValue(objectId: string): number {
  return store.getObjectPermission(objectId)?.permissions ?? 0
}

async function onPermissionChange(objectId: string, value: number) {
  try {
    await store.setObjectPermission(objectId, value)
  } catch (err) {
    toast.errorFromApi(err)
  }
}
</script>

<template>
  <div>
    <ErrorAlert v-if="olsError" :message="olsError" class="mb-4" />

    <div v-if="olsLoading" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <Table v-else-if="objectDefinitions.length > 0">
      <TableHeader>
        <TableRow>
          <TableHead>Объект</TableHead>
          <TableHead>Разрешения</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="obj in objectDefinitions" :key="obj.id">
          <TableCell class="font-medium">
            {{ obj.label }}
            <span class="text-xs text-muted-foreground ml-1">({{ obj.apiName }})</span>
          </TableCell>
          <TableCell>
            <BitmaskCheckboxGroup
              :flags="OLS_FLAGS"
              :model-value="getPermissionValue(obj.id)"
              @update:model-value="onPermissionChange(obj.id, $event)"
            />
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>
</template>
