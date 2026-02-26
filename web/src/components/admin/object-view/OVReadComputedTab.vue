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
import type { OVReadComputed } from '@/types/object-views'

const props = defineProps<{
  computed: OVReadComputed[]
}>()

const emit = defineEmits<{
  'update:computed': [value: OVReadComputed[]]
}>()

function addComputed() {
  const updated: OVReadComputed[] = [...props.computed, { name: '', type: 'string' as const, expr: '', when: '' }]
  emit('update:computed', updated)
}

function removeComputed(index: number) {
  const updated = [...props.computed]
  updated.splice(index, 1)
  emit('update:computed', updated)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onTypeChange(index: number, value: any) {
  const field = props.computed[index]
  if (field) field.type = String(value) as OVReadComputed['type']
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <Label class="text-base">Computed</Label>
      <IconButton
        :icon="Plus"
        tooltip="Add computed field"
        variant="outline"
        data-testid="add-read-computed-btn"
        @click="addComputed"
      />
    </div>
    <p class="text-sm text-muted-foreground">
      Computed fields derived from CEL expressions, scoped to this view.
    </p>

    <Card
      v-for="(vf, idx) in computed"
      :key="idx"
      data-testid="read-computed-card"
    >
      <CardContent class="pt-6 space-y-3">
        <div class="grid grid-cols-3 gap-3">
          <div class="space-y-1">
            <Label class="text-xs">Name</Label>
            <Input v-model="vf.name" placeholder="total_with_tax" class="font-mono" />
          </div>
          <div class="space-y-1">
            <Label class="text-xs">Type</Label>
            <Select
              :model-value="vf.type"
              @update:model-value="(v) => onTypeChange(idx, v)"
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="string">string</SelectItem>
                <SelectItem value="int">int</SelectItem>
                <SelectItem value="float">float</SelectItem>
                <SelectItem value="bool">bool</SelectItem>
                <SelectItem value="timestamp">timestamp</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="flex items-end gap-2">
            <div class="flex-1 space-y-1">
              <Label class="text-xs">When (CEL)</Label>
              <Input v-model="vf.when" placeholder="" class="font-mono" />
            </div>
            <IconButton
              :icon="Trash2"
              tooltip="Delete computed field"
              variant="ghost"
              class="text-destructive hover:text-destructive"
              @click="removeComputed(idx)"
            />
          </div>
        </div>
        <div class="space-y-1">
          <Label class="text-xs">Expression (CEL)</Label>
          <Textarea
            v-model="vf.expr"
            placeholder="record.amount * 1.2"
            class="font-mono text-sm"
            rows="2"
          />
        </div>
      </CardContent>
    </Card>

    <div v-if="computed.length === 0" class="text-sm text-muted-foreground">
      No computed fields configured.
    </div>
  </div>
</template>
