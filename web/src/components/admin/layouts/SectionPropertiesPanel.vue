<script setup lang="ts">
import { computed } from 'vue'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { SectionConfig } from '@/types/layouts'

const props = defineProps<{
  sectionKey: string
  sectionLabel: string
  config: SectionConfig
}>()

const emit = defineEmits<{
  'update:config': [config: SectionConfig]
}>()

function update(patch: Partial<SectionConfig>) {
  emit('update:config', { ...props.config, ...patch })
}

const columns = computed({
  get: () => props.config.columns ?? 1,
  set: (v: number) => update({ columns: Math.max(1, Math.min(4, v)) }),
})

const collapsible = computed({
  get: () => props.config.collapsible ?? false,
  set: (v: boolean) => {
    const patch: Partial<SectionConfig> = { collapsible: v }
    if (!v) patch.collapsed = false
    update(patch)
  },
})

const collapsed = computed({
  get: () => props.config.collapsed ?? false,
  set: (v: boolean) => update({ collapsed: v }),
})

const visibilityExpr = computed({
  get: () => props.config.visibilityExpr ?? '',
  set: (v: string) => update({ visibilityExpr: v || undefined }),
})
</script>

<template>
  <Card>
    <CardHeader class="py-3 px-4">
      <CardTitle class="text-sm">Section: {{ sectionLabel || sectionKey }}</CardTitle>
    </CardHeader>
    <CardContent class="px-4 pb-4 space-y-4">
      <div>
        <Label for="section-columns">Columns</Label>
        <Input
          id="section-columns"
          type="number"
          :min="1"
          :max="4"
          :model-value="columns"
          data-testid="section-columns"
          @update:model-value="columns = Number($event)"
        />
      </div>

      <div class="flex items-center justify-between">
        <Label for="section-collapsible">Collapsible</Label>
        <Switch
          id="section-collapsible"
          :checked="collapsible"
          data-testid="section-collapsible"
          @update:checked="collapsible = $event"
        />
      </div>

      <div v-if="collapsible" class="flex items-center justify-between">
        <Label for="section-collapsed">Collapsed by default</Label>
        <Switch
          id="section-collapsed"
          :checked="collapsed"
          data-testid="section-collapsed"
          @update:checked="collapsed = $event"
        />
      </div>

      <div>
        <Label for="section-visibility">Visibility expression</Label>
        <Textarea
          id="section-visibility"
          :model-value="visibilityExpr"
          placeholder="e.g. record.Status == 'Active'"
          rows="2"
          class="font-mono text-sm"
          data-testid="section-visibility"
          @update:model-value="visibilityExpr = String($event)"
        />
      </div>
    </CardContent>
  </Card>
</template>
