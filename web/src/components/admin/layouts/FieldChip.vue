<script setup lang="ts">
import { computed } from 'vue'
import { Badge } from '@/components/ui/badge'
import {
  Type, Hash, Calendar, ToggleLeft, Link, List,
} from 'lucide-vue-next'
import type { LayoutFieldConfig } from '@/types/layouts'

const props = defineProps<{
  fieldName: string
  label: string
  fieldType: string
  config?: LayoutFieldConfig
  selected?: boolean
}>()

defineEmits<{
  select: []
}>()

const colSpan = computed(() => props.config?.colSpan ?? 1)

const typeIcon = computed(() => {
  switch (props.fieldType) {
    case 'text': return Type
    case 'number': case 'currency': case 'percent': return Hash
    case 'date': case 'datetime': return Calendar
    case 'boolean': return ToggleLeft
    case 'reference': return Link
    case 'picklist': return List
    default: return Type
  }
})
</script>

<template>
  <div
    class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border cursor-pointer text-sm transition-colors"
    :class="[
      selected
        ? 'border-primary bg-primary/10 ring-1 ring-primary'
        : 'border-border bg-background hover:border-primary/50 hover:bg-muted/50',
    ]"
    :style="colSpan > 1 ? `grid-column: span ${colSpan}` : ''"
    data-testid="field-chip"
    @click.stop="$emit('select')"
  >
    <component :is="typeIcon" class="h-3.5 w-3.5 text-muted-foreground shrink-0" />
    <span class="truncate">{{ label }}</span>
    <Badge v-if="colSpan > 1" variant="secondary" class="ml-auto text-xs px-1 py-0">
      {{ colSpan }}col
    </Badge>
    <Badge v-if="config?.layoutRef" variant="outline" class="text-xs px-1 py-0">
      ref
    </Badge>
  </div>
</template>
