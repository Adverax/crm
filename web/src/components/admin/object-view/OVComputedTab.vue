<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Trash2, Plus } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import type { OVComputed } from '@/types/object-views'

const props = defineProps<{
  computed: OVComputed[]
}>()

const emit = defineEmits<{
  'update:computed': [value: OVComputed[]]
}>()

function addComputed() {
  const updated = [...props.computed, { field: '', expr: '' }]
  emit('update:computed', updated)
}

function removeComputed(index: number) {
  const updated = [...props.computed]
  updated.splice(index, 1)
  emit('update:computed', updated)
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
        data-testid="add-computed-btn"
        @click="addComputed"
      />
    </div>
    <p class="text-sm text-muted-foreground">
      Fields whose values are computed from CEL expressions on save.
    </p>

    <Card
      v-for="(comp, idx) in computed"
      :key="idx"
      data-testid="computed-card"
    >
      <CardContent class="pt-6 space-y-3">
        <div class="flex items-center gap-3">
          <div class="flex-1 space-y-1">
            <Label class="text-xs">Field</Label>
            <Input v-model="comp.field" placeholder="total_with_tax" class="font-mono" />
          </div>
          <div class="flex items-end">
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
            v-model="comp.expr"
            placeholder="record.amount * (1 + record.tax_rate / 100)"
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
