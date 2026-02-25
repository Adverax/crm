<script setup lang="ts">
import { computed } from 'vue'
import { IconButton } from '@/components/ui/icon-button'
import { Plus, Trash2 } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const props = defineProps<{
  modelValue: Record<string, string> | undefined
  label: string
  keyPlaceholder?: string
  valuePlaceholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, string> | undefined]
}>()

const entries = computed(() => {
  const map = props.modelValue ?? {}
  return Object.entries(map)
})

function updateKey(oldKey: string, newKey: string) {
  const map = { ...props.modelValue }
  const val = map[oldKey] ?? ''
  delete map[oldKey]
  map[newKey] = val
  emit('update:modelValue', Object.keys(map).length > 0 ? map : undefined)
}

function updateValue(key: string, val: string) {
  const map = { ...props.modelValue, [key]: val }
  emit('update:modelValue', map)
}

function addEntry() {
  const map = { ...props.modelValue, '': '' }
  emit('update:modelValue', map)
}

function removeEntry(key: string) {
  const map = { ...props.modelValue }
  delete map[key]
  emit('update:modelValue', Object.keys(map).length > 0 ? map : undefined)
}
</script>

<template>
  <div class="space-y-1" data-testid="kv-editor">
    <div class="flex items-center justify-between">
      <Label class="text-xs">{{ label }}</Label>
      <IconButton
        :icon="Plus"
        tooltip="Add entry"
        variant="ghost"
        size="icon-sm"
        data-testid="kv-add-btn"
        @click="addEntry"
      />
    </div>
    <div v-if="entries.length === 0" class="text-xs text-muted-foreground italic">
      No entries. Click "+" to add.
    </div>
    <div v-for="([key, val], i) in entries" :key="i" class="flex items-center gap-1">
      <Input
        :model-value="key"
        :placeholder="keyPlaceholder ?? 'key'"
        class="h-7 text-xs font-mono flex-1"
        @input="updateKey(key, ($event.target as HTMLInputElement).value)"
      />
      <span class="text-xs text-muted-foreground">=</span>
      <Input
        :model-value="val"
        :placeholder="valuePlaceholder ?? 'expression'"
        class="h-7 text-xs font-mono flex-1"
        @input="updateValue(key, ($event.target as HTMLInputElement).value)"
      />
      <IconButton
        :icon="Trash2"
        tooltip="Remove"
        variant="ghost"
        size="icon-sm"
        class="text-destructive h-7 w-7"
        @click="removeEntry(key)"
      />
    </div>
  </div>
</template>
