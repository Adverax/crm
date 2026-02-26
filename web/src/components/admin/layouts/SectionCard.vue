<script setup lang="ts">
import { computed } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ChevronDown, Columns2 } from 'lucide-vue-next'
import FieldChip from './FieldChip.vue'
import type { SectionConfig, LayoutFieldConfig } from '@/types/layouts'

export interface SectionField {
  apiName: string
  label: string
  type: string
}

const props = defineProps<{
  sectionKey: string
  label: string
  config: SectionConfig
  fields: SectionField[]
  fieldConfigs: Record<string, LayoutFieldConfig>
  selected?: boolean
  selectedField?: string
}>()

defineEmits<{
  select: []
  selectField: [fieldName: string]
}>()

const columns = computed(() => props.config.columns ?? 1)
const collapsible = computed(() => props.config.collapsible ?? false)

const gridClass = computed(() => {
  const cols = columns.value
  if (cols === 2) return 'grid grid-cols-2 gap-2'
  if (cols === 3) return 'grid grid-cols-3 gap-2'
  if (cols === 4) return 'grid grid-cols-4 gap-2'
  return 'grid grid-cols-1 gap-2'
})
</script>

<template>
  <Card
    class="cursor-pointer transition-colors"
    :class="[
      selected && !selectedField
        ? 'ring-2 ring-primary border-primary'
        : 'hover:border-primary/30',
    ]"
    data-testid="section-card"
    @click.self="$emit('select')"
  >
    <CardHeader class="py-3 px-4" data-testid="section-header" @click="$emit('select')">
      <div class="flex items-center justify-between">
        <CardTitle class="text-sm font-medium">{{ label || sectionKey }}</CardTitle>
        <div class="flex items-center gap-1.5">
          <Badge variant="secondary" class="text-xs gap-1">
            <Columns2 class="h-3 w-3" />
            {{ columns }}
          </Badge>
          <ChevronDown
            v-if="collapsible"
            class="h-3.5 w-3.5 text-muted-foreground"
          />
        </div>
      </div>
    </CardHeader>
    <CardContent class="px-4 pb-3 pt-0">
      <div v-if="fields.length" :class="gridClass">
        <FieldChip
          v-for="field in fields"
          :key="field.apiName"
          :field-name="field.apiName"
          :label="field.label"
          :field-type="field.type"
          :config="fieldConfigs[field.apiName]"
          :selected="selectedField === field.apiName"
          @select="$emit('selectField', field.apiName)"
        />
      </div>
      <p v-else class="text-xs text-muted-foreground italic">
        No fields in this section
      </p>
    </CardContent>
  </Card>
</template>
