<script setup lang="ts">
import type { FieldDescribe } from '@/types/records'

const props = defineProps<{
  field: FieldDescribe
  value: unknown
}>()

function formatValue(): string {
  const val = props.value
  if (val == null || val === '') return '—'

  if (props.field.fieldType === 'boolean') {
    return val ? 'Yes' : 'No'
  }

  if (props.field.fieldType === 'datetime') {
    try {
      const date = new Date(String(val))
      if (props.field.fieldSubtype === 'date') {
        return date.toLocaleDateString('en-US')
      }
      return date.toLocaleString('en-US')
    } catch {
      return String(val)
    }
  }

  if (props.field.fieldType === 'picklist' && props.field.config?.values) {
    const opt = props.field.config.values.find((v) => v.value === val)
    if (opt?.label) return opt.label
  }

  return String(val)
}
</script>

<template>
  <span>{{ formatValue() }}</span>
</template>
