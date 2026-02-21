<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { OVDefault } from '@/types/object-views'

const props = defineProps<{
  defaults: OVDefault[]
}>()

const emit = defineEmits<{
  'update:defaults': [value: OVDefault[]]
}>()

function addDefault() {
  const updated: OVDefault[] = [...props.defaults, { field: '', expr: '', on: 'create' as const, when: '' }]
  emit('update:defaults', updated)
}

function removeDefault(index: number) {
  const updated = [...props.defaults]
  updated.splice(index, 1)
  emit('update:defaults', updated)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onOnChange(index: number, value: any) {
  const d = props.defaults[index]
  if (d) d.on = String(value) as OVDefault['on']
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <Label class="text-base">Defaults</Label>
      <IconButton
        :icon="Plus"
        tooltip="Add default"
        variant="outline"
        data-testid="add-default-btn"
        @click="addDefault"
      />
    </div>
    <p class="text-sm text-muted-foreground">
      View-scoped default expressions (replace metadata-level defaults).
    </p>

    <Card
      v-for="(def, idx) in defaults"
      :key="idx"
      data-testid="default-card"
    >
      <CardContent class="pt-6 space-y-3">
        <div class="grid grid-cols-3 gap-3">
          <div class="space-y-1">
            <Label class="text-xs">Field</Label>
            <Input v-model="def.field" placeholder="status" class="font-mono" />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">On</Label>
            <Select
              :model-value="def.on"
              @update:model-value="(v) => onOnChange(idx, v)"
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="create">create</SelectItem>
                <SelectItem value="update">update</SelectItem>
                <SelectItem value="create,update">create,update</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="flex items-end gap-2">
            <div class="flex-1 space-y-1">
              <Label class="text-xs">When (CEL)</Label>
              <Input v-model="def.when" placeholder="" class="font-mono" />
            </div>
            <IconButton
              :icon="Trash2"
              tooltip="Delete default"
              variant="ghost"
              class="text-destructive hover:text-destructive"
              @click="removeDefault(idx)"
            />
          </div>
        </div>
        <div class="space-y-1">
          <Label class="text-xs">Expression (CEL)</Label>
          <Textarea
            v-model="def.expr"
            placeholder="'draft'"
            class="font-mono text-sm"
            rows="2"
          />
        </div>
      </CardContent>
    </Card>

    <div v-if="defaults.length === 0" class="text-sm text-muted-foreground">
      No view-scoped defaults configured.
    </div>
  </div>
</template>
