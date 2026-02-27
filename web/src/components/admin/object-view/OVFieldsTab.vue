<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Plus, Trash2 } from 'lucide-vue-next'
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
import type { OVViewField } from '@/types/object-views'

const props = defineProps<{
  fields: OVViewField[]
}>()

const emit = defineEmits<{
  'update:fields': [value: OVViewField[]]
}>()

function addField() {
  const updated: OVViewField[] = [...props.fields, { name: '' }]
  emit('update:fields', updated)
}

function removeField(index: number) {
  const updated = [...props.fields]
  updated.splice(index, 1)
  emit('update:fields', updated)
}

function updateName(index: number, value: string | number) {
  const updated = [...props.fields]
  updated[index] = { ...updated[index], name: String(value) }
  emit('update:fields', updated)
}

function updateType(index: number, value: string) {
  const updated = [...props.fields]
  const f = props.fields[index]
  if (!f) return
  const newType = value === 'none' ? undefined : value
  updated[index] = { name: f.name, type: newType as OVViewField['type'], expr: f.expr, when: f.when }
  emit('update:fields', updated)
}

function updateExpr(index: number, value: string) {
  const updated = [...props.fields]
  const f = props.fields[index]
  if (!f) return
  updated[index] = { name: f.name, type: f.type, expr: value || undefined, when: f.when }
  emit('update:fields', updated)
}

function updateWhen(index: number, value: string) {
  const updated = [...props.fields]
  const f = props.fields[index]
  if (!f) return
  updated[index] = { name: f.name, type: f.type, expr: f.expr, when: value || undefined }
  emit('update:fields', updated)
}

function isComputed(field: OVViewField): boolean {
  return !!field.expr
}
</script>

<template>
  <Card>
    <CardContent class="pt-6 space-y-4">
      <div class="flex items-center justify-between">
        <p class="text-sm text-muted-foreground">
          Order matters — first 3 are used as highlights. Fields with expr are computed.
        </p>
        <IconButton
          :icon="Plus"
          tooltip="Add field"
          variant="outline"
          data-testid="add-field-btn"
          @click="addField"
        />
      </div>

      <div
        v-for="(field, idx) in fields"
        :key="idx"
        class="border rounded-lg p-3 space-y-2"
        data-testid="field-entry"
      >
        <div class="grid grid-cols-4 gap-2 items-end">
          <div class="space-y-1">
            <Label class="text-xs">Name</Label>
            <Input
              :model-value="field.name"
              placeholder="field_api_name"
              class="font-mono"
              @update:model-value="(v: string | number) => updateName(idx, v)"
            />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Type</Label>
            <Select
              :model-value="field.type ?? 'none'"
              @update:model-value="(v) => updateType(idx, String(v))"
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">—</SelectItem>
                <SelectItem value="string">string</SelectItem>
                <SelectItem value="int">int</SelectItem>
                <SelectItem value="float">float</SelectItem>
                <SelectItem value="bool">bool</SelectItem>
                <SelectItem value="timestamp">timestamp</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="space-y-1">
            <Label class="text-xs">When (CEL)</Label>
            <Input
              :model-value="field.when ?? ''"
              placeholder=""
              class="font-mono"
              @update:model-value="(v: string | number) => updateWhen(idx, String(v))"
            />
          </div>
          <div class="flex justify-end">
            <IconButton
              :icon="Trash2"
              tooltip="Remove field"
              variant="ghost"
              class="text-destructive hover:text-destructive"
              @click="removeField(idx)"
            />
          </div>
        </div>
        <div v-if="isComputed(field) || field.type" class="space-y-1">
          <Label class="text-xs">Expression (CEL)</Label>
          <Textarea
            :model-value="field.expr ?? ''"
            placeholder="record.amount * 1.2"
            class="font-mono text-sm"
            rows="2"
            @update:model-value="(v: string | number) => updateExpr(idx, String(v))"
          />
        </div>
      </div>

      <div v-if="fields.length === 0" class="text-sm text-muted-foreground">
        No fields configured. The system will auto-generate the form from all accessible fields.
      </div>
    </CardContent>
  </Card>
</template>
