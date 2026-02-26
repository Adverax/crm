<script setup lang="ts">
import { IconButton } from '@/components/ui/icon-button'
import { Plus, X } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'

const props = defineProps<{
  fields: string[]
}>()

const emit = defineEmits<{
  'update:fields': [value: string[]]
}>()

function addField() {
  const updated = [...props.fields, '']
  emit('update:fields', updated)
}

function removeField(index: number) {
  const updated = [...props.fields]
  updated.splice(index, 1)
  emit('update:fields', updated)
}
</script>

<template>
  <Card>
    <CardContent class="pt-6 space-y-4">
      <div class="flex items-center justify-between">
        <p class="text-sm text-muted-foreground">
          Order matters — first 3 are used as highlights in the computed form.
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
        class="flex items-center gap-2"
        data-testid="field-entry"
      >
        <Input
          :model-value="field"
          placeholder="field_api_name"
          class="font-mono"
          @update:model-value="(v: string | number) => { const updated = [...fields]; updated[idx] = String(v); emit('update:fields', updated) }"
        />
        <IconButton
          :icon="X"
          tooltip="Remove"
          variant="ghost"
          @click="removeField(idx)"
        />
      </div>

      <div v-if="fields.length === 0" class="text-sm text-muted-foreground">
        No fields configured. The system will auto-generate the form from all accessible fields.
      </div>
    </CardContent>
  </Card>
</template>
